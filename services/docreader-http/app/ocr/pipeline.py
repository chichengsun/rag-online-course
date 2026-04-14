"""
pipeline：在 Document 与原始文件字节上按需执行 OCR（扫描 PDF、独立图片、或显式 use_ocr）。
"""

from __future__ import annotations

import logging
from typing import Any

import fitz  # pymupdf

from app.config import Settings
from app.models.document import Document
from app.ocr.backends import OCRBackend, NoneOCRBackend, build_ocr_backend

logger = logging.getLogger(__name__)

# 少于此字数的 PDF 正文视为「疑似扫描件」，在开启自动 OCR 时触发整页识别。
_PDF_AUTO_OCR_CHAR_THRESHOLD = 40

_IMAGE_SUFFIXES = frozenset(
    {"png", "jpg", "jpeg", "gif", "webp", "bmp", "tif", "tiff"},
)


def _suffix_from_filename(filename: str) -> str:
    lower = filename.lower()
    if "." not in lower:
        return ""
    return lower.rsplit(".", 1)[-1]


def _should_ocr_pdf(
    doc: Document,
    suffix: str,
    *,
    use_ocr: bool,
    auto_scan: bool,
    backend: OCRBackend,
) -> bool:
    if suffix != "pdf":
        return False
    if isinstance(backend, NoneOCRBackend):
        return False
    if use_ocr:
        return True
    if not auto_scan:
        return False
    text_len = len((doc.content or "").strip())
    if text_len >= _PDF_AUTO_OCR_CHAR_THRESHOLD:
        return False
    hint = (doc.metadata or {}).get("hint")
    if hint == "no_text_maybe_scan":
        return True
    return text_len < _PDF_AUTO_OCR_CHAR_THRESHOLD


def _should_ocr_standalone_image(
    suffix: str,
    *,
    use_ocr: bool,
    auto_image: bool,
    backend: OCRBackend,
) -> bool:
    if suffix not in _IMAGE_SUFFIXES:
        return False
    if isinstance(backend, NoneOCRBackend):
        return False
    if use_ocr:
        return True
    return auto_image


def _ocr_pdf_pages(
    pdf_bytes: bytes,
    backend: OCRBackend,
    max_pages: int,
) -> tuple[str, dict[str, Any]]:
    """逐页渲染为 PNG 再 OCR，返回 (markdown片段, 统计元数据)。"""
    parts: list[str] = []
    meta: dict[str, Any] = {"ocr_pages_attempted": 0, "ocr_pages_with_text": 0}
    try:
        pdoc = fitz.open(stream=pdf_bytes, filetype="pdf")
    except Exception as e:
        logger.info("ocr pdf open failed: %s", e)
        return "", {**meta, "ocr_pdf_error": str(e)}
    n = min(pdoc.page_count, max_pages)
    meta["ocr_pages_attempted"] = n
    for i in range(n):
        page = pdoc.load_page(i)
        pix = page.get_pixmap(matrix=fitz.Matrix(2, 2))
        png_bytes = pix.tobytes("png")
        text = backend.recognize_image_bytes(png_bytes, mime_hint="image/png")
        if text.strip():
            meta["ocr_pages_with_text"] += 1
            parts.append(f"## Page {i + 1}（OCR）\n\n{text.strip()}")
    pdoc.close()
    md = "\n\n".join(parts)
    return md, meta


def enrich_document_with_ocr(
    doc: Document,
    raw_bytes: bytes,
    filename: str,
    settings: Settings,
    *,
    use_ocr: bool,
) -> Document:
    """
    enrich_document_with_ocr 在配置允许时补充 OCR 文本。
    - PDF：use_ocr 或 ocr_auto_scan_pdf 且正文极短；
    - 独立图片：use_ocr 或 ocr_auto_standalone_image。
    结果追加在正文后（保留原有 Markdown），metadata 合并 ocr_* 字段。
    """
    backend = build_ocr_backend(
        settings.ocr_backend,
        tesseract_lang=settings.ocr_tesseract_lang,
        http_url=settings.ocr_http_url,
        http_timeout_seconds=settings.ocr_http_timeout_seconds,
        http_api_key=settings.ocr_http_api_key,
    )
    suffix = _suffix_from_filename(filename)
    meta = dict(doc.metadata or {})
    additions: list[str] = []
    ocr_meta: dict[str, Any] = {}

    if _should_ocr_pdf(
        doc,
        suffix,
        use_ocr=use_ocr,
        auto_scan=settings.ocr_auto_scan_pdf,
        backend=backend,
    ):
        max_p = min(settings.ocr_max_pages, settings.parse_max_pages_for_images)
        ocr_md, pdf_ocr_meta = _ocr_pdf_pages(raw_bytes, backend, max_p)
        ocr_meta.update(pdf_ocr_meta)
        if ocr_md.strip():
            additions.append(ocr_md)

    if _should_ocr_standalone_image(
        suffix,
        use_ocr=use_ocr,
        auto_image=settings.ocr_auto_standalone_image,
        backend=backend,
    ):
        mime = "image/jpeg" if suffix in ("jpg", "jpeg") else "image/png"
        if suffix == "gif":
            mime = "image/gif"
        elif suffix == "webp":
            mime = "image/webp"
        text = backend.recognize_image_bytes(raw_bytes, mime_hint=mime)
        ocr_meta["ocr_standalone_image"] = True
        ocr_meta["ocr_image_chars"] = len(text.strip())
        if text.strip():
            additions.append(f"## OCR（图片）\n\n{text.strip()}")

    if not additions:
        meta["ocr_applied"] = False
        return Document(content=doc.content, images=dict(doc.images), metadata=meta)

    block = "\n\n---\n\n# OCR 补充\n\n" + "\n\n".join(additions)
    new_content = (doc.content or "").rstrip() + block if (doc.content or "").strip() else block.lstrip()
    meta["ocr_applied"] = True
    meta["ocr_backend"] = settings.ocr_backend
    meta.update(ocr_meta)
    return Document(content=new_content, images=dict(doc.images), metadata=meta)

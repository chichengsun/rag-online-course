"""
parsers：对外稳定入口 extract_content，内部委托 facade.parse_bytes_to_document（WeKnora 式解析链）。
"""

from __future__ import annotations

import logging
import os
from typing import Any, Optional

from app.config import Settings
from app.models.document import Document
from app.ocr.pipeline import enrich_document_with_ocr
from app.parser.builtin_pdf import append_pdf_page_images
from app.parser.facade import parse_bytes_to_document

logger = logging.getLogger(__name__)


def _suffix(filename: str) -> str:
    lower = filename.lower()
    if "." not in lower:
        return ""
    return lower.rsplit(".", 1)[-1]


def _document_to_api_images(doc: Document) -> list[dict[str, Any]]:
    """将 Document.images 转为 Go/docreaderhttp 约定的 images 列表。"""
    out: list[dict[str, Any]] = []
    for path, b64 in doc.images.items():
        fname = os.path.basename(path) or "image.bin"
        ext = os.path.splitext(fname)[1].lower()
        mime = "image/png"
        if ext in (".jpg", ".jpeg"):
            mime = "image/jpeg"
        elif ext == ".gif":
            mime = "image/gif"
        elif ext == ".webp":
            mime = "image/webp"
        elif ext == ".bmp":
            mime = "image/bmp"
        out.append({"filename": fname, "mime_type": mime, "data_base64": b64})
    return out


def extract_content(
    filename: str,
    data: bytes,
    *,
    return_images: bool,
    max_pdf_pages_as_images: int = 20,
    settings: Optional[Settings] = None,
    use_ocr: bool = False,
) -> tuple[str, list[dict[str, Any]], dict[str, Any], str]:
    """
    根据文件名后缀解析字节内容。
    返回 (markdown, images, metadata, error)；error 非空表示未支持或失败。
    """
    s = settings or Settings()
    suffix = _suffix(filename)
    meta: dict[str, Any] = {"engine": "weknora-like", "suffix": suffix or "none"}

    if suffix == "doc":
        return "", [], meta, "legacy .doc 不支持本地解析，请上传 docx"
    if suffix == "ppt":
        return "", [], meta, "legacy .ppt 不支持本地解析，请上传 pptx"

    doc = parse_bytes_to_document(filename, data, return_images=return_images)

    if suffix == "pdf" and return_images:
        doc = append_pdf_page_images(doc, data, max_pages=max_pdf_pages_as_images)

    # OCR 在「是否有效」判定之前执行，以便扫描版 PDF 仅靠 OCR 也能得到正文。
    doc = enrich_document_with_ocr(doc, data, filename, s, use_ocr=use_ocr)

    if not doc.is_useful():
        if suffix == "pdf":
            err = (
                "PDF 未抽取到文本与图片（可能为扫描件）；"
                "请在网关配置 OCR_BACKEND（如 tesseract）或请求 use_ocr=true，"
                "或 return_images=true 导出页图"
            )
        elif suffix == "xls" and (doc.metadata or {}).get("hint") == "xls_no_local_fallback":
            err = ".xls 未安装完整转换链时请改用 xlsx 或仅依赖 MarkItDown"
        else:
            err = (doc.metadata or {}).get("markitdown_error") or (
                f"解析结果为空（后缀 .{suffix}）" if suffix else "解析结果为空"
            )
        return "", [], {**meta, **(doc.metadata or {})}, err

    api_meta = {**meta, **(doc.metadata or {})}
    return doc.content, _document_to_api_images(doc), api_meta, ""

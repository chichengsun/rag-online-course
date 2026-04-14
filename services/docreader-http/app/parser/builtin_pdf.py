"""
builtin_pdf：PyMuPDF 文本抽取 + 可选整页光栅图，作为 MarkItDown 对难解析 PDF 的回退。
"""

from __future__ import annotations

import base64
import logging
from typing import Any

import fitz  # pymupdf

from app.models.document import Document
from app.parser.base_parser import BaseParser

logger = logging.getLogger(__name__)


class PdfBuiltinParser(BaseParser):
    """PdfBuiltinParser 使用 PyMuPDF 抽取分页文本；不替代 OCR。"""

    def parse_into_text(self, content: bytes) -> Document:
        images_out: dict[str, str] = {}
        parts: list[str] = []
        meta: dict[str, Any] = {"parser": "pymupdf"}
        try:
            doc = fitz.open(stream=content, filetype="pdf")
            meta["page_count"] = doc.page_count
            for i in range(doc.page_count):
                page = doc.load_page(i)
                text = page.get_text() or ""
                if text.strip():
                    parts.append(f"## Page {i + 1}\n\n{text.strip()}")
            doc.close()
        except Exception as e:
            logger.info("pymupdf parse failed: %s", e)
            return Document(metadata={**meta, "error": str(e)})

        md = "\n\n".join(parts) if parts else ""
        if not md.strip():
            return Document(
                metadata={
                    **meta,
                    "hint": "no_text_maybe_scan",
                },
            )
        return Document(content=md, images=images_out, metadata=meta)


def append_pdf_page_images(
    doc: Document,
    pdf_bytes: bytes,
    *,
    max_pages: int,
) -> Document:
    """
    append_pdf_page_images 在 return_images 为真时，把前 max_pages 页渲染为 PNG 并入 images。
    不改变已有正文，仅在 metadata 中记录 exported_pdf_pages。
    """
    images = dict(doc.images)
    n = 0
    try:
        pdoc = fitz.open(stream=pdf_bytes, filetype="pdf")
        n = min(pdoc.page_count, max_pages)
        for i in range(n):
            page = pdoc.load_page(i)
            pix = page.get_pixmap(matrix=fitz.Matrix(2, 2))
            png_bytes = pix.tobytes("png")
            key = f"images/pdf_page_{i + 1}.png"
            images[key] = base64.standard_b64encode(png_bytes).decode("ascii")
        pdoc.close()
    except Exception as e:
        logger.warning("pdf page raster failed: %s", e)
        meta = dict(doc.metadata)
        meta["pdf_page_render_error"] = str(e)
        return Document(content=doc.content, images=images, metadata=meta)
    meta = dict(doc.metadata)
    meta["exported_pdf_pages"] = n
    return Document(content=doc.content, images=images, metadata=meta)

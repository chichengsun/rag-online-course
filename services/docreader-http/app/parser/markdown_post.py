"""
markdown_post：从 Markdown 中拆出 data:image/*;base64,...，与 WeKnora MarkdownImageBase64 行为一致。
拆出的图片写入 Document.images，正文中的占位改为 ![](images/<uuid>.<ext>)。
"""

from __future__ import annotations

import base64
import logging
import re
import uuid

from app.models.document import Document

logger = logging.getLogger(__name__)

# alt 允许含 ]；MIME 子类型允许含连字符。
_B64_IMG_RE = re.compile(r"!\[(.*?)\]\(data:image/([^;]+);base64,([^\)]+)\)", re.DOTALL)


def split_data_uri_images(doc: Document) -> Document:
    """
    split_data_uri_images 扫描 content 中的内联 Base64 图片，解码后并入 images，
    并把 Markdown 替换为文件路径引用，避免巨型 data URI 留在正文里。
    """
    if not doc.content or "data:image/" not in doc.content:
        return doc

    images = dict(doc.images)

    def repl(m: re.Match[str]) -> str:
        alt = m.group(1)
        subtype = (m.group(2) or "png").strip().lower()
        b64_part = (m.group(3) or "").strip()
        raw = base64.standard_b64decode(b64_part)
        if not raw:
            logger.warning("skip empty base64 image in markdown")
            return alt
        ext = "png"
        if subtype in ("jpeg", "jpg", "jpe"):
            ext = "jpg"
        elif subtype in ("gif", "webp", "bmp", "tiff", "svg+xml"):
            ext = subtype.split("+")[0] if "+" in subtype else subtype
            if ext == "svg":
                ext = "svg"
        path = f"images/{uuid.uuid4().hex}.{ext}"
        images[path] = base64.standard_b64encode(raw).decode("ascii")
        return f"![{alt}]({path})"

    new_content = _B64_IMG_RE.sub(repl, doc.content)
    if new_content == doc.content and "data:image/" in doc.content:
        return doc
    return Document(content=new_content, images=images, metadata=dict(doc.metadata))

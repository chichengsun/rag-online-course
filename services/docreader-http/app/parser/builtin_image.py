"""
builtin_image：独立图片文件，行为对齐 WeKnora ImageParser（Markdown 引用 + images 字典）。
"""

from __future__ import annotations

import base64
import logging
import os

from PIL import Image

from app.models.document import Document
from app.parser.base_parser import BaseParser

logger = logging.getLogger(__name__)


class ImageBuiltinParser(BaseParser):
    """ImageBuiltinParser 生成占位说明，并在需要时把原图以 Base64 放入 images。"""

    def __init__(self, file_name: str = "", file_type: str | None = None, **kwargs: object) -> None:
        super().__init__(file_name=file_name, file_type=file_type, **kwargs)
        self._return_raw: bool = bool(kwargs.get("return_images", False))

    def parse_into_text(self, content: bytes) -> Document:
        import io

        meta = {"parser": "pillow", "format": "unknown"}
        try:
            im = Image.open(io.BytesIO(content))
            meta["width"], meta["height"] = im.size
            meta["format"] = im.format or "unknown"
        except Exception as e:
            return Document(metadata={**meta, "error": str(e)})

        fname = os.path.basename(self.file_name) or "image"
        ref = f"images/{fname}"
        images: dict[str, str] = {}
        if self._return_raw:
            buf = io.BytesIO()
            if (im.format or "").upper() == "JPEG":
                im = im.convert("RGB")
                im.save(buf, format="JPEG", quality=85)
            else:
                im.save(buf, format="PNG")
            images[ref] = base64.standard_b64encode(buf.getvalue()).decode("ascii")
            md = (
                f"![{fname}]({ref})\n\n"
                f"（尺寸 {meta['width']}x{meta['height']}，格式 {meta['format']}；未启用 OCR）"
            )
        else:
            md = (
                f"图片 `{fname}`，尺寸 {meta['width']}x{meta['height']}，"
                f"格式 {meta['format']}。（未启用 return_images / OCR）"
            )
        return Document(content=md, images=images, metadata=meta)

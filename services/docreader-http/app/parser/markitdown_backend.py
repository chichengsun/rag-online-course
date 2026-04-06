"""
markitdown_backend：优先使用微软 MarkItDown 将多类办公文档转为 Markdown（keep_data_uris=True）。
与 WeKnora StdMarkitdownParser 一致；失败返回空 Document 以便链式回退。
"""

from __future__ import annotations

import io
import logging

from app.models.document import Document
from app.parser.base_parser import BaseParser

logger = logging.getLogger(__name__)


class MarkitdownBackend(BaseParser):
    """MarkitdownBackend 封装 MarkItDown.convert，按 file_type 推断扩展名。"""

    def parse_into_text(self, content: bytes) -> Document:
        if not content:
            return Document(metadata={"markitdown": "empty_input"})
        try:
            from markitdown import MarkItDown
        except ImportError as e:
            logger.warning("markitdown not installed: %s", e)
            return Document(metadata={"markitdown": "not_installed"})

        ext = self.file_type or ""
        if ext and not ext.startswith("."):
            ext = "." + ext
        try:
            md = MarkItDown()
            result = md.convert(
                io.BytesIO(content),
                file_extension=ext,
                keep_data_uris=True,
            )
            text = (result.text_content or "").strip()
            return Document(
                content=result.text_content or "",
                metadata={
                    "parser": "markitdown",
                    "title": getattr(result, "title", None) or "",
                    "chars": len(text),
                },
            )
        except Exception as e:
            logger.info("markitdown failed type=%s err=%s", self.file_type, e)
            return Document(metadata={"markitdown": "error", "markitdown_error": str(e)})

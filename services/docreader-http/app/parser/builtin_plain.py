"""
builtin_plain：纯文本类字节按 UTF-8 / GB18030 解码。
"""

from __future__ import annotations

import logging
from typing import Any

from app.models.document import Document
from app.parser.base_parser import BaseParser

logger = logging.getLogger(__name__)


class PlainTextParser(BaseParser):
    """PlainTextParser 用于 txt/md/csv/json/xml 等当作文本读取。"""

    def parse_into_text(self, content: bytes) -> Document:
        meta: dict[str, Any] = {"parser": "plain"}
        try:
            text = content.decode("utf-8")
        except UnicodeDecodeError:
            try:
                text = content.decode("gb18030")
            except UnicodeDecodeError as e:
                return Document(metadata={**meta, "error": f"decode_failed:{e}"})
        meta["chars"] = len(text)
        return Document(content=text, metadata=meta)

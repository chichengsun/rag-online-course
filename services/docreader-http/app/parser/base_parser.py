"""
base_parser：解析器抽象基类，约定 parse_into_text(bytes) -> Document。
"""

from __future__ import annotations

import logging
import os
from abc import ABC, abstractmethod

from app.models.document import Document

logger = logging.getLogger(__name__)


class BaseParser(ABC):
    """
    BaseParser 定义「单格式、无外部存储」的解析边界。
    子类只负责产出 Markdown 与可选的 images 字典（Base64 文本）。
    """

    def __init__(
        self,
        file_name: str = "",
        file_type: str | None = None,
        **kwargs: object,
    ) -> None:
        self.file_name = file_name
        self.file_type = (file_type or "").lower().lstrip(".")
        if not self.file_type and file_name and "." in file_name:
            self.file_type = file_name.rsplit(".", 1)[-1].lower()

    @abstractmethod
    def parse_into_text(self, content: bytes) -> Document:
        """将原始字节解析为 Document；失败时返回空 Document 或带 metadata 提示。"""

    def parse(self, content: bytes) -> Document:
        """parse 为对外统一入口，当前等同 parse_into_text。"""
        logger.debug(
            "parsing with %s file=%s type=%s bytes=%d",
            self.__class__.__name__,
            self.file_name,
            self.file_type,
            len(content),
        )
        return self.parse_into_text(content)

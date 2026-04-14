"""
document：解析结果的领域模型，与 WeKnora docreader.models.document 语义对齐。
content 为 Markdown 文本；images 为「逻辑路径 -> Base64 字符串」，供 HTTP 层展开为 images 列表。
"""

from __future__ import annotations

from typing import Any

from pydantic import BaseModel, Field


class Document(BaseModel):
    """
    Document 表示一次解析输出：正文 + 可选内联图片字典。
    不负责分块、落库与 OCR；仅做格式归一。
    """

    model_config = {"arbitrary_types_allowed": True}

    content: str = Field(default="", description="Markdown 正文")
    images: dict[str, str] = Field(
        default_factory=dict,
        description="键为展示用路径（如 images/uuid.png），值为标准 Base64 文本",
    )
    metadata: dict[str, Any] = Field(default_factory=dict, description="解析元数据")

    def is_useful(self) -> bool:
        """是否含有可展示的正文或非空图片集合（用于链式解析中选主结果）。"""
        if self.images:
            return True
        return bool(self.content and self.content.strip())

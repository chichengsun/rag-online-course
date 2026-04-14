# parser：按扩展名注册解析链（MarkItDown 优先 + 内置回退），对齐 WeKnora docreader 设计。

from app.parser.facade import parse_bytes_to_document

__all__ = ["parse_bytes_to_document"]

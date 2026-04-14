"""
facade：按扩展名选择「MarkItDown 优先 + 内置回退」解析链，对齐 WeKnora registry + FirstParser 思路。
不负责 HTTP 与 MinIO；输出统一为 Document，再由 markdown_post 拆分 data URI。
"""

from __future__ import annotations

import logging
from typing import Sequence, Type

from app.models.document import Document
from app.parser.base_parser import BaseParser
from app.parser.builtin_docx import DocxBuiltinParser
from app.parser.builtin_excel import ExcelBuiltinParser
from app.parser.builtin_image import ImageBuiltinParser
from app.parser.builtin_pdf import PdfBuiltinParser
from app.parser.builtin_plain import PlainTextParser
from app.parser.builtin_pptx import PptxBuiltinParser
from app.parser.markdown_post import split_data_uri_images
from app.parser.markitdown_backend import MarkitdownBackend

logger = logging.getLogger(__name__)

# 与 WeKnora builtin 引擎类似：Office/PDF/表格优先 MarkItDown，再回退轻量实现。
_PARSER_CHAINS: dict[str, tuple[Type[BaseParser], ...]] = {
    "pdf": (MarkitdownBackend, PdfBuiltinParser),
    "docx": (MarkitdownBackend, DocxBuiltinParser),
    "pptx": (MarkitdownBackend, PptxBuiltinParser),
    "xlsx": (MarkitdownBackend, ExcelBuiltinParser),
    "xls": (MarkitdownBackend, ExcelBuiltinParser),
    "html": (MarkitdownBackend, PlainTextParser),
    "htm": (MarkitdownBackend, PlainTextParser),
    "md": (MarkitdownBackend, PlainTextParser),
    "markdown": (MarkitdownBackend, PlainTextParser),
    "txt": (PlainTextParser,),
    "csv": (PlainTextParser,),
    "json": (PlainTextParser,),
    "xml": (PlainTextParser,),
    "png": (ImageBuiltinParser,),
    "jpg": (ImageBuiltinParser,),
    "jpeg": (ImageBuiltinParser,),
    "gif": (ImageBuiltinParser,),
    "webp": (ImageBuiltinParser,),
    "bmp": (ImageBuiltinParser,),
    "tiff": (ImageBuiltinParser,),
    "tif": (ImageBuiltinParser,),
}


def _suffix_from_filename(filename: str) -> str:
    lower = filename.lower()
    if "." not in lower:
        return ""
    return lower.rsplit(".", 1)[-1]


def parse_bytes_to_document(
    filename: str,
    content: bytes,
    *,
    return_images: bool = False,
) -> Document:
    """
    parse_bytes_to_document 对单文件字节执行解析链，并统一做 data URI 拆分。
    return_images 仅影响独立图片类（ImageBuiltinParser）；PDF 整页导出在 parsers 层处理。
    """
    suffix = _suffix_from_filename(filename)
    chain: Sequence[Type[BaseParser]] = _PARSER_CHAINS.get(suffix, (MarkitdownBackend, PlainTextParser))

    last_meta: dict = {}
    for cls in chain:
        try:
            parser = cls(
                file_name=filename,
                file_type=suffix,
                return_images=return_images,
            )
            doc = parser.parse(content)
            doc = split_data_uri_images(doc)
            # MarkItDown 偶发把 PDF 文件头当纯文本返回，避免误判为成功而跳过 PyMuPDF。
            if suffix == "pdf" and doc.content.strip().startswith("%PDF"):
                logger.info("skip false-positive pdf text from %s", cls.__name__)
                continue
            if doc.is_useful():
                meta = dict(doc.metadata)
                meta.setdefault("suffix", suffix)
                meta["chain_parser"] = cls.__name__
                return Document(content=doc.content, images=dict(doc.images), metadata=meta)
            last_meta = dict(doc.metadata)
        except Exception as e:
            logger.info("parser %s failed: %s", cls.__name__, e)
            last_meta = {"chain_error": str(e), "failed_parser": cls.__name__}

    return Document(metadata={"suffix": suffix, **last_meta, "chain_exhausted": True})

# ocr：可选光学识别后端（Tesseract / HTTP），供解析管线在扫描件或图片上补充文本。

from app.ocr.pipeline import enrich_document_with_ocr

__all__ = ["enrich_document_with_ocr"]

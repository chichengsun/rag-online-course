"""
backends：OCR 后端实现。Tesseract 适合容器内开箱即用；HTTP 适合对接自建/云端 OCR 服务。
"""

from __future__ import annotations

import base64
import logging
import shutil
import subprocess
from abc import ABC, abstractmethod
from typing import Optional

import httpx

logger = logging.getLogger(__name__)


class OCRBackend(ABC):
    """OCRBackend 定义「单张光栅图 -> 纯文本」的契约。"""

    @abstractmethod
    def recognize_image_bytes(self, image_bytes: bytes, *, mime_hint: str = "image/png") -> str:
        """对整张图片字节做 OCR，返回 UTF-8 文本；失败时返回空串并打日志。"""


class NoneOCRBackend(OCRBackend):
    """占位后端，不执行识别。"""

    def recognize_image_bytes(self, image_bytes: bytes, *, mime_hint: str = "image/png") -> str:
        return ""


class TesseractOCRBackend(OCRBackend):
    """
    TesseractOCRBackend 依赖系统 tesseract 可执行文件与 pytesseract。
    mime_hint 仅用于日志；实际按图像解码由 Pillow 完成。
    """

    def __init__(self, lang: str = "chi_sim+eng") -> None:
        self._lang = lang or "chi_sim+eng"

    def recognize_image_bytes(self, image_bytes: bytes, *, mime_hint: str = "image/png") -> str:
        if not image_bytes:
            return ""
        if not shutil.which("tesseract"):
            logger.warning("tesseract binary not found in PATH")
            return ""
        try:
            import io

            from PIL import Image
            import pytesseract

            im = Image.open(io.BytesIO(image_bytes))
            if im.mode not in ("RGB", "L"):
                im = im.convert("RGB")
            text = pytesseract.image_to_string(im, lang=self._lang) or ""
            return text.strip()
        except Exception as e:
            logger.info("tesseract ocr failed: %s", e)
            return ""


class HttpOCRBackend(OCRBackend):
    """
    HttpOCRBackend 将图片 POST 到远程服务。
    请求体 JSON：{"image_base64":"<标准 base64>","mime_type":"image/png"}
    响应 JSON：{"text":"..."}；若响应为 text/plain 则整段作为结果。
    """

    def __init__(
        self,
        url: str,
        *,
        timeout_seconds: float = 60.0,
        api_key: str = "",
    ) -> None:
        self._url = url.strip()
        self._timeout = timeout_seconds
        self._api_key = (api_key or "").strip()

    def recognize_image_bytes(self, image_bytes: bytes, *, mime_hint: str = "image/png") -> str:
        if not self._url or not image_bytes:
            return ""
        b64 = base64.standard_b64encode(image_bytes).decode("ascii")
        headers = {"Content-Type": "application/json"}
        if self._api_key:
            headers["Authorization"] = f"Bearer {self._api_key}"
        payload = {"image_base64": b64, "mime_type": mime_hint or "image/png"}
        try:
            with httpx.Client(timeout=self._timeout) as client:
                r = client.post(self._url, json=payload, headers=headers)
                r.raise_for_status()
                ct = (r.headers.get("content-type") or "").lower()
                if "application/json" in ct:
                    data = r.json()
                    if isinstance(data, dict):
                        t = data.get("text") or data.get("result") or ""
                        return str(t).strip()
                    return ""
                return (r.text or "").strip()
        except Exception as e:
            logger.info("http ocr failed: %s", e)
            return ""


def build_ocr_backend(
    backend: str,
    *,
    tesseract_lang: str,
    http_url: str,
    http_timeout_seconds: float,
    http_api_key: str,
) -> OCRBackend:
    """
    build_ocr_backend 按配置名构造后端；未知值回退为 NoneOCRBackend。
    """
    name = (backend or "none").strip().lower()
    if name in ("", "none", "off", "disabled", "no_ocr"):
        return NoneOCRBackend()
    if name in ("tesseract", "tess"):
        return TesseractOCRBackend(lang=tesseract_lang)
    if name in ("http", "api", "remote"):
        if not http_url.strip():
            logger.warning("ocr backend=http but OCR_HTTP_URL empty, OCR disabled")
            return NoneOCRBackend()
        return HttpOCRBackend(
            http_url,
            timeout_seconds=http_timeout_seconds,
            api_key=http_api_key,
        )
    logger.warning("unknown OCR_BACKEND=%s, OCR disabled", backend)
    return NoneOCRBackend()

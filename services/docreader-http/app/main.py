"""
docreader-http 入口：提供 /health 与 /v1/read，从 MinIO 或 URL 拉取对象后在进程内解析。
"""

from __future__ import annotations

import logging
from typing import Any, Optional

import httpx
from fastapi import Depends, FastAPI, Header, HTTPException
from minio import Minio
from pydantic import BaseModel, Field, model_validator

from app.config import Settings
from app.parsers import extract_content

logger = logging.getLogger(__name__)

app = FastAPI(title="docreader-http", version="0.2.0")


def get_settings() -> Settings:
    return Settings()


def _minio_client(s: Settings) -> Minio:
    return Minio(
        s.minio_endpoint,
        access_key=s.minio_access_key,
        secret_key=s.minio_secret_key,
        secure=s.minio_use_ssl,
    )


def verify_internal_token(
    s: Settings = Depends(get_settings),
    x_internal_token: Optional[str] = Header(default=None, alias="X-Internal-Token"),
) -> None:
    """校验内网共享密钥；未配置 internal_token 时不启用。"""
    if not (s.internal_token or "").strip():
        return
    if x_internal_token != s.internal_token:
        raise HTTPException(status_code=401, detail="invalid X-Internal-Token")


class V1ReadRequest(BaseModel):
    """
    V1ReadRequest 描述一次解析请求。
    object_key 与 url 二选一：前者由服务从 MinIO 拉取，后者为可访问的文件 URL（如预签名）。
    """

    object_key: Optional[str] = None
    bucket: Optional[str] = None
    url: Optional[str] = None
    file_name: Optional[str] = None
    return_images: bool = False
    # use_ocr 为 true 时强制对 PDF 逐页 OCR、对独立图片 OCR（仍受服务端 OCR_BACKEND 约束）。
    use_ocr: bool = False

    @model_validator(mode="after")
    def one_source(self) -> V1ReadRequest:
        has_key = bool(self.object_key and self.object_key.strip())
        has_url = bool(self.url and self.url.strip())
        if has_key == has_url:
            raise ValueError("必须且只能指定 object_key 或 url 之一")
        return self


class ImageItem(BaseModel):
    """ImageItem 表示解析结果中的内联图片（Base64）。"""

    filename: str
    mime_type: str
    data_base64: str


class V1ReadResponse(BaseModel):
    """V1ReadResponse 与 Go 侧 docreaderhttp 包反序列化字段一致。"""

    markdown: str = ""
    images: list[ImageItem] = Field(default_factory=list)
    metadata: dict[str, Any] = Field(default_factory=dict)
    error: str = ""


def _filename_from_key(object_key: str, fallback: str) -> str:
    part = object_key.rstrip("/").split("/")[-1]
    return part or fallback


async def _load_file_bytes(req: V1ReadRequest, s: Settings) -> tuple[bytes, str]:
    """从 MinIO 或 HTTP URL 加载原始文件字节与文件名。"""
    if req.url:
        async with httpx.AsyncClient(timeout=s.http_fetch_timeout_seconds) as client:
            r = await client.get(req.url.strip())
            r.raise_for_status()
            name = req.file_name or "remote.bin"
            return r.content, name

    bucket = (req.bucket or s.minio_bucket).strip()
    key = req.object_key.strip()
    name = req.file_name or _filename_from_key(key, "object.bin")
    client = _minio_client(s)
    try:
        obj = client.get_object(bucket, key)
        try:
            data = obj.read()
        finally:
            obj.close()
            obj.release_conn()
    except Exception as e:
        logger.exception("minio get_object failed bucket=%s key=%s", bucket, key)
        raise HTTPException(status_code=400, detail=f"minio get_object failed: {e}") from e
    return data, name


@app.get("/health")
async def health() -> dict[str, str]:
    """存活探针：本服务不依赖外部解析引擎。"""
    return {"status": "ok"}


@app.post("/v1/read", response_model=V1ReadResponse)
async def v1_read(
    body: V1ReadRequest,
    s: Settings = Depends(get_settings),
    _auth: None = Depends(verify_internal_token),
) -> V1ReadResponse:
    """
    v1_read：拉取文件后在容器内解析为 Markdown；可选 return_images、use_ocr 与网关 OCR 环境变量配合做页图与 OCR。
    """
    try:
        raw, fname = await _load_file_bytes(body, s)
    except HTTPException:
        raise

    md, img_dicts, meta, err = extract_content(
        fname,
        raw,
        return_images=body.return_images,
        max_pdf_pages_as_images=s.parse_max_pages_for_images,
        settings=s,
        use_ocr=body.use_ocr,
    )
    images = [ImageItem(**d) for d in img_dicts]
    if err:
        return V1ReadResponse(markdown=md, images=images, metadata=meta, error=err)
    return V1ReadResponse(markdown=md, images=images, metadata=meta, error="")

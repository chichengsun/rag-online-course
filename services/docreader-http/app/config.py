"""应用配置：从环境变量加载 MinIO 与鉴权参数。"""

from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """Settings 描述 docreader-http 进程级配置，与 docker-compose 注入的环境变量对齐。"""

    model_config = SettingsConfigDict(env_file=".env", extra="ignore")

    listen_host: str = "0.0.0.0"
    listen_port: int = 8080

    minio_endpoint: str = "minio:9000"
    minio_access_key: str = "minioadmin"
    minio_secret_key: str = "minioadmin"
    minio_use_ssl: bool = False
    minio_bucket: str = "course-resources"

    # http_fetch_timeout_seconds 从 URL 拉取对象时的超时（秒）。
    http_fetch_timeout_seconds: float = 120.0
    # parse_timeout_seconds 本地解析单文件的最大耗时提示（由调用方 HTTP 客户端控制为主，此处仅作元数据参考可略）。
    parse_max_pages_for_images: int = 20

    internal_token: str = ""

    # --- OCR（可选）：none | tesseract | http；默认关闭以保持镜像与行为兼容。
    ocr_backend: str = "none"
    # ocr_tesseract_lang Tesseract --lang，中文课件建议 chi_sim+eng。
    ocr_tesseract_lang: str = "chi_sim+eng"
    # ocr_http_url 远程 OCR 服务 URL（POST JSON，见 README）。
    ocr_http_url: str = ""
    ocr_http_timeout_seconds: float = 120.0
    ocr_http_api_key: str = ""
    # ocr_auto_scan_pdf 为 true 且后端非 none 时，正文极短的 PDF 自动整页 OCR。
    ocr_auto_scan_pdf: bool = True
    # ocr_auto_standalone_image 为 true 时，对独立图片文件自动 OCR（无需请求体 use_ocr）。
    ocr_auto_standalone_image: bool = True
    # ocr_max_pages PDF OCR 最大页数（与 parse_max_pages_for_images 取较小值在管线中生效）。
    ocr_max_pages: int = 20

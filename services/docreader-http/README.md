# docreader-http

面向 Go 后端的 **HTTP 解析网关**：从 MinIO 拉取对象（或拉取可访问 URL），在进程内解析为 Markdown 与可选图片，返回 `markdown` / `images` / `metadata` JSON，供 [internal/integration/docreaderhttp](../../internal/integration/docreaderhttp/client.go) 消费。

解析管线在设计上参考 [Tencent/WeKnora](https://github.com/Tencent/WeKnora) 的 **`docreader`** 模块（注册表 + 解析链 + `Document{content, images}`），在本仓库中以 **HTTP + 更轻依赖** 的方式落地：**不引入 gRPC、PaddleOCR、多进程巨型 Docx 管线**。

## 解析设计（与 WeKnora 对齐点）

| 概念 | 说明 |
|------|------|
| **Document** | `content`（Markdown 正文）+ `images`（逻辑路径 → Base64 字符串），见 `app/models/document.py`。 |
| **解析链** | 对 PDF / DOCX / PPTX / XLSX 等：**MarkItDown 优先**，失败或无效再回退 **PyMuPDF / python-docx / python-pptx / openpyxl**，见 `app/parser/facade.py`。 |
| **Markdown 后处理** | 将正文里的 `data:image/*;base64,...` 拆出到 `images`，避免巨型 data URI 留在 Markdown 中，见 `app/parser/markdown_post.py`（对应 WeKnora 的 Markdown 内联图处理思路）。 |
| **DOCX 内联图** | 回退路径下用 **python-docx + xpath** 抽取段落内嵌图片（简化自 WeKnora Docx 逻辑，单进程），见 `app/parser/builtin_docx.py`。 |
| **PDF `return_images`** | 在已有正文/拆图之后，可选再 **按页光栅化** 追加 PNG（页数上限可配），见 `app/parser/builtin_pdf.py` 中 `append_pdf_page_images`。 |
| **OCR** | 可选 **Tesseract**（镜像内）或 **HTTP** 远程服务，见 `app/ocr/`；扫描 PDF 可自动整页识别，见下文。 |

## 支持格式（概要）

| 类型 | 说明 |
|------|------|
| PDF / DOCX / PPTX / XLSX | MarkItDown 优先，再回退内置解析器 |
| XLS | 主要依赖 MarkItDown；无本地 openpyxl 回退 |
| HTML / MD | MarkItDown → 纯文本回退 |
| TXT / CSV / JSON / XML | UTF-8 / GB18030 文本 |
| 常见图片 | 元数据说明；`return_images` 时输出 Base64；配置 OCR 时可自动/按需识别图中文字 |
| .doc / .ppt | 明确返回不支持（需先转 OOXML） |

## 启动顺序（Docker Compose）

解析栈挂在 profile **`parser`** 与 **`dev`** 上（与 `make dev` 一致），**仅依赖 MinIO**（`depends_on` + `service_healthy`）。

### 与后端一键开发（推荐）

```bash
make dev
```

会依次：`docker compose --profile dev up -d`（含 docreader-http）→ 等待 `http://localhost:8090/health` → 数据库迁移 → `go run` 启动 API。首次需构建镜像：`docker compose --profile dev build docreader-http`。

前端在另一终端执行 `npm run dev`（或项目既有前端命令）即可联调；Go 默认从 `config/config.yaml` 读取 **`docreader.base_url: http://localhost:8090`**。

### 仅起解析栈

```bash
docker compose --profile parser up -d
```

- **本服务**：映射宿主机 **`8090:8080`**，`GET /health` 返回 `{"status":"ok"}`。

## 环境变量

| 变量 | 说明 |
|------|------|
| `MINIO_ENDPOINT` | MinIO 地址（Compose 内通常为 `minio:9000`） |
| `MINIO_ACCESS_KEY` / `MINIO_SECRET_KEY` | MinIO 凭证 |
| `MINIO_USE_SSL` | `true` / `false` |
| `MINIO_BUCKET` | 默认桶名，与 Go `minio.bucket` 一致（如 `course-resources`） |
| `HTTP_FETCH_TIMEOUT_SECONDS` | 从 URL 拉取对象超时（秒，默认 120） |
| `PARSE_MAX_PAGES_FOR_IMAGES` | PDF 转图时的最大页数（默认 20） |
| `INTERNAL_TOKEN` | 非空时要求请求头 `X-Internal-Token` 一致 |
| `OCR_BACKEND` | `none`（默认）\|`tesseract`\|`http` |
| `OCR_TESSERACT_LANG` | Tesseract `-l`，默认 `chi_sim+eng` |
| `OCR_HTTP_URL` | `OCR_BACKEND=http` 时必填；`POST` JSON 见下节 |
| `OCR_HTTP_TIMEOUT_SECONDS` | 远程 OCR 超时（默认 120） |
| `OCR_HTTP_API_KEY` | 可选；非空时带 `Authorization: Bearer <key>` |
| `OCR_AUTO_SCAN_PDF` | `true`/`false`，正文极短 PDF 是否自动整页 OCR（默认 `true`） |
| `OCR_AUTO_STANDALONE_IMAGE` | 对 png/jpg 等独立文件是否自动 OCR（默认 `true`） |
| `OCR_MAX_PAGES` | PDF OCR 最大页数（默认 20，与转图上限取较小值） |

### 远程 OCR（`OCR_BACKEND=http`）

`POST` 请求体 JSON：

```json
{"image_base64":"<标准 Base64>","mime_type":"image/png"}
```

响应：`Content-Type: application/json` 且 body 含 `text` 或 `result` 字段（字符串）；或 `text/plain` 整段为识别结果。

## HTTP 接口

- `GET /health`：进程存活检查。
- `POST /v1/read`：`Content-Type: application/json`  
  - **二选一**：`object_key`（从 MinIO 拉取，可选 `bucket`）或 `url`（HTTP GET 拉取，如预签名 URL）。  
  - 可选：`file_name`、`return_images`、`use_ocr`（为 `true` 时对 PDF 强制逐页 OCR、对独立图片 OCR）。  
  - 响应：`markdown`、`images[]`（`filename` / `mime_type` / `data_base64`）、`metadata`、`error`。

### curl 示例

```bash
curl -sS http://localhost:8090/health

curl -sS -X POST http://localhost:8090/v1/read \
  -H 'Content-Type: application/json' \
  -H "X-Internal-Token: $DOCREADER_INTERNAL_TOKEN" \
  -d '{"object_key":"courses/1/chapters/1/your.pdf","bucket":"course-resources"}'
```

## Go 后端配置

在 `config/config.yaml` 或环境中设置：

- `docreader.base_url` / `DOCREADER_BASE_URL`：如本机 `http://localhost:8090`，或与 API 同 Docker 网络时 `http://docreader-http:8080`。
- `docreader.internal_token` / `DOCREADER_INTERNAL_TOKEN`：与 `INTERNAL_TOKEN` 对齐。
- `docreader.use_ocr` / `DOCREADER_USE_OCR`：教师「解析资源」时是否向网关传 `use_ocr: true`（网关须启用 `OCR_BACKEND`）。
- 分块与嵌入向量见迁移 **`000005_resource_embedding_chunks`**；教师解析接口：  
  `POST /api/v1/teacher/resources/:resourceId/parse`（仅调 docreader，不落解析结果表）

## 限制说明

- **OCR 准确率**：Tesseract 对复杂排版、手写、低分辨率效果有限；生产可改用 `http` 对接更强模型。
- **镜像体积**：引入 MarkItDown 会连带 **onnxruntime / magika / pandas** 等，镜像比纯 PyMuPDF 方案更大；若需极简部署可再评估裁剪 extras。
- `return_images: true` 时响应体可能很大，请谨慎在前端直接展示。

## 本地构建镜像

```bash
docker build -t roc-docreader-http:latest -f services/docreader-http/Dockerfile services/docreader-http
```

Compose 中 `docreader-http` 服务默认使用 `build: ./services/docreader-http`。

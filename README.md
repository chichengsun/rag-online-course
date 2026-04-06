# rag-online-course

基于 Go + Gin + PostgreSQL + Redis + MinIO 的在线课程系统，当前已包含教师端知识库分块/嵌入、模型管理、课程知识库对话（RAG + 流式）等能力。

## 快速开始

1. 复制配置：
   - `cp .env.example .env`
2. 启动中间件（Postgres/Redis/MinIO/docreader）：
   - `make dev`
3. 执行迁移：
   - `make migrate-up`
4. 启动后端：
   - `go run ./cmd/server`
5. 启动前端（新终端）：
   - `cd frontend && npm install && npm run dev`

## 常用命令

- 启动依赖：`make dev`
- 停止依赖：`make down`
- 停止并清理数据卷：`make clean`
- 迁移升级：`make migrate-up`
- 迁移回滚：`make migrate-down`
- 强制版本：`make migrate-force version=<n>`
- 修复脏版本（历史嵌入失败场景）：`make migrate-repair-after-embedding-fail`

## 已实现能力

- 认证与课程基础能力：
  - 学生/教师登录（`email` 或 `username`）
  - 教师课程、章节、资源管理
  - 学生选课、目录学习、进度更新
- 教师知识库管理：
  - 资源解析、分块预览与保存
  - 分块编辑/删除/重排、查看全文
  - 向量嵌入与状态管理
- 教师模型管理：
  - `qa` / `embedding` / `rerank` 三类模型持久化
  - 模型连通性测试
- 教师对话管理：
  - 会话创建、重命名、删除、历史分页
  - SSE 流式对话
  - 会话内切换问答模型
  - Markdown 回答渲染
  - 回答底部引用块 + 完整片段查看

## RAG 对话链路（当前实现）

每次提问会走以下步骤：

1. 加载会话历史（最近若干轮）
2. 意图识别（`simple` / `rag`）
3. `rag` 路径：
   - 问题改写（rewrite）
   - 关键词分解（LLM 输出关键词数组）
   - 语义检索（向量）与关键词检索（token）并行
   - RRF 融合
   - 可选 rerank
   - 生成回答（支持 SSE）
4. 引用规范化与落库（包含 `full_content`）
5. 若检索为空，返回配置兜底文案（不会中断会话）

## 检索参数（按每轮对话可调）

前端顶部支持以下参数并随请求下发：

- `top_k`
- `semantic_min_score`
- `keyword_min_score`

当阈值过高导致两路候选均为空时，后端会自动回退到未过滤候选，并记录 warning 日志。

## 关键配置（`config/config.yaml`）

`rag` 段包含以下可配置 prompt：

- `qa_system_prompt`
- `intent_system_prompt`
- `rewrite_system_prompt`
- `keyword_system_prompt`
- `simple_qa_system_prompt`
- `fallback_answer`

以上 prompt 默认采用 Markdown 结构（角色/问题/目标/要求），便于持续迭代。

## API 与文档

- OpenAPI：`docs/openapi.yaml`
- 实现说明：`docs/implementation.md`

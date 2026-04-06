# rag-online-course

Go + Gin + Postgres + Redis + MinIO 的在线课程系统基础实现。

## 快速开始
1. 一键启动中间件：
   - `docker compose up -d`
2. 复制 `.env.example` 为 `.env` 并按需修改配置。
3. 执行数据库迁移：
   - `make migrate-up`
4. 本地运行服务（服务本身不在 compose 里）：
   - `go mod tidy`
   - `go run ./cmd/server`

## 停止中间件
- `docker compose down`
- 如需清理数据卷：`docker compose down -v`

## 核心能力
- 角色登录：学生 / 教师
- 登录账号支持：`email` 或 `username`
- 教师：课程、章节、资源管理与排序
- 学生：选课、课程目录学习、进度更新
- 文件存储：MinIO 对象存储，数据库仅存 `object_url` 和 `object_key`

## Migration 工具
- 引入 `golang-migrate`，入口在 `cmd/migrate`
- 常用命令：
  - `make migrate-up`
  - `make migrate-down`
  - `make migrate-force version=1`

## API 文档
- 见 `docs/openapi.yaml`
- 设计说明：`docs/implementation.md`

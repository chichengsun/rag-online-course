APP_NAME=rag-online-course

.PHONY: run server migrate-up migrate-down migrate-force migrate-repair-after-embedding-fail test test-integration wait-deps wait-dev-deps compose-up compose-up-dev compose-down compose-logs dev dev-lite

server:
	set -a; [ -f .env ] && . ./.env; set +a; go run ./cmd/server

run:
	$(MAKE) server

migrate-up:
	set -a; [ -f .env ] && . ./.env; set +a; go run ./cmd/migrate -action up

migrate-down:
	set -a; [ -f .env ] && . ./.env; set +a; go run ./cmd/migrate -action down

# 强制写入 schema_migrations 版本并清除 dirty（不执行 SQL）。用于某次 up 中途失败后解锁。
# 例：000005 因缺少 pgvector 失败 → Dirty version 5 → 装好 pgvector 镜像后：
#   make migrate-force version=4 && make migrate-up
migrate-force:
	set -a; [ -f .env ] && . ./.env; set +a; go run ./cmd/migrate -action force -version $(version)

# 从 dirty version 5（历史 000005 失败）回退到 4 并重新 up；等价于 migrate-force version=4 + migrate-up。
migrate-repair-after-embedding-fail:
	$(MAKE) migrate-force version=4
	$(MAKE) migrate-up

test:
	go test ./...

test-integration:
	$(MAKE) compose-up
	$(MAKE) wait-deps
	$(MAKE) migrate-up
	set -a; [ -f .env ] && . ./.env; set +a; RUN_INTEGRATION_TESTS=1 go test ./tests/integration -v

wait-deps:
	@echo "等待 Postgres 就绪..."
	@for i in $$(seq 1 60); do \
		docker compose exec -T postgres pg_isready -U postgres -d online_course >/dev/null 2>&1 && break; \
		if [ $$i -eq 60 ]; then echo "Postgres 未就绪"; exit 1; fi; \
		sleep 1; \
	done
	@echo "等待 Redis 就绪..."
	@for i in $$(seq 1 60); do \
		docker compose exec -T redis redis-cli ping >/dev/null 2>&1 && break; \
		if [ $$i -eq 60 ]; then echo "Redis 未就绪"; exit 1; fi; \
		sleep 1; \
	done
	@echo "等待 MinIO 就绪..."
	@for i in $$(seq 1 60); do \
		curl -fsS http://localhost:9000/minio/health/live >/dev/null 2>&1 && break; \
		if [ $$i -eq 60 ]; then echo "MinIO 未就绪"; exit 1; fi; \
		sleep 1; \
	done
	@echo "所有依赖已就绪。"

compose-up:
	docker compose up -d

# 开发：额外启动 docreader-http（进程内解析）。
compose-up-dev:
	docker compose --profile dev up -d

compose-down:
	docker compose down

compose-logs:
	docker compose logs -f --tail=200

wait-dev-deps: wait-deps
	@echo "等待 docreader-http 就绪..."
	@for i in $$(seq 1 60); do \
		curl -fsS http://localhost:8090/health >/dev/null 2>&1 && break; \
		if [ $$i -eq 60 ]; then echo "docreader-http 未就绪（首次需构建: docker compose --profile dev build docreader-http）"; exit 1; fi; \
		sleep 2; \
	done
	@echo "docreader-http 已就绪（http://localhost:8090）。"

# 一键本地开发：含 docreader-http → 迁移 → Go API。
dev:
	$(MAKE) compose-up-dev
	$(MAKE) wait-dev-deps
	$(MAKE) migrate-up
	$(MAKE) server

# 仅 Postgres/Redis/MinIO + Go，不启 docreader-http。
dev-lite:
	$(MAKE) compose-up
	$(MAKE) wait-deps
	$(MAKE) migrate-up
	$(MAKE) server

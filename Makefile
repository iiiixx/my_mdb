GO_CMD=go run ./cmd/app

PYTHON?=python3
RECS_ENTRY=recommender/main.py

RECS_HOST=0.0.0.0
RECS_PORT=50051
RECS_MODEL_DIR=recommender/model/current
RECS_WORKERS=8
RECS_LOG_LEVEL=INFO

REC_GRPC_ADDR=localhost:50051

.PHONY: api recs dev stop

api:
	@echo "Starting Go API..."
	REC_GRPC_ADDR=$(REC_GRPC_ADDR) \
	$(GO_CMD)

recs:
	@echo "Starting Python recommender..."
	RECS_HOST=$(RECS_HOST) \
	RECS_PORT=$(RECS_PORT) \
	RECS_MODEL_DIR=$(RECS_MODEL_DIR) \
	RECS_WORKERS=$(RECS_WORKERS) \
	RECS_LOG_LEVEL=$(RECS_LOG_LEVEL) \
	$(PYTHON) $(RECS_ENTRY)

dev:
	@echo "Starting both services..."
	@make -j2 recs api

stop:
	@echo "Stopping services..."
	@pkill -f "$(RECS_ENTRY)" || true
	@pkill -f "go run ./cmd/app" || true
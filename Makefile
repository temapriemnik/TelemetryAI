.PHONY: help install backend frontend test clean db-reset

PYTHON := python3
PYTHONPATH := $(shell pwd)
FRONTEND_DIR := frontend
NODE_BIN := /tmp/node-v22.14.0-linux-x64/bin
PATH := $(NODE_BIN):$(PATH)

help:
	@echo "TelemetryAI Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  make install    - Install all dependencies"
	@echo "  make backend    - Start backend server"
	@echo "  make frontend   - Start frontend dev server"
	@echo "  make dev        - Start both servers"
	@echo "  make test       - Run tests"
	@echo "  make clean      - Clean cache and temp files"
	@echo "  make db-reset   - Reset database"

install:
	~/.local/bin/pip install --break-system-packages -r requirements.txt
	cd $(FRONTEND_DIR) && npm install

backend:
	@echo "Starting backend on http://localhost:8000"
	cd $(shell pwd) && PYTHONPATH=$(shell pwd) $(PYTHON) -c "import uvicorn; from backend.app.main import app; uvicorn.run(app, host='0.0.0.0', port=8000)"

frontend:
	@echo "Starting frontend on http://localhost:5173"
	cd $(FRONTEND_DIR) && npm run dev

dev:
	@echo "Starting TelemetryAI dev servers..."
	@echo "Backend: http://localhost:8000"
	@echo "API Docs: http://localhost:8000/docs"
	@echo "Frontend: http://localhost:5173"
	@echo ""
	@echo "Press Ctrl+C to stop both servers"
	cd $(FRONTEND_DIR) && npm run dev & \
	cd $(shell pwd) && PYTHONPATH=$(shell pwd) $(PYTHON) -c "import uvicorn; from backend.app.main import app; uvicorn.run(app, host='0.0.0.0', port=8000)"

test:
	cd $(shell pwd) && PYTHONPATH=$(shell pwd) $(PYTHON) -m pytest -v

clean:
	find . -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null || true
	find . -type f -name "*.pyc" -delete 2>/dev/null || true
	rm -rf $(FRONTEND_DIR)/node_modules/.vite 2>/dev/null || true
	rm -f telemetryai.db

db-reset:
	rm -f telemetryai.db
	@echo "Database reset. Run 'make backend' to recreate tables."
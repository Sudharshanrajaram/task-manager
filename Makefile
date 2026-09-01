.PHONY: dev dev-backend dev-frontend build test

dev:
	@echo "Starting backend and frontend..."
	@make -j 2 dev-backend dev-frontend

dev-backend:
	cd backend && go run cmd/api/main.go

dev-frontend:
	cd frontend && npm run dev

build:
	cd backend && go build -o bin/taskflow-api cmd/api/main.go
	cd frontend && npm run build

test:
	cd backend && go test -v -race ./...


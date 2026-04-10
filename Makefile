.PHONY: help proto build-go build-web build run-allinone run-coordinator run-worker run-worker2 docker docker-scale clean

PROTOC := $(HOME)/go/bin/protoc
PROTO_GO_OUT := internal/gen/fractal

help: ## Show this help
	@echo "Fractal Cluster — available targets:"
	@echo
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_-]+:.*?## / {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo

proto: ## Regenerate Go gRPC stubs from proto/fractal.proto
	mkdir -p $(PROTO_GO_OUT)
	$(PROTOC) --go_out=$(PROTO_GO_OUT) --go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_GO_OUT) --go-grpc_opt=paths=source_relative \
		-I proto proto/fractal.proto

build-go: ## Build all Go binaries (allinone, coordinator, worker)
	go build -o bin/allinone ./cmd/allinone
	go build -o bin/worker ./cmd/worker
	go build -o bin/coordinator ./cmd/coordinator

build-web: ## Build the React frontend into web/dist
	cd web && npm install && npm run build

build: build-go build-web ## Build everything (Go binaries + frontend)

run-allinone: ## Run coordinator+worker+webserver in a single process (no gRPC)
	go run ./cmd/allinone -port 8080 -web web/dist

run-coordinator: ## Run the distributed coordinator (HTTP :8080, gRPC :9090)
	go run ./cmd/coordinator -port 8080 -grpc-port 9090 -web web/dist

run-worker: ## Run a distributed worker on :50051
	go run ./cmd/worker -port 50051 -coordinator localhost:9090 -advertise localhost:50051

run-worker2: ## Run a second distributed worker on :50052
	go run ./cmd/worker -port 50052 -coordinator localhost:9090 -advertise localhost:50052

docker: ## Bring up the full stack via docker compose
	docker compose -f docker/docker-compose.yaml up --build

docker-scale: ## Bring up the stack with an extra scaled worker
	docker compose -f docker/docker-compose.yaml up --build --scale worker2=2

clean: ## Remove build artifacts (bin/, web/dist, web/node_modules)
	rm -rf bin/ web/dist web/node_modules

.DEFAULT_GOAL := help

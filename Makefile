.PHONY: proto build-go build-web build run-worker run-coordinator docker clean

PROTOC := $(HOME)/go/bin/protoc
PROTO_GO_OUT := internal/gen/fractal

proto:
	mkdir -p $(PROTO_GO_OUT)
	$(PROTOC) --go_out=$(PROTO_GO_OUT) --go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_GO_OUT) --go-grpc_opt=paths=source_relative \
		-I proto proto/fractal.proto

build-go:
	go build -o bin/worker ./cmd/worker
	go build -o bin/coordinator ./cmd/coordinator

build-web:
	cd web && npm install && npm run build

build: build-go build-web

run-worker:
	go run ./cmd/worker -port 50051

run-worker2:
	go run ./cmd/worker -port 50052

run-coordinator:
	go run ./cmd/coordinator -port 8080 -workers workers.yaml -web web/dist

docker:
	docker compose up --build

docker-scale:
	docker compose up --build --scale worker2=2

clean:
	rm -rf bin/ web/dist web/node_modules

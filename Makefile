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

run-coordinator:
	go run ./cmd/coordinator -port 8080 -grpc-port 9090 -web web/dist

run-worker:
	go run ./cmd/worker -port 50051 -coordinator localhost:9090 -advertise localhost:50051

run-worker2:
	go run ./cmd/worker -port 50052 -coordinator localhost:9090 -advertise localhost:50052

docker:
	docker compose -f docker/docker-compose.yaml up --build

docker-scale:
	docker compose -f docker/docker-compose.yaml up --build --scale worker2=2

clean:
	rm -rf bin/ web/dist web/node_modules

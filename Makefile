.PHONY: build test image
build:
	go build -o bin/kubecis ./cmd/kubecis
test:
	go test ./...
image:
	docker build -t kubecis:0.1.0 .

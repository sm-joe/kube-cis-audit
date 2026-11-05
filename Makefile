.PHONY: test vet build

test:
	go test ./... -count=1
vet:
	go vet ./...
build:
	go build -trimpath -o kubecis ./cmd/kubecis

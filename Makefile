.PHONY: build test vet fmt image image-push clean

VERSION ?= dev
CONTAINER_ENGINE ?= podman
IMAGE ?= kubecis:$(VERSION)

build:
	go build -trimpath -ldflags="-s -w -X github.com/sm-joe/kube-cis-audit/internal/audit.BuildVersion=$(VERSION)" -o kubecis ./cmd/kubecis

test:
	go test ./... -count=1

vet:
	go vet ./...

fmt:
	gofmt -w ./cmd ./internal

image:
	$(CONTAINER_ENGINE) build --build-arg VERSION=$(VERSION) -t $(IMAGE) .

image-push:
	$(CONTAINER_ENGINE) push $(IMAGE)

clean:
	go clean

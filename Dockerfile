# syntax=docker/dockerfile:1.7
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS build
WORKDIR /src

ARG TARGETOS
ARG TARGETARCH
ARG VERSION=3.0.0-dev

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath -buildvcs=false \
    -ldflags="-s -w -X github.com/sm-joe/kube-cis-audit/internal/audit.BuildVersion=${VERSION}" \
    -o /out/kubecis ./cmd/kubecis

FROM gcr.io/distroless/static-debian12:nonroot AS runtime
WORKDIR /app
COPY --from=build --chown=65532:65532 /out/kubecis /usr/local/bin/kubecis
COPY --chown=65532:65532 benchmarks /app/benchmarks
USER 65532:65532

LABEL org.opencontainers.image.title="KubeCIS" \
      org.opencontainers.image.description="Read-only Kubernetes security and CIS benchmark auditor" \
      org.opencontainers.image.source="https://github.com/sm-joe/kube-cis-audit"

ENTRYPOINT ["/usr/local/bin/kubecis"]
CMD ["scan", "--benchmark", "auto", "--target", "all", "--output", "json", "--exit-code-on-fail=false"]

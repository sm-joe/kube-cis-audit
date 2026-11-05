# KubeCIS Native: no kube-bench executable, source or benchmark bundle.
FROM --platform=$BUILDPLATFORM golang:1.27.3-bookworm AS build

WORKDIR /src

ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG VERSION=3.1.0-native-eks-rc1

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -trimpath -buildvcs=false \
    -ldflags="-s -w -X github.com/sm-joe/kube-cis-audit/internal/audit.BuildVersion=${VERSION}" \
    -o /out/kubecis ./cmd/kubecis
RUN go version -m /out/kubecis


FROM alpine:3.24.3

RUN apk upgrade --no-cache \
    && apk add --no-cache ca-certificates \
    && addgroup -S -g 65532 kubecis \
    && adduser -S -D -H -u 65532 -G kubecis kubecis

COPY --from=build /out/kubecis /usr/local/bin/kubecis
COPY --chown=65532:65532 benchmarks /app/benchmarks

WORKDIR /app

ENV KUBECIS_CONFIG_DIR=/app/benchmarks

USER 65532:65532

LABEL org.opencontainers.image.title="KubeCIS Native" \
      org.opencontainers.image.description="Independent native Go Kubernetes security auditor" \
      org.opencontainers.image.source="https://github.com/sm-joe/kube-cis-audit"
ENTRYPOINT ["/usr/local/bin/kubecis"]

CMD ["scan", "--benchmark", "auto", "--target", "all", "--engine", "native", "--output", "json", "--exit-code-on-fail=false"]

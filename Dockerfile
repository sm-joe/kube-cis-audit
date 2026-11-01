FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/kubecis ./cmd/kubecis

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/kubecis /usr/local/bin/kubecis
COPY benchmarks /app/benchmarks
USER 65532:65532
ENTRYPOINT ["/usr/local/bin/kubecis"]
CMD ["scan", "--benchmark", "auto"]

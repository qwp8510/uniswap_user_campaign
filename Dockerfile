# Build the manager binary
FROM golang:1.21.0-bookworm as builder

WORKDIR /home/nonroot

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY cmd/ cmd/
COPY internal/ internal/
COPY migrations/ migrations/
COPY pkg/ pkg/

RUN go build -a -o app main.go

FROM debian:12

RUN apt-get update && \
    apt-get install -y \
    ca-certificates \
    curl \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /home/nonroot
COPY --from=builder --chown=65532:65532 /home/nonroot/app /home/nonroot/app
COPY --from=builder --chown=65532:65532 /home/nonroot/migrations /home/nonroot/migrations

USER 65532:65532

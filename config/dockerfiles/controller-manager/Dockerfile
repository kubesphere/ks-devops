# Build the manager binary
FROM golang:1.23 AS builder

ARG GOPROXY
WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Copy the go source
COPY controllers/ controllers/
COPY cmd/ cmd/
COPY assets assets/
COPY pkg/ pkg/

# Build
RUN CGO_ENABLED=0 GO111MODULE=on go build -a -o controller-manager cmd/controller/main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/controller-manager .
USER nonroot:nonroot

ENTRYPOINT ["/controller-manager"]

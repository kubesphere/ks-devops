FROM golang:1.23 as downloader
RUN go install github.com/linuxsuren/http-downloader@v0.0.98
RUN http-downloader install kubesphere-sigs/ks@v0.0.73

# Build the tool binary
FROM golang:1.23 AS builder

ARG GOPROXY
WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Copy the go source
COPY cmd/ cmd/
COPY assets assets/
COPY pkg/ pkg/

# Build
RUN CGO_ENABLED=0 GO111MODULE=on go build -a -o devops-tools cmd/tools/main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
COPY --from=builder /workspace/devops-tools /usr/local/bin/devops-tools
COPY --from=downloader /usr/local/bin/ks /usr/local/bin/ks
USER nonroot:nonroot

CMD ["devops-tools"]

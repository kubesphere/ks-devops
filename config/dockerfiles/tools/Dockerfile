# Build the manager binary
FROM golang:1.16 as builder

ARG GOPROXY
WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd/ cmd/
COPY pkg/ pkg/

# Build
RUN CGO_ENABLED=0 GO111MODULE=on go build -a -o jwt cmd/tools/jwt/jwt_cmd.go

FROM golang:1.16 as downloader
RUN go install github.com/linuxsuren/http-downloader@v0.0.49
RUN http-downloader install kubesphere-sigs/ks@v0.0.60

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/jwt .
COPY --from=downloader /usr/local/bin/ks /usr/local/bin/ks
USER nonroot:nonroot

CMD ["/jwt"]

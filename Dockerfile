# Build the manager binary
FROM golang:1.13 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY pkg/ pkg/
COPY controllers/ controllers/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o genoa main.go

FROM alpine:3.10

RUN wget https://get.helm.sh/helm-v3.2.4-linux-amd64.tar.gz \
    && tar -xvf helm-v3.2.4-linux-amd64.tar.gz \
    && mv linux-amd64/helm /usr/local/bin \
    && rm -f helm-v3.2.4-linux-amd64.tar.gz && rm -rf linux-amd64
COPY repositories.yaml /root/.config/helm/repositories.yaml
RUN helm repo udpate
COPY --from=builder /workspace/genoa /

ENTRYPOINT ["/genoa"]

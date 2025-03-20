# Build stage: compile the manager binary
FROM docker.io/golang:1.23 AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
# Copy module manifests and download dependencies
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

# Copy the Go source code
COPY cmd/main.go cmd/main.go
COPY api/ api/
COPY internal/ internal/

# Build the binary (statically linked)
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -o manager cmd/main.go

# Debug: list the contents of /workspace to verify the binary exists.
RUN ls -l /workspace

# Runtime stage: use Google Cloud SDK alpine image which includes gsutil and /bin/sh.
FROM google/cloud-sdk:alpine
WORKDIR /

# Install MySQL client (which provides mysqldump) along with any other needed packages
RUN apk add --no-cache mysql-client
# Copy the operator binary from the builder stage
COPY --from=builder /workspace/manager .
# Set HOME and CLOUDSDK_CONFIG variables and create directories for them.
ENV HOME=/home/65532
ENV CLOUDSDK_CONFIG=$HOME/gcloud
RUN mkdir -p /home/65532 && chmod -R 777 /home/65532 && \
    mkdir -p /home/65532/gcloud && chmod -R 777 /home/65532/gcloud
# Optionally, run as a non-root user
USER 65532:65532

ENTRYPOINT ["/manager"]

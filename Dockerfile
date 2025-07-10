# Multi-stage build for IBC relayers and web dashboard

# Build Hermes (Rust relayer)
FROM rust:1-bookworm AS hermes-builder

ARG PROTOC_VERSION=28.3

WORKDIR /build

# Install protoc and dependencies
RUN ARCH=$(uname -m) && \
	if [ "$ARCH" = "x86_64" ]; then \
		ARCH=x86_64; \
	elif [ "$ARCH" = "aarch64" ]; then \
		ARCH=aarch_64;\
	else \
		echo "Unsupported architecture: $ARCH"; exit 1; \
	fi && \
	wget https://github.com/protocolbuffers/protobuf/releases/download/v$PROTOC_VERSION/protoc-$PROTOC_VERSION-linux-$ARCH.zip -O /tmp/protoc.zip && \
	unzip /tmp/protoc.zip -d /usr/local && \
	rm -rf /tmp/protoc.zip

RUN apt update && apt install -y clang libssl-dev pkg-config

COPY hermes/ ./
RUN cargo build --release

# Build Go relayer
FROM golang:1.21-alpine AS go-relayer-builder

RUN apk add --update --no-cache curl make git libc-dev bash gcc linux-headers eudev-dev

WORKDIR /build
COPY relayer/ ./

RUN CGO_ENABLED=1 LDFLAGS='-linkmode external -extldflags "-static"' make install

# Build web dashboard backend
FROM golang:1.21-alpine AS backend-builder

RUN apk add --update --no-cache git

WORKDIR /app
COPY go.mod go.sum ./
COPY api/ ./api/

RUN go mod download
RUN go build -o relayer-dashboard ./api/cmd/server

# Build web dashboard frontend
FROM node:20-alpine AS frontend-builder

WORKDIR /app
COPY web/package.json ./
RUN npm install

COPY web/ ./
RUN npm run build

# Final runtime image
FROM ubuntu:22.04

RUN apt-get update && apt-get install -y --no-install-recommends \
	ca-certificates \
	curl \
	jq \
	supervisor \
	nginx \
	&& rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN groupadd -g 1000 relayer && useradd -m -u 1000 -g relayer relayer

# Copy binaries
COPY --from=hermes-builder /build/target/release/hermes /usr/local/bin/hermes
COPY --from=go-relayer-builder /go/bin/rly /usr/local/bin/rly
COPY --from=backend-builder /app/relayer-dashboard /usr/local/bin/relayer-dashboard

# Copy frontend build
COPY --from=frontend-builder /app/dist /var/www/html

# Copy configurations
COPY docker/nginx.conf /etc/nginx/nginx.conf
COPY docker/supervisord.conf /etc/supervisor/conf.d/supervisord.conf

# Create directories
RUN mkdir -p /home/relayer/.hermes /home/relayer/.relayer /var/log/supervisor
RUN chown -R relayer:relayer /home/relayer /var/log/supervisor

# Expose ports
EXPOSE 80 3000 5184 5185

WORKDIR /home/relayer

# Start supervisor
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/conf.d/supervisord.conf"]
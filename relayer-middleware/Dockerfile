# Multi-stage Dockerfile for IBC Relayers

# Build Hermes (latest version)
FROM rust:1-bookworm AS hermes-builder
RUN apt-get update && apt-get install -y git protobuf-compiler
WORKDIR /build
RUN git clone https://github.com/informalsystems/hermes.git
WORKDIR /build/hermes
RUN cargo build --release --bin hermes

# Build Hermes (legacy version for older networks)
FROM rust:1-bookworm AS hermes-legacy-builder
RUN apt-get update && apt-get install -y git protobuf-compiler
WORKDIR /build
RUN git clone https://github.com/informalsystems/hermes.git
WORKDIR /build/hermes
# Checkout a specific older version that works with legacy chains
RUN git checkout v1.4.1
RUN cargo build --release --bin hermes

# Build Go relayer (latest)
FROM golang:1.21-alpine AS go-relayer-builder
RUN apk add --no-cache git make
WORKDIR /build
RUN git clone https://github.com/cosmos/relayer.git
WORKDIR /build/relayer
RUN make install

# Build Go relayer (legacy)
FROM golang:1.21-alpine AS go-relayer-legacy-builder
RUN apk add --no-cache git make
WORKDIR /build
RUN git clone https://github.com/cosmos/relayer.git
WORKDIR /build/relayer
# Checkout a specific older version
RUN git checkout v2.3.1
RUN make install

# Final runtime image
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y \
    ca-certificates \
    curl \
    jq \
    supervisor \
    && rm -rf /var/lib/apt/lists/*

# Copy relayer binaries
COPY --from=hermes-builder /build/hermes/target/release/hermes /usr/local/bin/hermes
COPY --from=hermes-legacy-builder /build/hermes/target/release/hermes /usr/local/bin/hermes-legacy
COPY --from=go-relayer-builder /go/bin/rly /usr/local/bin/rly
COPY --from=go-relayer-legacy-builder /go/bin/rly /usr/local/bin/rly-legacy

# Create directories
RUN mkdir -p /root/.hermes /root/.relayer /etc/supervisor/conf.d /var/log/relayers

# Copy configuration files
COPY supervisord.conf /etc/supervisor/conf.d/supervisord.conf
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Expose metrics ports
EXPOSE 3001 3002 3003 3004

ENTRYPOINT ["/entrypoint.sh"]
CMD ["supervisord", "-c", "/etc/supervisor/conf.d/supervisord.conf"]
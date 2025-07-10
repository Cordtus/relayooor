FROM rust:1-bookworm AS builder

# Install dependencies
RUN apt-get update && apt-get install -y \
    git \
    protobuf-compiler \
    pkg-config \
    libssl-dev \
    && rm -rf /var/lib/apt/lists/*

# Clone and build chainpulse
WORKDIR /build
RUN git clone https://github.com/informalsystems/chainpulse.git
WORKDIR /build/chainpulse

# Build the binary
RUN cargo build --release

# Runtime stage
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    ca-certificates \
    libssl3 \
    && rm -rf /var/lib/apt/lists/*

# Copy binary from builder
COPY --from=builder /build/chainpulse/target/release/chainpulse /usr/local/bin/chainpulse

# Create data directory
RUN mkdir -p /data

# Copy entrypoint script
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Expose metrics port
EXPOSE 3001

VOLUME ["/data", "/config"]

ENTRYPOINT ["/entrypoint.sh"]
CMD ["chainpulse", "--config", "/config/chainpulse.toml"]
# Build stage - use latest Rust for compatibility
FROM rust:1-slim-bullseye as builder

# Install git to clone the repository
RUN apt-get update && apt-get install -y git && rm -rf /var/lib/apt/lists/*

WORKDIR /usr/src

# Clone the forked chainpulse from main branch
RUN git clone -b main https://github.com/cordtus/chainpulse.git
WORKDIR /usr/src/chainpulse

# Show commit info for debugging
RUN git log -1 --oneline

# Build the binary
RUN cargo build --release

# Runtime stage
FROM debian:bullseye-slim

RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy binary from builder
COPY --from=builder /usr/src/chainpulse/target/release/chainpulse /usr/local/bin/chainpulse

# Create data directory
RUN mkdir -p /data

# Copy entrypoint script
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Expose metrics port
EXPOSE 3001

VOLUME ["/data", "/config"]

ENTRYPOINT ["/entrypoint.sh"]
CMD ["chainpulse"]
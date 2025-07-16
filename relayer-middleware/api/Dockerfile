# Build stage
FROM golang:1.21-alpine AS builder

# Install git and ca-certificates for fetching dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy go mod files
COPY go.mod ./

# Copy the source code
COPY . .

# Download dependencies and tidy
RUN go mod tidy && go mod download

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server cmd/server/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/server .

# Expose port
EXPOSE 3000

# Run the binary
CMD ["./server"]
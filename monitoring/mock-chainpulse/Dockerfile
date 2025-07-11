FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git

WORKDIR /build

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN go build -o mock-chainpulse .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /build/mock-chainpulse .

EXPOSE 3001

CMD ["./mock-chainpulse"]
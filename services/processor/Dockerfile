FROM golang:1.23-bullseye AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o processor ./cmd/processor

FROM debian:bullseye-slim

RUN apt-get update && \
    apt-get install -y ffmpeg fonts-dejavu-core && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /app/processor /app/processor
COPY .env /app/.env

ENTRYPOINT ["/app/processor"]

# Stage 1: Build binary
FROM golang:alpine AS builder
WORKDIR /app

# Download dependencies dulu (layer ni di-cache kalau go.mod tak berubah)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code dan build
COPY . .
RUN go build -o server .

# Stage 2: Image kecil untuk run (tanpa Go toolchain)
FROM alpine:latest
WORKDIR /app

# Copy binary je dari stage builder
COPY --from=builder /app/server .

EXPOSE 8080
CMD ["./server"]

# Build Stage
FROM golang:1.24.2-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN go build -o main .

# Run Stage
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .

# Expose the port
EXPOSE 8080

# Run the binary
CMD ["./main"]

# trunk-ignore-all(checkov/CKV_DOCKER_3)
FROM golang:alpine AS builder

WORKDIR /app

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o mshkbot ./cmd/mshkbot

# Final stage
FROM alpine:3.19

WORKDIR /root/

# Copy the pre-built binary
COPY --from=builder /app/mshkbot .

# Add a health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD pgrep mshkbot || exit 1

# Run the bot
CMD ["./mshkbot"]

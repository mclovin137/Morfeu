# Builder stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/app

# Runtime stage
FROM scratch

COPY --from=builder /app/app /app

EXPOSE 8080

ENTRYPOINT ["/app"]

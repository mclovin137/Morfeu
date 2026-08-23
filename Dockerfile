# Builder stage
# Digest do index multi-arch de golang:1.25-alpine (supply chain — PRD 0004 RNF01)
FROM golang:1.27-alpine@sha256:4c9fe60190a2a3350ddc51de80d0224b8a6698d12bdfc999fee45ea9d6c46dbc AS builder

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/morfeu

# Runtime stage
FROM scratch

COPY --from=builder /app/app /app
# main.go roda migrations no startup via migrate.New("file://migrations", ...) —
# sem esta cópia a imagem sobe e morre com "no such file or directory".
COPY --from=builder /app/migrations /migrations

EXPOSE 8080

ENTRYPOINT ["/app"]

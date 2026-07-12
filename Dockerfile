# Builder stage
# Digest do index multi-arch de golang:1.25-alpine (supply chain — PRD 0004 RNF01)
FROM golang:1.25-alpine@sha256:56961d79ea8129efddcc0b8643fd8a5416b4e6228cfd477e3fd61deb2672c587 AS builder

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

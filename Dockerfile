# ---------- Build stage ----------
FROM golang:1.23 AS builder

# Set working directory inside container
WORKDIR /app

# Copy go mod files first (for caching deps)
COPY go.mod ./
RUN go mod download

# Copy the source code
COPY . .

# Build the Go binary (statically linked)
RUN CGO_ENABLED=0 GOOS=linux go build -o server main.go

# ---------- Runtime stage ----------
FROM alpine:3.20

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server .

# Copy static files and templates
COPY static/ ./static/
COPY templates/ ./templates/
COPY config.json.example ./config.json

# Expose HTTP port
EXPOSE 8080

# Run the app
CMD ["./server"]


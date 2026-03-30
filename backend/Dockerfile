FROM golang:alpine AS builder

WORKDIR /app

# Install dependencies first for better dockers layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /cargo-backend ./cmd/server/main.go

# Minimal runner image
FROM alpine:3.18

WORKDIR /
COPY --from=builder /cargo-backend /cargo-backend

EXPOSE 8080

CMD ["/cargo-backend"]

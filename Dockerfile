# ---------- Builder ----------
FROM golang:1.25.3-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o product-service ./cmd/main.go


# ---------- Runtime ----------
FROM alpine:3.18

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/product-service .

EXPOSE 8082

CMD ["./product-service"]

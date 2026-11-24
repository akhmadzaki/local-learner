# ── Stage 1: Build ──────────────────────────────────────────────────────────────
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Cache module downloads in a separate layer
COPY go.mod go.sum* ./
RUN go mod download

# Copy source and compile a statically-linked binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server .

# ── Stage 2: Runtime ────────────────────────────────────────────────────────────
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/server .

EXPOSE ${APP_PORT:-8080}

ENTRYPOINT ["/app/server"]
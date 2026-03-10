# ─────────────────────────────────────────────
# Stage 1: Build
# ─────────────────────────────────────────────
FROM golang:1.26.1-alpine3.23 AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.work go.work.sum ./

COPY apps/    ./apps

RUN go work sync

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /bin/worker ./apps/worker/cmd/main.go


#for debugging
#EXPOSE 8080

#CMD ["sleep", "infinity"]

# ─────────────────────────────────────────────
# Stage 2: Runtime
# ─────────────────────────────────────────────
FROM alpine:3.20 AS runtime

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    ffmpeg

WORKDIR /app

COPY --from=builder /bin/worker ./worker

RUN mkdir -p /mnt/media && \
    chown -R appuser:appgroup /app /mnt/media

USER appuser

EXPOSE 8080

ENTRYPOINT ["./worker"]

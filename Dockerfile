FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/what-pic .

FROM alpine:3.21
RUN apk add --no-cache curl
COPY --from=builder /app/what-pic /app/what-pic
COPY static /app/static
COPY docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh
WORKDIR /app

ENV OLLAMA_HOST=http://localhost:11434
ENV PORT=8080
EXPOSE 8080
HEALTHCHECK --interval=60s --timeout=5s --retries=3 \
  CMD curl -sf http://localhost:${PORT}/ || exit 1
ENTRYPOINT ["/app/docker-entrypoint.sh"]

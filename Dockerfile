# syntax=docker/dockerfile:1

# 1. Build the PWA (React + Vite)
FROM node:20-alpine AS web-build
WORKDIR /app/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# 2. Build the Go backend with the PWA embedded in the binary
FROM golang:1.22-alpine AS backend-build
WORKDIR /app/backend
COPY backend/ ./
RUN go mod tidy
# The compiled PWA goes where //go:embed expects it (see httpapi/server.go).
COPY --from=web-build /app/web/dist/ ./internal/adapters/inbound/httpapi/static/
RUN CGO_ENABLED=0 go build -o /app/hacelo ./cmd/server

# 3. Slim runtime image
FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=backend-build /app/hacelo ./hacelo

ENV PORT=8080
ENV DB_PATH=/data/hacelo.db
VOLUME ["/data"]
EXPOSE 8080

ENTRYPOINT ["./hacelo"]

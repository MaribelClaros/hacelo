#!/usr/bin/env bash
# Builds a single self-contained hacelo binary (backend + embedded PWA) for
# 64-bit ARM Linux, which is what Termux on a modern Android phone runs.
# Output: dist/hacelo-arm64
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
STATIC="$ROOT/backend/internal/adapters/inbound/httpapi/static"

echo "==> Building the PWA"
(cd "$ROOT/web" && npm ci && npm run build)

echo "==> Embedding the PWA into the backend"
find "$STATIC" -mindepth 1 -not -name .gitkeep -exec rm -rf {} +
cp -R "$ROOT/web/dist/." "$STATIC/"

echo "==> Compiling for linux/arm64"
mkdir -p "$ROOT/dist"
(cd "$ROOT/backend" && go mod tidy && GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o "$ROOT/dist/hacelo-arm64" ./cmd/server)

echo "Done: $ROOT/dist/hacelo-arm64"

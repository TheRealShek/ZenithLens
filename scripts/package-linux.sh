#!/usr/bin/env sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
DIST_DIR="$ROOT_DIR/dist"
PACKAGE_DIR="$DIST_DIR/zenithlens-linux-x86_64"

rm -rf "$PACKAGE_DIR"
mkdir -p "$PACKAGE_DIR"

cd "$ROOT_DIR/frontend"
bun install --frozen-lockfile
bun run build

cd "$ROOT_DIR"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o "$PACKAGE_DIR/zenithlens" .

cp README.md "$PACKAGE_DIR/"
cp LICENSE "$PACKAGE_DIR/" 2>/dev/null || true
cp packaging/linux/zenithlens.desktop "$PACKAGE_DIR/"

tar -C "$DIST_DIR" -czf "$DIST_DIR/zenithlens-linux-x86_64.tar.gz" "zenithlens-linux-x86_64"

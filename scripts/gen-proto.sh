#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PROTO_DIR="$ROOT_DIR/proto"
GO_OUT_DIR="$ROOT_DIR/internal/sidecar/proto"
PY_OUT_DIR="$ROOT_DIR/sidecar/proto"

if ! command -v protoc >/dev/null 2>&1; then
  echo "Error: protoc is required but not installed. Please install protobuf-compiler and try again." >&2
  exit 1
fi

if ! command -v protoc-gen-go >/dev/null 2>&1; then
  echo "Error: protoc-gen-go is required but not installed. Run: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest" >&2
  exit 1
fi

if ! command -v protoc-gen-go-grpc >/dev/null 2>&1; then
  echo "Error: protoc-gen-go-grpc is required but not installed. Run: go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest" >&2
  exit 1
fi

mkdir -p "$GO_OUT_DIR" "$PY_OUT_DIR"
rm -rf "$GO_OUT_DIR"/*.pb.go "$PY_OUT_DIR"/*.py

PROTO_FILE="$PROTO_DIR/assistant.proto"

PATH="$(go env GOPATH)/bin:$PATH"
protoc \
  -I "$PROTO_DIR" \
  --go_out="$GO_OUT_DIR" \
  --go_opt=paths=source_relative \
  --go-grpc_out="$GO_OUT_DIR" \
  --go-grpc_opt=paths=source_relative \
  "$PROTO_FILE"

python3 -m grpc_tools.protoc \
  -I "$PROTO_DIR" \
  --python_out="$PY_OUT_DIR" \
  --grpc_python_out="$PY_OUT_DIR" \
  "$PROTO_FILE"

touch "$PY_OUT_DIR/__init__.py"

echo "Generated Go and Python stubs successfully."

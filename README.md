# TerminalVoiceAssist

This repository is set up as a Go + Python voice-assistant workspace following the roadmap in the attached design notes.

## Structure

- `cmd/assistant` — Go entrypoint
- `internal/core` — shared interfaces
- `internal/sidecar` — gRPC client-side Go code
- `internal/config` — configuration utilities
- `internal/audio`, `internal/hotkey`, `internal/skills`, `internal/orchestrator`, `internal/router` — planned modules
- `proto` — protobuf source of truth
- `sidecar` — Python sidecar service
- `scripts` — generation and build helpers

## Quick start

1. Generate protobuf bindings:

```bash
chmod +x scripts/gen-proto.sh
./scripts/gen-proto.sh
```

2. Install Go plugins if missing:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

3. Run the Python sidecar skeleton:

```bash
python3 -m pip install --break-system-packages grpcio grpcio-tools
python3 sidecar/server.py
```

## Notes

This repo currently contains the initial scaffold and contract definitions, and is ready for the next implementation steps in the roadmap.

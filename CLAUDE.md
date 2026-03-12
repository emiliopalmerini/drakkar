# Drakkar

## Build & Test

```bash
# Run all tests
go test ./...

# Run tests for a specific slice
go test ./search/...
go test ./content/...
go test ./browse/...
go test ./memory/...
go test ./openviking/...

# Build
go build -o drakkar .

# Build with Nix
nix build

# Test MCP handshake
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0.1.0"}}}' | ./drakkar
```

## Architecture

Hexagonal architecture with vertical slices. Each slice (search, content, browse, memory) defines a port interface in `port.go` and an MCP handler in `handler.go`. The `openviking/` package implements all ports via HTTP.

Handlers depend only on their port interface — never on `openviking` directly.

## Conventions

- Follow red/green TDD: write failing test first, then implement
- One feature = one package (vertical slice)
- Port interfaces go in `port.go`, MCP handlers in `handler.go`
- Each slice exposes a `Register(s *server.MCPServer, dep Interface)` function
- Tests use fake implementations of the port interface
- MCP SDK: `github.com/mark3labs/mcp-go`

## Config

- `OPENVIKING_URL` env var (default `http://localhost:1933`)

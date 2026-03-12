# Drakkar

OpenViking MCP bridge for Claude Code.

```
Claude Code <--stdio/MCP--> drakkar (Go binary) <--HTTP--> OpenViking server
```

Drakkar is a Go MCP server that wraps [OpenViking](https://github.com/emiliopalmerini/openviking)'s HTTP API, making its tiered knowledge retrieval (L0/L1/L2) available as Claude Code tools.

## Tools

| Tool | Description |
|---|---|
| `search_memories` | Semantic search across knowledge base |
| `context_search` | Context-aware search with auto-managed session |
| `read_content` | Read content at abstract, overview, or full detail |
| `add_memory` | Store new memory via session orchestration |
| `browse` | Navigate knowledge structure (ls/tree/stat) |

## Usage

```bash
# Build
go build -o drakkar .

# Run (reads OPENVIKING_URL, defaults to http://localhost:1933)
export OPENVIKING_URL=http://localhost:1933
./drakkar
```

## Nix

```bash
# Build
nix build

# Use as flake input in your dotfiles
# See PLAN.md for Nix module integration details
```

## Architecture

Hexagonal architecture with vertical slices. Each feature (search, content, browse, memory) is a self-contained package with a port interface and MCP handler. A single `openviking/` adapter implements all ports via HTTP.

See [PLAN.md](PLAN.md) for full implementation details.

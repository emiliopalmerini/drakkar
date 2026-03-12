# Drakkar — OpenViking MCP Bridge

## What

Go MCP server (stdio) that wraps OpenViking's HTTP API, exposing it as Claude Code tools. Replaces the memorizer MCP server (4 Docker containers) with a single binary.

## Architecture

```
Claude Code <--stdio/MCP--> drakkar (Go binary) <--HTTP--> OpenViking server
```

- **Transport**: stdio (no port management, simple Nix integration)
- **Config**: `OPENVIKING_URL` env var (default `http://localhost:8080`)
- **Session**: auto-managed per process for context_search
- **Pattern**: Hexagonal architecture, vertical slices, red/green TDD
- **MCP SDK**: `github.com/mark3labs/mcp-go`

## Project Structure

Organized by vertical slice (one feature = one package), not by technical layer:

```
drakkar/
├── flake.nix
├── go.mod
├── main.go                          # Wire all slices, serve stdio
├── search/                          # Slice: search_memories + context_search
│   ├── port.go                      # Searcher interface (port)
│   ├── handler.go                   # MCP tool handler (adapter: MCP → port)
│   └── handler_test.go              # Tests against fake Searcher
├── content/                         # Slice: read_content (L0/L1/L2)
│   ├── port.go                      # ContentReader interface
│   ├── handler.go                   # MCP tool handler
│   └── handler_test.go
├── memory/                          # Slice: add_memory (session orchestration)
│   ├── port.go                      # MemoryWriter interface
│   ├── handler.go                   # MCP tool handler
│   └── handler_test.go
├── browse/                          # Slice: browse (ls/tree/stat)
│   ├── port.go                      # Browser interface
│   ├── handler.go                   # MCP tool handler
│   └── handler_test.go
└── openviking/                      # Adapter: implements all ports via HTTP
    ├── client.go                    # HTTP client (implements all interfaces)
    └── client_test.go               # Integration tests against httptest server
```

## Hexagonal Boundaries

Each slice defines a port (interface) describing what it needs. The `openviking/client.go` adapter implements all four interfaces via HTTP calls. Handlers depend only on their port interface — never on `openviking` directly.

### Ports

```go
// search/port.go
type Searcher interface {
    Find(ctx context.Context, req FindRequest) (*FindResult, error)
    Search(ctx context.Context, req SearchRequest) (*FindResult, error)
}

// content/port.go
type ContentReader interface {
    ReadAbstract(ctx context.Context, uri string) (string, error)
    ReadOverview(ctx context.Context, uri string) (string, error)
    ReadFull(ctx context.Context, uri string) (string, error)
}

// memory/port.go
type MemoryWriter interface {
    AddMemory(ctx context.Context, content string, role string) (*MemoryResult, error)
}

// browse/port.go
type Browser interface {
    List(ctx context.Context, uri string) (string, error)
    Tree(ctx context.Context, uri string) (string, error)
    Stat(ctx context.Context, uri string) (string, error)
}
```

### main.go Wiring

```go
func main() {
    url := os.Getenv("OPENVIKING_URL")
    if url == "" { url = "http://localhost:8080" }
    client := openviking.NewClient(url)

    s := server.NewMCPServer("drakkar", "0.1.0", server.WithToolCapabilities(true))
    search.Register(s, client)
    content.Register(s, client)
    memory.Register(s, client)
    browse.Register(s, client)

    server.ServeStdio(s)
}
```

## v1 MCP Tools (5)

| Tool | Port Method | OpenViking Endpoint | Purpose |
|---|---|---|---|
| `search_memories` | `Searcher.Find` | `POST /api/v1/search/find` | Semantic search |
| `context_search` | `Searcher.Search` | `POST /api/v1/search/search` | Context-aware search (auto-managed session) |
| `read_content` | `ContentReader.Read{Abstract,Overview,Full}` | `GET /api/v1/content/{abstract,overview,read}` | Read at chosen detail level |
| `add_memory` | `MemoryWriter.AddMemory` | Sessions orchestration | Store new memory |
| `browse` | `Browser.{List,Tree,Stat}` | `GET /api/v1/fs/{ls,tree,stat}` | Navigate knowledge structure |

### Tool Schemas

**search_memories**: `query` (required string), `limit` (optional int, default 10), `score_threshold` (optional float), `target_uri` (optional string)

**context_search**: same as search_memories. Session auto-created on first call, reused for process lifetime.

**read_content**: `uri` (required string), `level` (optional enum: `"abstract"` / `"overview"` / `"full"`, default `"overview"`)

**add_memory**: `content` (required string), `role` (optional string, default `"user"`)

**browse**: `uri` (required string), `mode` (optional enum: `"ls"` / `"tree"` / `"stat"`, default `"ls"`)

## Implementation Order

Red/green TDD, one slice at a time.

### Slice 1: search (simplest — one endpoint, no orchestration)

1. **Red**: Write `search/handler_test.go` — fake `Searcher`, test `search_memories` tool returns results
2. **Green**: Implement `search/port.go` (interface + types), `search/handler.go` (MCP handler)
3. **Red**: Add test for `context_search` tool
4. **Green**: Extend handler with context_search (session auto-management)

### Slice 2: content

5. **Red**: Write `content/handler_test.go` — fake `ContentReader`, test each level routing
6. **Green**: Implement `content/port.go`, `content/handler.go`

### Slice 3: browse

7. **Red**: Write `browse/handler_test.go` — fake `Browser`, test ls/tree/stat routing
8. **Green**: Implement `browse/port.go`, `browse/handler.go`

### Slice 4: memory (most complex — session orchestration)

9. **Red**: Write `memory/handler_test.go` — fake `MemoryWriter`, test add_memory flow
10. **Green**: Implement `memory/port.go`, `memory/handler.go`

### Adapter: openviking client

11. Implement `openviking/client.go` — satisfies all 4 interfaces via HTTP
12. Write `openviking/client_test.go` — integration tests against `httptest.Server`

### Wire & Package

13. Implement `main.go`
14. Add `flake.nix` with `buildGoModule`
15. Test stdio locally

### Nix Integration (dotfiles repo)

16. Create `nix/modules/home/ai/claude/mcp/drakkar.nix`
17. Add to `mcp/default.nix` imports
18. Add `"drakkar"` to `enabledMcpServers` enum
19. Add flake input in `flake.nix`
20. Enable in machine `home.nix`

## Nix Module (dotfiles repo)

```nix
# nix/modules/home/ai/claude/mcp/drakkar.nix
{ lib, config, pkgs, inputs, ... }:
with lib;
let cfg = config.ai.claude.mcp.drakkar;
in {
  options.ai.claude.mcp.drakkar = {
    enable = mkEnableOption "Drakkar MCP server (OpenViking memory bridge)";
    openVikingUrl = mkOption {
      type = types.str;
      default = "http://localhost:8080";
    };
  };
  config = mkIf cfg.enable {
    ai.claude.mcpServers.drakkar = {
      command = "${inputs.drakkar.packages.${pkgs.stdenv.hostPlatform.system}.default}/bin/drakkar";
      env = { OPENVIKING_URL = cfg.openVikingUrl; };
    };
    ai.claude.allowedPermissions = [
      "mcp__drakkar__search_memories"
      "mcp__drakkar__context_search"
      "mcp__drakkar__read_content"
      "mcp__drakkar__add_memory"
      "mcp__drakkar__browse"
    ];
  };
}
```

## Verification

1. `go test ./...` — all slice tests pass against fakes
2. `go test ./openviking/...` — integration tests pass against httptest
3. `echo '{"jsonrpc":"2.0","id":1,"method":"initialize",...}' | ./drakkar` — MCP handshake works
4. With OpenViking running: call each tool via stdio
5. `darwin-rebuild switch` — Claude Code picks up drakkar
6. In Claude Code session: use all 5 tools

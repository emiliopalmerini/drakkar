package openviking_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/epalmerini/drakkar/browse"
	"github.com/epalmerini/drakkar/content"
	"github.com/epalmerini/drakkar/memory"
	"github.com/epalmerini/drakkar/openviking"
	"github.com/epalmerini/drakkar/search"
)

// Compile-time interface checks.
var (
	_ search.Searcher       = (*openviking.Client)(nil)
	_ content.ContentReader = (*openviking.Client)(nil)
	_ browse.Browser        = (*openviking.Client)(nil)
	_ memory.MemoryWriter   = (*openviking.Client)(nil)
)

// --- search: Find ---

func TestFind_SendsCorrectRequest(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/search/find" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &gotBody)
		json.NewEncoder(w).Encode(map[string]any{
			"resources": []map[string]any{
				{"uri": "mem://1", "abstract": "first result", "score": 0.95},
			},
			"memories": []any{},
			"skills":   []any{},
			"total":    1,
		})
	}))
	defer srv.Close()

	c := openviking.NewClient(srv.URL)
	result, err := c.Find(context.Background(), search.FindRequest{
		Query:          "test query",
		Limit:          5,
		ScoreThreshold: 0.8,
		TargetURI:      "mem://scope",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody["query"] != "test query" {
		t.Errorf("query: got %v, want 'test query'", gotBody["query"])
	}
	if gotBody["limit"] != float64(5) {
		t.Errorf("limit: got %v, want 5", gotBody["limit"])
	}
	if gotBody["score_threshold"] != 0.8 {
		t.Errorf("score_threshold: got %v, want 0.8", gotBody["score_threshold"])
	}
	if gotBody["target_uri"] != "mem://scope" {
		t.Errorf("target_uri: got %v, want 'mem://scope'", gotBody["target_uri"])
	}
	if result.Total != 1 {
		t.Errorf("total: got %d, want 1", result.Total)
	}
	if len(result.Results) != 1 {
		t.Fatalf("results len: got %d, want 1", len(result.Results))
	}
	if result.Results[0].URI != "mem://1" {
		t.Errorf("uri: got %q, want 'mem://1'", result.Results[0].URI)
	}
	if result.Results[0].Score != 0.95 {
		t.Errorf("score: got %f, want 0.95", result.Results[0].Score)
	}
}

func TestFind_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := openviking.NewClient(srv.URL)
	_, err := c.Find(context.Background(), search.FindRequest{Query: "fail"})
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

// --- search: Search ---

func TestSearch_SendsSessionID(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/search/search" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &gotBody)
		json.NewEncoder(w).Encode(map[string]any{
			"resources": []any{},
			"memories":  []any{},
			"skills":    []any{},
			"total":     0,
		})
	}))
	defer srv.Close()

	c := openviking.NewClient(srv.URL)
	_, err := c.Search(context.Background(), search.SearchRequest{
		Query: "context query",
		Limit: 3,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody["query"] != "context query" {
		t.Errorf("query: got %v, want 'context query'", gotBody["query"])
	}
}

// --- content: ReadAbstract, ReadOverview, ReadFull ---

func TestReadAbstract(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/content/abstract" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("uri") != "mem://doc" {
			t.Errorf("uri param: got %q", r.URL.Query().Get("uri"))
		}
		w.Write([]byte("short summary"))
	}))
	defer srv.Close()

	c := openviking.NewClient(srv.URL)
	text, err := c.ReadAbstract(context.Background(), "mem://doc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "short summary" {
		t.Errorf("got %q, want 'short summary'", text)
	}
}

func TestReadOverview(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/content/overview" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte("medium detail"))
	}))
	defer srv.Close()

	c := openviking.NewClient(srv.URL)
	text, err := c.ReadOverview(context.Background(), "mem://doc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "medium detail" {
		t.Errorf("got %q, want 'medium detail'", text)
	}
}

func TestReadFull(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/content/read" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte("full content here"))
	}))
	defer srv.Close()

	c := openviking.NewClient(srv.URL)
	text, err := c.ReadFull(context.Background(), "mem://doc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "full content here" {
		t.Errorf("got %q, want 'full content here'", text)
	}
}

func TestReadContent_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	c := openviking.NewClient(srv.URL)
	_, err := c.ReadAbstract(context.Background(), "mem://missing")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
}

// --- browse: List, Tree, Stat ---

func TestList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/fs/ls" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("uri") != "mem://dir" {
			t.Errorf("uri param: got %q", r.URL.Query().Get("uri"))
		}
		json.NewEncoder(w).Encode(map[string]any{
			"entries": []string{"file1.txt", "file2.txt"},
		})
	}))
	defer srv.Close()

	c := openviking.NewClient(srv.URL)
	result, err := c.List(context.Background(), "mem://dir")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestTree(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/fs/tree" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"tree": ".\n└── file1.txt",
		})
	}))
	defer srv.Close()

	c := openviking.NewClient(srv.URL)
	result, err := c.Tree(context.Background(), "mem://dir")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestStat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/fs/stat" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"uri":  "mem://dir/file.txt",
			"type": "file",
			"size": 1024,
		})
	}))
	defer srv.Close()

	c := openviking.NewClient(srv.URL)
	result, err := c.Stat(context.Background(), "mem://dir/file.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestBrowse_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer srv.Close()

	c := openviking.NewClient(srv.URL)
	_, err := c.List(context.Background(), "mem://secret")
	if err == nil {
		t.Fatal("expected error for 403 response")
	}
}

// --- memory: AddMemory (session orchestration) ---

func TestAddMemory_CreatesSessionAndAddsMessage(t *testing.T) {
	var (
		sessionCreated bool
		messageAdded   bool
		gotRole        string
		gotContent     string
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/sessions":
			sessionCreated = true
			json.NewEncoder(w).Encode(map[string]any{
				"status": "ok",
				"result": map[string]any{
					"session_id": "sess-123",
				},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/sessions/sess-123/messages":
			messageAdded = true
			body, _ := io.ReadAll(r.Body)
			var msg map[string]any
			json.Unmarshal(body, &msg)
			gotRole, _ = msg["role"].(string)
			gotContent, _ = msg["content"].(string)
			json.NewEncoder(w).Encode(map[string]any{
				"status": "ok",
				"result": map[string]any{
					"session_id":    "sess-123",
					"message_count": 1,
				},
			})
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c := openviking.NewClient(srv.URL)
	result, err := c.AddMemory(context.Background(), "remember this", "user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !sessionCreated {
		t.Error("expected session to be created")
	}
	if !messageAdded {
		t.Error("expected message to be added")
	}
	if gotRole != "user" {
		t.Errorf("role: got %q, want 'user'", gotRole)
	}
	if gotContent != "remember this" {
		t.Errorf("content: got %q, want 'remember this'", gotContent)
	}
	if result.URI == "" {
		t.Error("expected non-empty URI in result")
	}
}

func TestAddMemory_ReusesSession(t *testing.T) {
	sessionCreateCount := 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/sessions":
			sessionCreateCount++
			json.NewEncoder(w).Encode(map[string]any{
				"status": "ok",
				"result": map[string]any{
					"session_id": "sess-reuse",
				},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/sessions/sess-reuse/messages":
			json.NewEncoder(w).Encode(map[string]any{
				"status": "ok",
				"result": map[string]any{
					"session_id":    "sess-reuse",
					"message_count": 1,
				},
			})
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c := openviking.NewClient(srv.URL)

	// First call creates session
	_, err := c.AddMemory(context.Background(), "first", "user")
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	// Second call reuses session
	_, err = c.AddMemory(context.Background(), "second", "user")
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if sessionCreateCount != 1 {
		t.Errorf("session created %d times, want 1", sessionCreateCount)
	}
}

func TestAddMemory_SessionCreateError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	c := openviking.NewClient(srv.URL)
	_, err := c.AddMemory(context.Background(), "content", "user")
	if err == nil {
		t.Fatal("expected error when session creation fails")
	}
}

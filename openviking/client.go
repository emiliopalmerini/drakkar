package openviking

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/epalmerini/drakkar/browse"
	"github.com/epalmerini/drakkar/content"
	"github.com/epalmerini/drakkar/memory"
	"github.com/epalmerini/drakkar/search"
)

// Compile-time interface checks.
var (
	_ search.Searcher       = (*Client)(nil)
	_ content.ContentReader = (*Client)(nil)
	_ browse.Browser        = (*Client)(nil)
	_ memory.MemoryWriter   = (*Client)(nil)
)

// Client implements all port interfaces via HTTP calls to an OpenViking server.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client

	mu        sync.Mutex
	sessionID string
}

// NewClient creates a new OpenViking HTTP client.
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

// --- Searcher ---

// findResponse maps the OpenViking search response JSON.
type findResponse struct {
	Resources []matchedContext `json:"resources"`
	Memories  []matchedContext `json:"memories"`
	Skills    []matchedContext `json:"skills"`
	Total     int              `json:"total"`
}

type matchedContext struct {
	URI      string  `json:"uri"`
	Abstract string  `json:"abstract"`
	Score    float64 `json:"score"`
}

func (c *Client) Find(ctx context.Context, req search.Request) (*search.FindResult, error) {
	return c.doSearch(ctx, "/api/v1/search/find", req)
}

func (c *Client) Search(ctx context.Context, req search.Request) (*search.FindResult, error) {
	return c.doSearch(ctx, "/api/v1/search/search", req)
}

func (c *Client) doSearch(ctx context.Context, path string, req search.Request) (*search.FindResult, error) {
	body := map[string]any{
		"query":           req.Query,
		"limit":           req.Limit,
		"score_threshold": req.ScoreThreshold,
	}
	if req.TargetURI != "" {
		body["target_uri"] = req.TargetURI
	}

	var resp findResponse
	if err := c.postJSON(ctx, path, body, &resp); err != nil {
		return nil, err
	}
	return toFindResult(&resp), nil
}

func toFindResult(resp *findResponse) *search.FindResult {
	var results []search.Result
	for _, r := range resp.Resources {
		results = append(results, search.Result{
			URI:   r.URI,
			Title: r.Abstract,
			Score: r.Score,
		})
	}
	for _, m := range resp.Memories {
		results = append(results, search.Result{
			URI:   m.URI,
			Title: m.Abstract,
			Score: m.Score,
		})
	}
	return &search.FindResult{
		Results: results,
		Total:   resp.Total,
	}
}

// --- ContentReader ---

func (c *Client) ReadAbstract(ctx context.Context, uri string) (string, error) {
	return c.getString(ctx, "/api/v1/content/abstract", uri)
}

func (c *Client) ReadOverview(ctx context.Context, uri string) (string, error) {
	return c.getString(ctx, "/api/v1/content/overview", uri)
}

func (c *Client) ReadFull(ctx context.Context, uri string) (string, error) {
	return c.getString(ctx, "/api/v1/content/read", uri)
}

// --- Browser ---

func (c *Client) List(ctx context.Context, uri string) (string, error) {
	return c.getString(ctx, "/api/v1/fs/ls", uri)
}

func (c *Client) Tree(ctx context.Context, uri string) (string, error) {
	return c.getString(ctx, "/api/v1/fs/tree", uri)
}

func (c *Client) Stat(ctx context.Context, uri string) (string, error) {
	return c.getString(ctx, "/api/v1/fs/stat", uri)
}

// --- MemoryWriter ---

func (c *Client) AddMemory(ctx context.Context, content string, role string) (*memory.MemoryResult, error) {
	sessionID, err := c.ensureSession(ctx)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	msgBody := map[string]any{
		"role":    role,
		"content": content,
	}

	path := fmt.Sprintf("/api/v1/sessions/%s/messages", sessionID)
	var msgResp struct {
		Status string `json:"status"`
		Result struct {
			SessionID    string `json:"session_id"`
			MessageCount int    `json:"message_count"`
		} `json:"result"`
	}
	if err := c.postJSON(ctx, path, msgBody, &msgResp); err != nil {
		return nil, fmt.Errorf("add message: %w", err)
	}

	return &memory.MemoryResult{
		URI:     fmt.Sprintf("session://%s", sessionID),
		Message: fmt.Sprintf("message added (count: %d)", msgResp.Result.MessageCount),
	}, nil
}

func (c *Client) ensureSession(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.sessionID != "" {
		return c.sessionID, nil
	}

	var resp struct {
		Status string `json:"status"`
		Result struct {
			SessionID string `json:"session_id"`
		} `json:"result"`
	}
	if err := c.postJSON(ctx, "/api/v1/sessions", map[string]any{}, &resp); err != nil {
		return "", err
	}

	c.sessionID = resp.Result.SessionID
	return c.sessionID, nil
}

// --- HTTP helpers ---

func (c *Client) postJSON(ctx context.Context, path string, body any, dst any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	if dst != nil {
		if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

func (c *Client) getString(ctx context.Context, path, uri string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	q := req.URL.Query()
	q.Set("uri", uri)
	req.URL.RawQuery = q.Encode()
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	return string(body), nil
}

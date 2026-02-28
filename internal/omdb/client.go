package omdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"my_mdb/internal/service"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

type Option func(*Client)

func WithBaseURL(u string) Option {
	return func(c *Client) { c.baseURL = u }
}

func New(apiKey string, opts ...Option) (*Client, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, errors.New("omdb: api key is empty")
	}

	c := &Client{
		apiKey:  apiKey,
		baseURL: "https://www.omdbapi.com/",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	if _, err := url.Parse(c.baseURL); err != nil {
		return nil, fmt.Errorf("omdb: invalid base url: %w", err)
	}

	return c, nil
}

var _ service.OMDbClient = (*Client)(nil)

type omdbResponseMeta struct {
	Response string `json:"Response"`
	Error    string `json:"Error"`
	Poster   string `json:"Poster"`
}

func (c *Client) FetchMovie(ctx context.Context, imdbID string) ([]byte, *string, error) {
	imdbID = strings.TrimSpace(imdbID)
	if imdbID == "" {
		return nil, nil, errors.New("omdb: imdbID is empty")
	}
	if !strings.HasPrefix(imdbID, "tt") || len(imdbID) < 3 {
		return nil, nil, fmt.Errorf("omdb: invalid imdbID format: %q", imdbID)
	}

	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, nil, fmt.Errorf("omdb: parse base url: %w", err)
	}

	q := u.Query()
	q.Set("apikey", c.apiKey)
	q.Set("i", imdbID)
	q.Set("plot", "full")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("omdb: build request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("omdb: do request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, nil, fmt.Errorf("omdb: read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet := string(body)
		if len(snippet) > 300 {
			snippet = snippet[:300] + "..."
		}
		return nil, nil, fmt.Errorf("omdb: http %d: %s", resp.StatusCode, snippet)
	}

	var meta omdbResponseMeta
	if err := json.Unmarshal(body, &meta); err != nil {
		return body, nil, fmt.Errorf("omdb: unmarshal meta: %w", err)
	}

	if strings.EqualFold(meta.Response, "False") {
		msg := strings.TrimSpace(meta.Error)
		if msg == "" {
			msg = "unknown error"
		}
		return nil, nil, fmt.Errorf("omdb: %s", msg)
	}

	var poster *string
	if p := strings.TrimSpace(meta.Poster); p != "" && !strings.EqualFold(p, "N/A") {
		poster = &p
	}

	return body, poster, nil
}

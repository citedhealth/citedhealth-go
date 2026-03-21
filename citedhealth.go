// Package citedhealth provides a Go client for the CITED Health API.
//
// CITED Health is an evidence-based supplement database across 6 sites covering
// 188 ingredients, 84 conditions, 323 evidence links, and 6,197 PubMed papers.
// This client requires no authentication and has zero external dependencies.
//
// Usage:
//
//	client := citedhealth.New()
//	ingredients, err := client.SearchIngredients(ctx, "biotin", "")
//	evidence, err := client.GetEvidence(ctx, "biotin", "hair-loss")
package citedhealth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// DefaultBaseURL is the default base URL for the CITED Health API.
const DefaultBaseURL = "https://haircited.com"

// DefaultTimeout is the default HTTP timeout.
const DefaultTimeout = 30 * time.Second

// Client is a CITED Health API client.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// Option configures the Client.
type Option func(*Client)

// WithBaseURL sets a custom base URL.
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = strings.TrimRight(url, "/")
	}
}

// WithTimeout sets a custom HTTP timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = d
	}
}

// New creates a new CITED Health API client.
func New(opts ...Option) *Client {
	c := &Client{
		baseURL:    DefaultBaseURL,
		httpClient: &http.Client{Timeout: DefaultTimeout},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Client) doRequest(ctx context.Context, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("citedhealth: create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("citedhealth: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("citedhealth: read body: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return body, nil
	case http.StatusNotFound:
		return nil, &NotFoundError{Resource: "resource", Identifier: path}
	case http.StatusTooManyRequests:
		retryAfter := 0
		if v := resp.Header.Get("Retry-After"); v != "" {
			retryAfter, _ = strconv.Atoi(v)
		}
		return nil, &RateLimitError{RetryAfter: retryAfter}
	default:
		return nil, &CitedHealthError{
			Message: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)),
		}
	}
}

// SearchIngredients searches ingredients by name or category.
func (c *Client) SearchIngredients(ctx context.Context, query, category string) ([]Ingredient, error) {
	params := url.Values{}
	if query != "" {
		params.Set("q", query)
	}
	if category != "" {
		params.Set("category", category)
	}

	path := "/api/ingredients/"
	if encoded := params.Encode(); encoded != "" {
		path += "?" + encoded
	}

	body, err := c.doRequest(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("citedhealth: search ingredients: %w", err)
	}

	var page PaginatedResponse
	if err := json.Unmarshal(body, &page); err != nil {
		return nil, fmt.Errorf("citedhealth: decode response: %w", err)
	}

	var ingredients []Ingredient
	if err := json.Unmarshal(page.Results, &ingredients); err != nil {
		return nil, fmt.Errorf("citedhealth: decode ingredients: %w", err)
	}
	return ingredients, nil
}

// GetIngredient returns a single ingredient by slug.
func (c *Client) GetIngredient(ctx context.Context, slug string) (*Ingredient, error) {
	body, err := c.doRequest(ctx, "/api/ingredients/"+url.PathEscape(slug)+"/")
	if err != nil {
		return nil, fmt.Errorf("citedhealth: get ingredient: %w", err)
	}

	var ingredient Ingredient
	if err := json.Unmarshal(body, &ingredient); err != nil {
		return nil, fmt.Errorf("citedhealth: decode ingredient: %w", err)
	}
	return &ingredient, nil
}

// GetEvidence returns evidence for an ingredient-condition pair.
// It queries the paginated evidence endpoint and returns the first result.
func (c *Client) GetEvidence(ctx context.Context, ingredientSlug, conditionSlug string) (*EvidenceLink, error) {
	params := url.Values{}
	params.Set("ingredient", ingredientSlug)
	params.Set("condition", conditionSlug)

	path := "/api/evidence/?" + params.Encode()

	body, err := c.doRequest(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("citedhealth: get evidence: %w", err)
	}

	var page PaginatedResponse
	if err := json.Unmarshal(body, &page); err != nil {
		return nil, fmt.Errorf("citedhealth: decode response: %w", err)
	}

	var results []EvidenceLink
	if err := json.Unmarshal(page.Results, &results); err != nil {
		return nil, fmt.Errorf("citedhealth: decode evidence: %w", err)
	}
	if len(results) == 0 {
		return nil, &NotFoundError{
			Resource:   "evidence",
			Identifier: ingredientSlug + " \u00d7 " + conditionSlug,
		}
	}
	return &results[0], nil
}

// GetEvidenceByID returns a single evidence link by numeric ID.
func (c *Client) GetEvidenceByID(ctx context.Context, id int) (*EvidenceLink, error) {
	path := fmt.Sprintf("/api/evidence/%d/", id)

	body, err := c.doRequest(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("citedhealth: get evidence by id: %w", err)
	}

	var evidence EvidenceLink
	if err := json.Unmarshal(body, &evidence); err != nil {
		return nil, fmt.Errorf("citedhealth: decode evidence: %w", err)
	}
	return &evidence, nil
}

// SearchPapers searches PubMed papers by keyword or publication year.
func (c *Client) SearchPapers(ctx context.Context, query string, year *int) ([]Paper, error) {
	params := url.Values{}
	if query != "" {
		params.Set("q", query)
	}
	if year != nil {
		params.Set("year", strconv.Itoa(*year))
	}

	path := "/api/papers/"
	if encoded := params.Encode(); encoded != "" {
		path += "?" + encoded
	}

	body, err := c.doRequest(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("citedhealth: search papers: %w", err)
	}

	var page PaginatedResponse
	if err := json.Unmarshal(body, &page); err != nil {
		return nil, fmt.Errorf("citedhealth: decode response: %w", err)
	}

	var papers []Paper
	if err := json.Unmarshal(page.Results, &papers); err != nil {
		return nil, fmt.Errorf("citedhealth: decode papers: %w", err)
	}
	return papers, nil
}

// GetPaper returns a single paper by PubMed ID.
func (c *Client) GetPaper(ctx context.Context, pmid string) (*Paper, error) {
	body, err := c.doRequest(ctx, "/api/papers/"+url.PathEscape(pmid)+"/")
	if err != nil {
		return nil, fmt.Errorf("citedhealth: get paper: %w", err)
	}

	var paper Paper
	if err := json.Unmarshal(body, &paper); err != nil {
		return nil, fmt.Errorf("citedhealth: decode paper: %w", err)
	}
	return &paper, nil
}

// ListConditions returns health conditions, optionally filtered by featured status.
func (c *Client) ListConditions(ctx context.Context, isFeatured *bool) ([]Condition, error) {
	params := url.Values{}
	if isFeatured != nil {
		params.Set("is_featured", strconv.FormatBool(*isFeatured))
	}

	path := "/api/conditions/"
	if encoded := params.Encode(); encoded != "" {
		path += "?" + encoded
	}

	body, err := c.doRequest(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("citedhealth: list conditions: %w", err)
	}

	var page PaginatedResponse
	if err := json.Unmarshal(body, &page); err != nil {
		return nil, fmt.Errorf("citedhealth: decode response: %w", err)
	}

	var conditions []Condition
	if err := json.Unmarshal(page.Results, &conditions); err != nil {
		return nil, fmt.Errorf("citedhealth: decode conditions: %w", err)
	}
	return conditions, nil
}

// GetCondition returns a single condition by slug.
func (c *Client) GetCondition(ctx context.Context, slug string) (*Condition, error) {
	body, err := c.doRequest(ctx, "/api/conditions/"+url.PathEscape(slug)+"/")
	if err != nil {
		return nil, fmt.Errorf("citedhealth: get condition: %w", err)
	}

	var condition Condition
	if err := json.Unmarshal(body, &condition); err != nil {
		return nil, fmt.Errorf("citedhealth: decode condition: %w", err)
	}
	return &condition, nil
}

// ListGlossary returns glossary terms, optionally filtered by category.
func (c *Client) ListGlossary(ctx context.Context, category string) ([]GlossaryTerm, error) {
	params := url.Values{}
	if category != "" {
		params.Set("category", category)
	}

	path := "/api/glossary/"
	if encoded := params.Encode(); encoded != "" {
		path += "?" + encoded
	}

	body, err := c.doRequest(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("citedhealth: list glossary: %w", err)
	}

	var page PaginatedResponse
	if err := json.Unmarshal(body, &page); err != nil {
		return nil, fmt.Errorf("citedhealth: decode response: %w", err)
	}

	var terms []GlossaryTerm
	if err := json.Unmarshal(page.Results, &terms); err != nil {
		return nil, fmt.Errorf("citedhealth: decode glossary: %w", err)
	}
	return terms, nil
}

// GetGlossaryTerm returns a single glossary term by slug.
func (c *Client) GetGlossaryTerm(ctx context.Context, slug string) (*GlossaryTerm, error) {
	body, err := c.doRequest(ctx, "/api/glossary/"+url.PathEscape(slug)+"/")
	if err != nil {
		return nil, fmt.Errorf("citedhealth: get glossary term: %w", err)
	}

	var term GlossaryTerm
	if err := json.Unmarshal(body, &term); err != nil {
		return nil, fmt.Errorf("citedhealth: decode glossary term: %w", err)
	}
	return &term, nil
}

// ListGuides returns educational guides, optionally filtered by category.
func (c *Client) ListGuides(ctx context.Context, category string) ([]Guide, error) {
	params := url.Values{}
	if category != "" {
		params.Set("category", category)
	}

	path := "/api/guides/"
	if encoded := params.Encode(); encoded != "" {
		path += "?" + encoded
	}

	body, err := c.doRequest(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("citedhealth: list guides: %w", err)
	}

	var page PaginatedResponse
	if err := json.Unmarshal(body, &page); err != nil {
		return nil, fmt.Errorf("citedhealth: decode response: %w", err)
	}

	var guides []Guide
	if err := json.Unmarshal(page.Results, &guides); err != nil {
		return nil, fmt.Errorf("citedhealth: decode guides: %w", err)
	}
	return guides, nil
}

// GetGuide returns a single guide by slug.
func (c *Client) GetGuide(ctx context.Context, slug string) (*Guide, error) {
	body, err := c.doRequest(ctx, "/api/guides/"+url.PathEscape(slug)+"/")
	if err != nil {
		return nil, fmt.Errorf("citedhealth: get guide: %w", err)
	}

	var guide Guide
	if err := json.Unmarshal(body, &guide); err != nil {
		return nil, fmt.Errorf("citedhealth: decode guide: %w", err)
	}
	return &guide, nil
}

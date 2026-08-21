package n8n

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// NotFoundError indicates the requested resource does not exist.
type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s %q not found", e.Resource, e.ID)
}

// APIError is a non-success response from the n8n Public API.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("n8n API error: HTTP %d", e.StatusCode)
	}
	return fmt.Sprintf("n8n API error: HTTP %d: %s", e.StatusCode, e.Body)
}

// Workflow is the Public API workflow document used for create/update/read.
type Workflow struct {
	ID           string          `json:"id,omitempty"`
	Name         string          `json:"name"`
	Active       bool            `json:"active,omitempty"`
	Nodes        json.RawMessage `json:"nodes"`
	Connections  json.RawMessage `json:"connections"`
	Settings     json.RawMessage `json:"settings"`
	VersionID    string          `json:"versionId,omitempty"`
	CreatedAt    string          `json:"createdAt,omitempty"`
	UpdatedAt    string          `json:"updatedAt,omitempty"`
	IsArchived   bool            `json:"isArchived,omitempty"`
	TriggerCount int             `json:"triggerCount,omitempty"`
}

// WorkflowCreate is the request body for POST /workflows.
type WorkflowCreate struct {
	Name        string          `json:"name"`
	Nodes       json.RawMessage `json:"nodes"`
	Connections json.RawMessage `json:"connections"`
	Settings    json.RawMessage `json:"settings"`
}

// WorkflowUpdate is the request body for PUT /workflows/{id}.
type WorkflowUpdate struct {
	Name        string          `json:"name"`
	Nodes       json.RawMessage `json:"nodes"`
	Connections json.RawMessage `json:"connections"`
	Settings    json.RawMessage `json:"settings"`
}

func (c *Client) url(parts ...string) string {
	base := strings.TrimSuffix(c.Endpoint, "/")
	for _, p := range parts {
		base += "/" + strings.Trim(p, "/")
	}
	return base
}

func (c *Client) doJSON(ctx context.Context, method, url string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		reader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return &NotFoundError{Resource: "workflow", ID: url}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{StatusCode: resp.StatusCode, Body: strings.TrimSpace(string(respBody))}
	}

	if out == nil || len(respBody) == 0 {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

// CreateWorkflow creates a workflow via POST /workflows.
func (c *Client) CreateWorkflow(ctx context.Context, in WorkflowCreate) (*Workflow, error) {
	if len(in.Settings) == 0 {
		in.Settings = json.RawMessage(`{}`)
	}
	var out Workflow
	if err := c.doJSON(ctx, http.MethodPost, c.url("workflows"), in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetWorkflow retrieves a workflow via GET /workflows/{id}.
func (c *Client) GetWorkflow(ctx context.Context, id string) (*Workflow, error) {
	var out Workflow
	if err := c.doJSON(ctx, http.MethodGet, c.url("workflows", id), nil, &out); err != nil {
		if nf, ok := err.(*NotFoundError); ok {
			nf.ID = id
			return nil, nf
		}
		return nil, err
	}
	return &out, nil
}

// UpdateWorkflow updates a workflow via PUT /workflows/{id}.
func (c *Client) UpdateWorkflow(ctx context.Context, id string, in WorkflowUpdate) (*Workflow, error) {
	if len(in.Settings) == 0 {
		in.Settings = json.RawMessage(`{}`)
	}
	var out Workflow
	if err := c.doJSON(ctx, http.MethodPut, c.url("workflows", id), in, &out); err != nil {
		if nf, ok := err.(*NotFoundError); ok {
			nf.ID = id
			return nil, nf
		}
		return nil, err
	}
	return &out, nil
}

// DeleteWorkflow deletes a workflow via DELETE /workflows/{id}.
func (c *Client) DeleteWorkflow(ctx context.Context, id string) error {
	err := c.doJSON(ctx, http.MethodDelete, c.url("workflows", id), nil, nil)
	if nf, ok := err.(*NotFoundError); ok {
		nf.ID = id
		return nf
	}
	return err
}

// ActivateWorkflow activates a workflow via POST /workflows/{id}/activate.
func (c *Client) ActivateWorkflow(ctx context.Context, id string) (*Workflow, error) {
	var out Workflow
	if err := c.doJSON(ctx, http.MethodPost, c.url("workflows", id, "activate"), nil, &out); err != nil {
		if nf, ok := err.(*NotFoundError); ok {
			nf.ID = id
			return nil, nf
		}
		return nil, err
	}
	return &out, nil
}

// DeactivateWorkflow deactivates a workflow via POST /workflows/{id}/deactivate.
func (c *Client) DeactivateWorkflow(ctx context.Context, id string) (*Workflow, error) {
	var out Workflow
	if err := c.doJSON(ctx, http.MethodPost, c.url("workflows", id, "deactivate"), nil, &out); err != nil {
		if nf, ok := err.(*NotFoundError); ok {
			nf.ID = id
			return nil, nf
		}
		return nil, err
	}
	return &out, nil
}

package n8n

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestClientRespectsMaxConcurrent(t *testing.T) {
	var peak atomic.Int32
	var active atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cur := active.Add(1)
		for {
			old := peak.Load()
			if cur <= old {
				break
			}
			if peak.CompareAndSwap(old, cur) {
				break
			}
		}
		time.Sleep(40 * time.Millisecond)
		active.Add(-1)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	client, err := New(srv.URL, "secret", &Options{
		MaxConcurrent: 3,
		RPS:           500,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	const workers = 25
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, client.Endpoint, nil)
			if err != nil {
				t.Errorf("NewRequest: %v", err)
				return
			}
			resp, err := client.HTTP.Do(req)
			if err != nil {
				t.Errorf("Do: %v", err)
				return
			}
			_ = resp.Body.Close()
		}()
	}
	wg.Wait()

	if got := peak.Load(); got > 3 {
		t.Fatalf("expected peak concurrent <= 3, got %d", got)
	}
}

func TestClientRateLimitsSerialRequests(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	client, err := New(srv.URL, "secret", &Options{
		MaxConcurrent: 10,
		RPS:           2,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	start := time.Now()
	for i := 0; i < 6; i++ {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, client.Endpoint, nil)
		if err != nil {
			t.Fatalf("NewRequest: %v", err)
		}
		resp, err := client.HTTP.Do(req)
		if err != nil {
			t.Fatalf("Do: %v", err)
		}
		_ = resp.Body.Close()
	}
	elapsed := time.Since(start)
	if elapsed < 1200*time.Millisecond {
		t.Fatalf("expected serial requests to take at least ~1.2s with RPS=2, got %v", elapsed)
	}
}

func TestClientRefusesCrossHostRedirect(t *testing.T) {
	var targetHit atomic.Bool
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		targetHit.Store(true)
		if r.Header.Get("X-N8N-API-KEY") != "" {
			t.Error("API key forwarded to cross-host redirect target")
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(target.Close)

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL+"/landed", http.StatusFound)
	}))
	t.Cleanup(origin.Close)

	client, err := New(origin.URL, "leak-me-key", &Options{RPS: 100})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, client.Endpoint, nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	resp, err := client.HTTP.Do(req)
	if err == nil {
		_ = resp.Body.Close()
		t.Fatal("expected error refusing cross-host redirect")
	}
	if targetHit.Load() {
		t.Fatal("cross-host redirect target was contacted")
	}
}

func TestWorkflowCRUD(t *testing.T) {
	var activated, deactivated bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-N8N-API-KEY") != "secret" {
			t.Errorf("missing API key header")
		}
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/workflows":
			var in WorkflowWrite
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				t.Errorf("decode create: %v", err)
			}
			_ = json.NewEncoder(w).Encode(Workflow{
				ID:          "wf-1",
				Name:        in.Name,
				Nodes:       in.Nodes,
				Connections: in.Connections,
				Settings:    in.Settings,
				VersionID:   "v1",
				CreatedAt:   "2026-01-01T00:00:00.000Z",
				UpdatedAt:   "2026-01-01T00:00:00.000Z",
			})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/workflows/wf-1":
			_ = json.NewEncoder(w).Encode(Workflow{
				ID:           "wf-1",
				Name:         "demo",
				Active:       activated && !deactivated,
				Nodes:        json.RawMessage(`[{"id":"1"}]`),
				Connections:  json.RawMessage(`{}`),
				Settings:     json.RawMessage(`{}`),
				VersionID:    "v2",
				CreatedAt:    "2026-01-01T00:00:00.000Z",
				UpdatedAt:    "2026-01-02T00:00:00.000Z",
				TriggerCount: 0,
			})
		case r.Method == http.MethodPut && r.URL.Path == "/api/v1/workflows/wf-1":
			var in WorkflowWrite
			_ = json.NewDecoder(r.Body).Decode(&in)
			_ = json.NewEncoder(w).Encode(Workflow{
				ID:          "wf-1",
				Name:        in.Name,
				Nodes:       in.Nodes,
				Connections: in.Connections,
				Settings:    in.Settings,
				VersionID:   "v3",
			})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/workflows/wf-1/activate":
			activated = true
			deactivated = false
			_ = json.NewEncoder(w).Encode(Workflow{ID: "wf-1", Active: true, Name: "demo"})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/workflows/wf-1/deactivate":
			deactivated = true
			activated = false
			_ = json.NewEncoder(w).Encode(Workflow{ID: "wf-1", Active: false, Name: "demo"})
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/workflows/wf-1":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/workflows/missing":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"not found"}`))
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	client, err := New(srv.URL, "secret", &Options{RPS: 100})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	created, err := client.CreateWorkflow(context.Background(), WorkflowWrite{
		Name:        "demo",
		Nodes:       json.RawMessage(`[{"id":"1"}]`),
		Connections: json.RawMessage(`{}`),
		Settings:    json.RawMessage(`{}`),
	})
	if err != nil {
		t.Fatalf("CreateWorkflow: %v", err)
	}
	if created.ID != "wf-1" {
		t.Fatalf("created id=%q", created.ID)
	}

	got, err := client.GetWorkflow(context.Background(), "wf-1")
	if err != nil {
		t.Fatalf("GetWorkflow: %v", err)
	}
	if got.Name != "demo" {
		t.Fatalf("got name=%q", got.Name)
	}

	_, err = client.UpdateWorkflow(context.Background(), "wf-1", WorkflowWrite{
		Name:        "demo2",
		Nodes:       json.RawMessage(`[{"id":"1"}]`),
		Connections: json.RawMessage(`{}`),
		Settings:    json.RawMessage(`{}`),
	})
	if err != nil {
		t.Fatalf("UpdateWorkflow: %v", err)
	}

	act, err := client.ActivateWorkflow(context.Background(), "wf-1")
	if err != nil || !act.Active {
		t.Fatalf("ActivateWorkflow: err=%v active=%v", err, act != nil && act.Active)
	}
	deact, err := client.DeactivateWorkflow(context.Background(), "wf-1")
	if err != nil || deact.Active {
		t.Fatalf("DeactivateWorkflow: err=%v active=%v", err, deact != nil && deact.Active)
	}

	if err := client.DeleteWorkflow(context.Background(), "wf-1"); err != nil {
		t.Fatalf("DeleteWorkflow: %v", err)
	}

	_, err = client.GetWorkflow(context.Background(), "missing")
	if _, ok := err.(*NotFoundError); !ok {
		t.Fatalf("expected NotFoundError, got %T %v", err, err)
	}
}

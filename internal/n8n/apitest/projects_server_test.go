package apitest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yu-iskw/terraform-provider-n8n/internal/n8n/models"
)

func TestProjectsServerCRUD(t *testing.T) {
	srv := httptest.NewServer(NewProjectsServer("secret").Handler())
	t.Cleanup(srv.Close)

	req := func(method, path string, body any) *http.Response {
		t.Helper()
		var r *http.Request
		var err error
		if body != nil {
			b, mErr := json.Marshal(body)
			if mErr != nil {
				t.Fatal(mErr)
			}
			r, err = http.NewRequest(method, srv.URL+path, bytes.NewReader(b))
		} else {
			r, err = http.NewRequest(method, srv.URL+path, nil)
		}
		if err != nil {
			t.Fatal(err)
		}
		r.Header.Set("X-N8N-API-KEY", "secret")
		r.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(r)
		if err != nil {
			t.Fatal(err)
		}
		return resp
	}

	resp := req(http.MethodPost, "/api/v1/projects", models.ProjectWrite{Name: "alpha"})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create status=%d", resp.StatusCode)
	}
	var created models.Project
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if created.ID == "" || created.Type != models.ProjectTypeTeam || created.Name != "alpha" {
		t.Fatalf("created=%+v", created)
	}

	resp = req(http.MethodPut, "/api/v1/projects/"+created.ID, models.ProjectWrite{Name: "beta"})
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("update status=%d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	resp = req(http.MethodGet, "/api/v1/projects?limit=250", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list status=%d", resp.StatusCode)
	}
	var list models.ProjectList
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	found := false
	for _, p := range list.Data {
		if p.ID == created.ID && p.Name == "beta" {
			found = true
		}
	}
	if !found {
		t.Fatalf("updated project missing from list: %+v", list.Data)
	}

	resp = req(http.MethodDelete, "/api/v1/projects/"+created.ID, nil)
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete status=%d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	resp = req(http.MethodDelete, "/api/v1/projects/"+created.ID, nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("second delete status=%d", resp.StatusCode)
	}
	_ = resp.Body.Close()
}

func TestProjectsServerUnauthorized(t *testing.T) {
	srv := httptest.NewServer(NewProjectsServer("secret").Handler())
	t.Cleanup(srv.Close)
	resp, err := http.Get(srv.URL + "/api/v1/projects")
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status=%d", resp.StatusCode)
	}
}

func TestProjectsServerReady(t *testing.T) {
	srv := httptest.NewServer(NewProjectsServer("secret").Handler())
	t.Cleanup(srv.Close)
	resp, err := http.Get(srv.URL + "/healthz/readiness")
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d", resp.StatusCode)
	}
}

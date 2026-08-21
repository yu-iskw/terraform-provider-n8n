package folders

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/yu-iskw/terraform-provider-n8n/internal/n8n"
)

func testdata(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func testClient(t *testing.T, handler http.HandlerFunc) *n8n.Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	client, err := n8n.New(srv.URL, "secret", &n8n.Options{RPS: 100})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return client
}

func writeFixture(t *testing.T, w http.ResponseWriter, status int, name string) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(testdata(t, name)); err != nil {
		t.Errorf("write fixture: %v", err)
	}
}

func decodeBody(t *testing.T, r *http.Request, out any) {
	t.Helper()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if err := json.Unmarshal(b, out); err != nil {
		t.Fatalf("decode body %s: %v", b, err)
	}
}

func assertAPIKey(t *testing.T, r *http.Request) {
	t.Helper()
	if r.Header.Get("X-N8N-API-KEY") != "secret" {
		t.Errorf("missing API key header")
	}
}

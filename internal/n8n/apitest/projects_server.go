package apitest

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/yu-iskw/terraform-provider-n8n/internal/n8n/models"
)

const DefaultAPIKey = "tf-acc-fixture-key"

// ProjectsServer is an in-memory n8n Public API stand-in for GET/POST/PUT/DELETE /api/v1/projects.
// Community n8n Docker returns 403 for feat:projectRole:admin; this fixture lets acceptance tests run.
type ProjectsServer struct {
	APIKey string

	mu       sync.Mutex
	projects map[string]models.Project
}

// NewProjectsServer returns a server with one personal project, matching a typical n8n instance.
func NewProjectsServer(apiKey string) *ProjectsServer {
	if apiKey == "" {
		apiKey = DefaultAPIKey
	}
	s := &ProjectsServer{
		APIKey:   apiKey,
		projects: make(map[string]models.Project),
	}
	s.projects["personal-1"] = models.Project{
		ID:   "personal-1",
		Name: "Personal",
		Type: models.ProjectTypePersonal,
	}
	return s
}

// Handler serves readiness plus /api/v1/projects.
func (s *ProjectsServer) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz/readiness", s.handleReady)
	mux.HandleFunc("GET /api/v1/projects", s.withAuth(s.handleList))
	mux.HandleFunc("POST /api/v1/projects", s.withAuth(s.handleCreate))
	mux.HandleFunc("PUT /api/v1/projects/{id}", s.withAuth(s.handleUpdate))
	mux.HandleFunc("DELETE /api/v1/projects/{id}", s.withAuth(s.handleDelete))
	return mux
}

func (s *ProjectsServer) handleReady(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *ProjectsServer) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-N8N-API-KEY") != s.APIKey {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "'X-N8N-API-KEY' header required"})
			return
		}
		next(w, r)
	}
}

func (s *ProjectsServer) handleList(w http.ResponseWriter, r *http.Request) {
	limit := 250
	if raw := r.URL.Query().Get("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err == nil && n > 0 {
			limit = n
		}
	}
	offset := 0
	if raw := r.URL.Query().Get("cursor"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err == nil && n > 0 {
			offset = n
		}
	}

	s.mu.Lock()
	all := make([]models.Project, 0, len(s.projects))
	for _, p := range s.projects {
		all = append(all, p)
	}
	s.mu.Unlock()
	sort.Slice(all, func(i, j int) bool { return all[i].ID < all[j].ID })

	if offset > len(all) {
		offset = len(all)
	}
	end := offset + limit
	if end > len(all) {
		end = len(all)
	}
	page := all[offset:end]
	out := models.ProjectList{Data: page}
	if end < len(all) {
		next := strconv.Itoa(end)
		out.NextCursor = &next
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *ProjectsServer) handleCreate(w http.ResponseWriter, r *http.Request) {
	var in models.ProjectWrite
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid json"})
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "name is required"})
		return
	}
	p := models.Project{
		ID:   newProjectID(),
		Name: in.Name,
		Type: models.ProjectTypeTeam,
		Role: "project:admin",
	}
	s.mu.Lock()
	s.projects[p.ID] = p
	s.mu.Unlock()
	writeJSON(w, http.StatusCreated, p)
}

func (s *ProjectsServer) handleUpdate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in models.ProjectWrite
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid json"})
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "name is required"})
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.projects[id]
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"message": "project not found"})
		return
	}
	p.Name = in.Name
	s.projects[id] = p
	w.WriteHeader(http.StatusNoContent)
}

func (s *ProjectsServer) handleDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.projects[id]; !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"message": "project not found"})
		return
	}
	delete(s.projects, id)
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func newProjectID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "proj-fallback"
	}
	return hex.EncodeToString(b)
}

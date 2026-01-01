package server

import (
	"encoding/json"
	"net/http"

	"github.com/DaikonSushi/bot-platform/internal/pluginmgr"
)

// AdminServer provides HTTP API for plugin management
type AdminServer struct {
	pm   *pluginmgr.PluginManager
	addr string
}

// NewAdminServer creates a new admin server
func NewAdminServer(addr string, pm *pluginmgr.PluginManager) *AdminServer {
	return &AdminServer{
		addr: addr,
		pm:   pm,
	}
}

// Start starts the admin HTTP server
func (s *AdminServer) Start() error {
	mux := http.NewServeMux()

	// Plugin management endpoints
	mux.HandleFunc("/api/plugins", s.handlePlugins)
	mux.HandleFunc("/api/plugins/install", s.handleInstall)
	mux.HandleFunc("/api/plugins/start", s.handleStart)
	mux.HandleFunc("/api/plugins/stop", s.handleStop)
	mux.HandleFunc("/api/plugins/uninstall", s.handleUninstall)
	mux.HandleFunc("/api/health", s.handleHealth)

	return http.ListenAndServe(s.addr, mux)
}

// handlePlugins returns list of all plugins
func (s *AdminServer) handlePlugins(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	plugins := s.pm.ListPlugins()

	type pluginResponse struct {
		Name        string   `json:"name"`
		Version     string   `json:"version"`
		Description string   `json:"description"`
		Author      string   `json:"author"`
		Commands    []string `json:"commands"`
		Status      string   `json:"status"`
		RepoURL     string   `json:"repo_url,omitempty"`
	}

	result := make([]pluginResponse, 0)
	for _, p := range plugins {
		result = append(result, pluginResponse{
			Name:        p.Info.Name,
			Version:     p.Info.Version,
			Description: p.Info.Description,
			Author:      p.Info.Author,
			Commands:    p.Info.Commands,
			Status:      p.Status,
			RepoURL:     p.Info.RepoURL,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    0,
		"message": "success",
		"data":    result,
	})
}

// handleInstall installs a plugin from GitHub
func (s *AdminServer) handleInstall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		RepoURL   string `json:"repo_url"`
		AutoStart bool   `json:"auto_start"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.RepoURL == "" {
		jsonError(w, "repo_url is required", http.StatusBadRequest)
		return
	}

	meta, err := s.pm.InstallFromGitHub(r.Context(), req.RepoURL)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Auto-start plugin if requested
	if req.AutoStart {
		if err := s.pm.StartPlugin(r.Context(), meta.Name); err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"code":    0,
				"message": "Plugin installed but failed to start: " + err.Error(),
				"data": map[string]interface{}{
					"name":    meta.Name,
					"version": meta.Version,
					"started": false,
				},
			})
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    0,
		"message": "Plugin installed successfully",
		"data": map[string]interface{}{
			"name":    meta.Name,
			"version": meta.Version,
			"started": req.AutoStart,
		},
	})
}

// handleStart starts a plugin
func (s *AdminServer) handleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		jsonError(w, "name is required", http.StatusBadRequest)
		return
	}

	if err := s.pm.StartPlugin(r.Context(), req.Name); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, "Plugin started successfully")
}

// handleStop stops a plugin
func (s *AdminServer) handleStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		jsonError(w, "name is required", http.StatusBadRequest)
		return
	}

	if err := s.pm.StopPlugin(r.Context(), req.Name); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, "Plugin stopped successfully")
}

// handleUninstall uninstalls a plugin
func (s *AdminServer) handleUninstall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		jsonError(w, "name is required", http.StatusBadRequest)
		return
	}

	if err := s.pm.UninstallPlugin(r.Context(), req.Name); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, "Plugin uninstalled successfully")
}

// handleHealth returns health status
func (s *AdminServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	plugins := s.pm.GetRunningPlugins()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    0,
		"message": "ok",
		"data": map[string]interface{}{
			"status":          "healthy",
			"running_plugins": len(plugins),
		},
	})
}

func jsonError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    -1,
		"message": message,
	})
}

func jsonSuccess(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    0,
		"message": message,
	})
}

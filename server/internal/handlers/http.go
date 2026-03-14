package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bullshit-wtf/server/internal/hub"
)

func NewRouter(h *hub.Hub, _ *sql.DB) http.Handler {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/health", healthHandler)
	mux.HandleFunc("/api/create-game", createGameHandler(h))

	// WebSocket
	mux.HandleFunc("/ws", wsHandler(h))

	// Static files (React frontend)
	staticDir := "./static"
	if _, err := os.Stat(staticDir); err == nil {
		mux.Handle("/", spaHandler(staticDir))
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write([]byte("Bullshit.wtf API Server"))
		})
	}

	// CORS middleware
	return corsMiddleware(mux)
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

type CreateGameRequest struct {
	Lang           string `json:"lang"`
	TotalQuestions int    `json:"totalQuestions"`
}

type CreateGameResponse struct {
	PIN string `json:"pin"`
}

func createGameHandler(h *hub.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req CreateGameRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		if req.Lang == "" {
			req.Lang = "en"
		}
		if req.TotalQuestions <= 0 {
			req.TotalQuestions = 7
		}

		pin, _, err := h.CreateGame(req.Lang, req.TotalQuestions)
		if err != nil {
			http.Error(w, "failed to create game", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(CreateGameResponse{PIN: pin})
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// spaHandler serves a single-page application.
type spaFileHandler struct {
	staticDir string
}

func spaHandler(staticDir string) http.Handler {
	return &spaFileHandler{staticDir: staticDir}
}

func (h *spaFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(h.staticDir, filepath.Clean(r.URL.Path))

	// Check if the requested file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// If not, serve index.html for client-side routing
		if !strings.HasPrefix(r.URL.Path, "/api/") && !strings.HasPrefix(r.URL.Path, "/ws") {
			http.ServeFile(w, r, filepath.Join(h.staticDir, "index.html"))
			return
		}
		http.NotFound(w, r)
		return
	}

	http.FileServer(http.Dir(h.staticDir)).ServeHTTP(w, r)
}

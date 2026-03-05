package api

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"

	"devopstoolkit/youtube-automation/internal/aspect"
	"devopstoolkit/youtube-automation/internal/filesystem"
	"devopstoolkit/youtube-automation/internal/service"
	"devopstoolkit/youtube-automation/internal/video"

	"github.com/go-chi/chi/v5"
)

// Server is the HTTP API server.
type Server struct {
	router        chi.Router
	httpServer    *http.Server
	videoService  *service.VideoService
	videoManager  *video.Manager
	aspectService *aspect.Service
	filesystem    *filesystem.Operations
	apiToken      string
	frontendFS    fs.FS
}

// NewServer creates a new API server wired to the given service and manager.
// frontendFS should be the sub-directory containing the built frontend (e.g. the "dist" folder).
// Pass nil to disable frontend serving.
func NewServer(videoService *service.VideoService, videoManager *video.Manager, aspectService *aspect.Service, fsOps *filesystem.Operations, apiToken string, frontendFS fs.FS) *Server {
	s := &Server{
		router:        chi.NewRouter(),
		videoService:  videoService,
		videoManager:  videoManager,
		aspectService: aspectService,
		filesystem:    fsOps,
		apiToken:      apiToken,
		frontendFS:    frontendFS,
	}
	s.setupMiddleware()
	s.setupRoutes()
	return s
}

// setupRoutes registers all API routes.
func (s *Server) setupRoutes() {
	// Health check at root level
	s.router.Get("/health", s.handleHealth)

	// API routes - protected by bearer token auth
	s.router.Route("/api", func(r chi.Router) {
		r.Use(bearerTokenAuth(s.apiToken))
		// Categories
		r.Get("/categories", s.handleGetCategories)

		// Aspects
		r.Route("/aspects", func(r chi.Router) {
			r.Get("/", s.handleGetAspects)
			r.Get("/overview", s.handleGetAspectsOverview)
			r.Get("/{key}/fields", s.handleGetAspectFields)
			r.Get("/{key}/fields/{field}/completion", s.handleGetFieldCompletion)
		})

		// Videos
		r.Route("/videos", func(r chi.Router) {
			r.Get("/phases", s.handleGetPhases)
			r.Get("/list", s.handleGetVideosList)
			r.Get("/", s.handleGetVideos)
			r.Post("/", s.handleCreateVideo)
			r.Get("/{videoName}", s.handleGetVideo)
			r.Put("/{videoName}", s.handleUpdateVideo)
			r.Patch("/{videoName}", s.handlePatchVideoAspect)
			r.Delete("/{videoName}", s.handleDeleteVideo)
			r.Get("/{videoName}/progress", s.handleGetVideoProgress)
			r.Get("/{videoName}/progress/{aspect}", s.handleGetVideoAspectProgress)
			r.Get("/{videoName}/manuscript", s.handleGetVideoManuscript)
			r.Get("/{videoName}/animations", s.handleGetVideoAnimations)
		})
	})

	// SPA fallback: serve frontend static files, fall back to index.html
	if s.frontendFS != nil {
		s.router.Get("/*", s.spaHandler())
	}
}

// spaHandler returns an http.HandlerFunc that serves static files from the
// embedded frontend FS. Non-file paths (client-side routes) receive index.html.
func (s *Server) spaHandler() http.HandlerFunc {
	fileServer := http.FileServer(http.FS(s.frontendFS))
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")

		// Try to open the file from the embedded FS
		f, err := s.frontendFS.Open(path)
		if err == nil {
			f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		// File not found — serve index.html for client-side routing
		indexFile, err := fs.ReadFile(s.frontendFS, "index.html")
		if err != nil {
			http.Error(w, "index.html not found", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(indexFile)
	}
}

// Start begins listening on the given host and port.
func (s *Server) Start(host string, port int) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}
	slog.Info("starting API server", "addr", addr)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
}

// Router returns the chi router for testing.
func (s *Server) Router() chi.Router {
	return s.router
}

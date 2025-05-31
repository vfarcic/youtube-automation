package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"devopstoolkit/youtube-automation/internal/configuration"
	"devopstoolkit/youtube-automation/internal/filesystem"
	"devopstoolkit/youtube-automation/internal/service"
	"devopstoolkit/youtube-automation/internal/video"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Server represents the REST API server
type Server struct {
	router       chi.Router
	videoService *service.VideoService
	port         int
	httpServer   *http.Server
}

// NewServer creates a new API server
func NewServer() *Server {
	// Initialize dependencies
	filesystem := &filesystem.Operations{}
	videoManager := video.NewManager(filesystem.GetFilePath)
	videoService := service.NewVideoService("index.yaml", filesystem, videoManager)

	server := &Server{
		videoService: videoService,
		port:         configuration.GlobalSettings.API.Port,
	}

	server.setupRoutes()
	return server
}

// setupRoutes configures the API routes
func (s *Server) setupRoutes() {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS for development
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Videos endpoints
		r.Route("/videos", func(r chi.Router) {
			r.Post("/", s.createVideo)
			r.Get("/phases", s.getVideoPhases)
			r.Get("/", s.getVideos)         // with ?phase= query param
			r.Get("/list", s.getVideosList) // optimized lightweight endpoint
			r.Route("/{videoName}", func(r chi.Router) {
				r.Get("/", s.getVideo)
				r.Put("/", s.updateVideo)
				r.Delete("/", s.deleteVideo)
				r.Post("/move", s.moveVideo)

				// Phase-specific endpoints
				r.Put("/initial-details", s.updateVideoInitialDetails)
				r.Put("/work-progress", s.updateVideoWorkProgress)
				r.Put("/definition", s.updateVideoDefinition)
				r.Put("/post-production", s.updateVideoPostProduction)
				r.Put("/publishing", s.updateVideoPublishing)
				r.Put("/post-publish", s.updateVideoPostPublish)
			})
		})

		// Categories endpoint
		r.Get("/categories", s.getCategories)
	})

	// Health check
	r.Get("/health", s.healthCheck)

	s.router = r
}

// Start starts the API server
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)

	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Starting API server on port %d", s.port)
	return s.httpServer.ListenAndServe()
}

// Stop gracefully stops the API server
func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}

	log.Println("Stopping API server...")
	return s.httpServer.Shutdown(ctx)
}

// healthCheck handles health check requests
func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

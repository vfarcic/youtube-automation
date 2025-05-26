package api

import (
	"fmt"
	"log"
	"net/http"

	"devopstoolkit/youtube-automation/internal/api/handlers"
	"devopstoolkit/youtube-automation/internal/api/middleware"
	"devopstoolkit/youtube-automation/internal/service"
	"devopstoolkit/youtube-automation/internal/storage"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// Server represents the API server
type Server struct {
	router       *chi.Mux
	videoService *service.VideoService
	port         int
}

// NewServer creates a new API server
func NewServer(indexPath string, port int) *Server {
	// Create router
	r := chi.NewRouter()

	// Create storage operations
	storageOps := storage.NewOperations(indexPath)

	// Create services
	videoService := service.NewVideoService(storageOps)

	return &Server{
		router:       r,
		videoService: videoService,
		port:         port,
	}
}

// Start starts the API server
func (s *Server) Start() error {
	// Add middleware
	s.router.Use(chimiddleware.RequestID)
	s.router.Use(chimiddleware.RealIP)
	s.router.Use(middleware.RequestLogger)
	s.router.Use(middleware.ErrorHandler)
	s.router.Use(chimiddleware.Recoverer)
	s.router.Use(chimiddleware.SetHeader("Content-Type", "application/json"))

	// Setup routes
	s.setupRoutes()

	// Start server
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("Starting API server on %s", addr)
	return http.ListenAndServe(addr, s.router)
}

// setupRoutes configures all the API routes
func (s *Server) setupRoutes() {
	// Create handlers
	videoHandlers := handlers.NewVideoHandlers(s.videoService)
	phaseHandlers := handlers.NewPhaseHandlers(s.videoService)
	fileHandlers := handlers.NewFileHandlers(s.videoService)

	// Define routes
	s.router.Route("/api", func(r chi.Router) {
		// Videos endpoints
		r.Route("/videos", func(r chi.Router) {
			r.Post("/", videoHandlers.CreateVideo)
			r.Get("/phases", videoHandlers.GetVideoPhases)
			r.Get("/", videoHandlers.GetVideosByPhase)
			r.Route("/{video_id}", func(r chi.Router) {
				r.Get("/", videoHandlers.GetVideo)
				r.Put("/", videoHandlers.UpdateVideo)
				r.Delete("/", videoHandlers.DeleteVideo)
				r.Post("/move", fileHandlers.MoveVideoFiles)

				// Phase-specific endpoints
				r.Put("/initial", phaseHandlers.UpdateInitialPhase)
				r.Put("/work", phaseHandlers.UpdateWorkPhase)
				r.Put("/definition", phaseHandlers.UpdateDefinitionPhase)
				r.Put("/post-production", phaseHandlers.UpdatePostProductionPhase)
				r.Put("/publishing", phaseHandlers.UpdatePublishingPhase)
				r.Put("/post-publish", phaseHandlers.UpdatePostPublishPhase)
			})
		})

		// Categories endpoint
		r.Get("/categories", videoHandlers.GetCategories)
	})

	// Add health check endpoint
	s.router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})
}
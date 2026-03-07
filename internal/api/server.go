package api

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"

	"devopstoolkit/youtube-automation/internal/aspect"
	"devopstoolkit/youtube-automation/internal/configuration"
	"devopstoolkit/youtube-automation/internal/filesystem"
	"devopstoolkit/youtube-automation/internal/gdrive"
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
	aiService     AIService
	driveService  gdrive.DriveService
	driveFolderID string
	emailService      EmailService
	emailSettings     *configuration.SettingsEmail
	publishingService PublishingService
	apiToken          string
	frontendFS        fs.FS
}

// SetDriveService configures Google Drive upload support.
// If ds is nil, Drive upload endpoints return 501 Not Implemented.
func (s *Server) SetDriveService(ds gdrive.DriveService, folderID string) {
	s.driveService = ds
	s.driveFolderID = folderID
}

// SetPublishingService configures publishing support (YouTube, Hugo, social).
// If ps is nil, publishing endpoints return 501 Not Implemented.
func (s *Server) SetPublishingService(ps PublishingService) {
	s.publishingService = ps
}

// NewServer creates a new API server wired to the given service and manager.
// frontendFS should be the sub-directory containing the built frontend (e.g. the "dist" folder).
// Pass nil to disable frontend serving.
func NewServer(videoService *service.VideoService, videoManager *video.Manager, aspectService *aspect.Service, fsOps *filesystem.Operations, aiService AIService, apiToken string, frontendFS fs.FS) *Server {
	s := &Server{
		router:        chi.NewRouter(),
		videoService:  videoService,
		videoManager:  videoManager,
		aspectService: aspectService,
		filesystem:    fsOps,
		aiService:     aiService,
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

		// AI content generation
		r.Route("/ai", func(r chi.Router) {
			r.Post("/titles/{category}/{name}", s.handleAITitles)
			r.Post("/description/{category}/{name}", s.handleAIDescription)
			r.Post("/tags/{category}/{name}", s.handleAITags)
			r.Post("/tweets/{category}/{name}", s.handleAITweets)
			r.Post("/description-tags/{category}/{name}", s.handleAIDescriptionTags)
			r.Post("/shorts/{category}/{name}", s.handleAIShorts)
			r.Post("/thumbnails", s.handleAIThumbnails)
			r.Post("/translate", s.handleAITranslate)
			r.Post("/ama/content", s.handleAIAMAContent)
			r.Post("/ama/title", s.handleAIAMATitle)
			r.Post("/ama/description", s.handleAIAMADescription)
			r.Post("/ama/timecodes", s.handleAIAMATimecodes)
		})

		// Drive upload/download
		r.Route("/drive", func(r chi.Router) {
			r.Post("/upload/thumbnail/{videoName}", s.handleDriveUploadThumbnail)
			r.Post("/upload/video/{videoName}", s.handleDriveUploadVideo)
			r.Get("/download/video/{videoName}", s.handleDriveDownloadVideo)
		})

		// Action buttons (send emails, set flags)
		r.Route("/actions", func(r chi.Router) {
			r.Post("/request-thumbnail/{videoName}", s.handleRequestThumbnail)
			r.Post("/request-edit/{videoName}", s.handleRequestEdit)
		})

		// Publishing (YouTube upload, Hugo, transcript, metadata)
		r.Route("/publish", func(r chi.Router) {
			r.Post("/youtube/{videoName}", s.handlePublishYouTube)
			r.Post("/youtube/{videoName}/thumbnail", s.handlePublishThumbnail)
			r.Post("/youtube/{videoName}/shorts/{shortId}", s.handlePublishShort)
			r.Post("/hugo/{videoName}", s.handlePublishHugo)
			r.Get("/transcript/{videoId}", s.handleGetTranscript)
			r.Get("/metadata/{videoId}", s.handleGetMetadata)
		})

		// Social media posting
		r.Post("/social/{platform}/{videoName}", s.handleSocialPost)

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

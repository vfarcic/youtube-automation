package api

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"devopstoolkit/youtube-automation/internal/aspect"
	"devopstoolkit/youtube-automation/internal/configuration"
	"devopstoolkit/youtube-automation/internal/filesystem"
	"devopstoolkit/youtube-automation/internal/gdrive"
	"devopstoolkit/youtube-automation/internal/scheduler"
	"devopstoolkit/youtube-automation/internal/service"
	"devopstoolkit/youtube-automation/internal/thumbnail"
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
	analyzeService    AnalyzeService
	gitSync           GitSyncService
	imageGenerators   []thumbnail.ImageGenerator
	imageStore        *thumbnail.GeneratedImageStore
	photoDir          string
	dataDir           string
	apiToken          string
	frontendFS        fs.FS

	amaScheduler       *scheduler.Scheduler
	amaSchedulerCancel context.CancelFunc
	amaShutdownTimeout time.Duration

	// lifecycleMu guards httpServer, amaSchedulerCancel during Start/Shutdown
	// so calling Shutdown from one goroutine while Start runs in another is
	// race-clean.
	lifecycleMu sync.Mutex
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

// SetDataDir configures the data directory for serve mode.
func (s *Server) SetDataDir(dataDir string) {
	s.dataDir = dataDir
}

// SetAnalyzeService configures the analyze service for title analysis.
func (s *Server) SetAnalyzeService(as AnalyzeService) {
	s.analyzeService = as
}

// SetGitSync configures git sync for commit+push after file writes.
func (s *Server) SetGitSync(gs GitSyncService) {
	s.gitSync = gs
}

// SetThumbnailGeneration configures thumbnail generation support.
func (s *Server) SetThumbnailGeneration(generators []thumbnail.ImageGenerator, store *thumbnail.GeneratedImageStore, photoDir string) {
	s.imageGenerators = generators
	s.imageStore = store
	s.photoDir = photoDir
}

// SetAMAScheduler attaches a fully-configured AMA scheduler to the server.
// The scheduler is started by Server.Start and stopped by Server.Shutdown
// using the configured shutdown timeout (defaults to 30s).
func (s *Server) SetAMAScheduler(sched *scheduler.Scheduler, shutdownTimeout time.Duration) {
	s.amaScheduler = sched
	s.amaShutdownTimeout = shutdownTimeout
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
			r.Post("/tagline-and-illustrations/{category}/{name}", s.handleAITaglineAndIllustrations)
			r.Post("/thumbnails", s.handleAIThumbnails)
			r.Post("/translate", s.handleAITranslate)
			r.Post("/ama/content", s.handleAIAMAContent)
			r.Post("/ama/title", s.handleAIAMATitle)
			r.Post("/ama/description", s.handleAIAMADescription)
			r.Post("/ama/timecodes", s.handleAIAMATimecodes)
		})

		// Thumbnail generation
		r.Route("/thumbnails", func(r chi.Router) {
			r.Post("/generate", s.handleGenerateThumbnails)
			r.Get("/generated/{id}", s.handleGetGeneratedThumbnail)
			r.Post("/generated/{id}/select", s.handleSelectGeneratedThumbnail)
		})

		// Drive upload/download
		r.Route("/drive", func(r chi.Router) {
			r.Post("/upload/thumbnail/{videoName}", s.handleDriveUploadThumbnail)
			r.Post("/upload/video/{videoName}", s.handleDriveUploadVideo)
			r.Post("/upload/short/{videoName}/{shortId}", s.handleDriveUploadShort)
			r.Get("/download/video/{videoName}", s.handleDriveDownloadVideo)
			r.Get("/download/short/{videoName}/{shortId}", s.handleDriveDownloadShort)
		})

		// Action buttons (send emails, set flags)
		r.Route("/actions", func(r chi.Router) {
			r.Post("/request-thumbnail/{videoName}", s.handleRequestThumbnail)
			r.Post("/request-edit/{videoName}", s.handleRequestEdit)
			r.Post("/notify-sponsors/{videoName}", s.handleNotifySponsors)
		})

		// Publishing (YouTube upload, Hugo, transcript, metadata)
		r.Route("/publish", func(r chi.Router) {
			r.Post("/youtube/{videoName}", s.handlePublishYouTube)
			r.Post("/youtube/{videoName}/reupload", s.handleReuploadYouTube)
			r.Post("/youtube/{videoName}/thumbnail", s.handlePublishThumbnail)
			r.Post("/youtube/{videoName}/shorts/{shortId}", s.handlePublishShort)
			r.Post("/hugo/{videoName}", s.handlePublishHugo)
			r.Get("/transcript/{videoId}", s.handleGetTranscript)
			r.Get("/metadata/{videoId}", s.handleGetMetadata)
		})

		// Analyze (title analysis pipeline, timing recommendations)
		r.Route("/analyze", func(r chi.Router) {
			r.Post("/titles", s.handleAnalyzeTitles)
			r.Post("/titles/apply", s.handleApplyTitlesTemplate)
			r.Get("/timing", s.handleGetTimingRecommendations)
			r.Put("/timing", s.handlePutTimingRecommendations)
			r.Post("/timing/generate", s.handleGenerateTimingRecommendations)
		})

		// AMA workflow (generate from YouTube video, apply to YouTube)
		r.Route("/ama", func(r chi.Router) {
			r.Post("/generate", s.handleAMAGenerate)
			r.Post("/apply", s.handleAMAApply)
		})

		// Social media posting
		r.Post("/social/{platform}/{videoName}", s.handleSocialPost)

		// Videos
		r.Route("/videos", func(r chi.Router) {
			r.Get("/phases", s.handleGetPhases)
			r.Get("/list", s.handleGetVideosList)
			r.Get("/search", s.handleSearchVideos)
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
			r.Post("/{videoName}/apply-random-timing", s.handleApplyRandomTiming)
			r.Post("/{videoName}/thumbnail-config", s.handleSaveThumbnailConfig)
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

// Start begins listening on the given host and port. If an AMA scheduler is
// attached via SetAMAScheduler, it is started before the HTTP listener so
// scheduler-startup failures (invalid cron) surface before the server begins
// accepting traffic. If ListenAndServe itself fails (e.g. port already in
// use), the AMA scheduler is stopped before Start returns so the cron
// goroutines do not leak past the failed-start boundary.
func (s *Server) Start(host string, port int) error {
	s.lifecycleMu.Lock()
	if s.amaScheduler != nil {
		schedCtx, cancel := context.WithCancel(context.Background())
		s.amaSchedulerCancel = cancel
		if err := s.amaScheduler.Start(schedCtx); err != nil {
			cancel()
			s.amaSchedulerCancel = nil
			s.lifecycleMu.Unlock()
			return fmt.Errorf("start AMA scheduler: %w", err)
		}
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      s.router,
		WriteTimeout: 1 * time.Hour,
	}
	s.httpServer = srv
	s.lifecycleMu.Unlock()

	slog.Info("starting API server", "addr", addr)
	listenErr := srv.ListenAndServe()

	// http.ErrServerClosed is the normal-shutdown signal: Server.Shutdown
	// already stopped the scheduler. For any other error the listener
	// never came up (or died unexpectedly) and we must stop the scheduler
	// here. Scheduler.Stop is idempotent, so a concurrent Shutdown that
	// also nils amaSchedulerCancel is safe.
	if !errors.Is(listenErr, http.ErrServerClosed) {
		s.stopAMASchedulerAfterStartFailure()
	}
	return listenErr
}

// stopAMASchedulerAfterStartFailure stops the AMA scheduler when the HTTP
// listener fails before Server.Shutdown has had a chance to run. The original
// listener error is preserved by the caller; cleanup errors are logged only.
func (s *Server) stopAMASchedulerAfterStartFailure() {
	s.lifecycleMu.Lock()
	sched := s.amaScheduler
	cancel := s.amaSchedulerCancel
	s.amaSchedulerCancel = nil
	timeout := s.amaShutdownTimeout
	s.lifecycleMu.Unlock()

	if sched == nil {
		return
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	stopCtx, stopCancel := context.WithTimeout(context.Background(), timeout)
	defer stopCancel()
	if err := sched.Stop(stopCtx); err != nil {
		slog.Warn("AMA scheduler stop after listener failure returned error", "err", err)
	}
	if cancel != nil {
		cancel()
	}
}

// Shutdown gracefully shuts down the server. The AMA scheduler is stopped
// first (with its own bounded timeout) so in-flight ticks can finish before
// the HTTP listener closes. The scheduler-stop context is derived from the
// caller's ctx so a tight shutdown deadline (e.g. SIGTERM with a 10s budget)
// short-circuits the configured AMA timeout instead of blocking past it.
func (s *Server) Shutdown(ctx context.Context) error {
	s.lifecycleMu.Lock()
	sched := s.amaScheduler
	cancel := s.amaSchedulerCancel
	s.amaSchedulerCancel = nil
	timeout := s.amaShutdownTimeout
	httpSrv := s.httpServer
	s.lifecycleMu.Unlock()

	if sched != nil {
		if timeout <= 0 {
			timeout = 30 * time.Second
		}
		stopCtx, stopCancel := context.WithTimeout(ctx, timeout)
		if err := sched.Stop(stopCtx); err != nil {
			slog.Warn("AMA scheduler stop returned error", "err", err)
		}
		stopCancel()
		if cancel != nil {
			cancel()
		}
	}

	if httpSrv == nil {
		return nil
	}
	return httpSrv.Shutdown(ctx)
}

// Router returns the chi router for testing.
func (s *Server) Router() chi.Router {
	return s.router
}

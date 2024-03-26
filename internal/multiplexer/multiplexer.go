package multiplexer

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"
	"yegorov-boris/affise-test-task/configs"
	"yegorov-boris/affise-test-task/internal/handlers"
	"yegorov-boris/affise-test-task/internal/middleware"
	"yegorov-boris/affise-test-task/internal/services/cleaner"
	"yegorov-boris/affise-test-task/internal/services/progress"
	"yegorov-boris/affise-test-task/internal/services/scraper"
	"yegorov-boris/affise-test-task/internal/services/store"
	"yegorov-boris/affise-test-task/pkg/httpclient"
)

func Run(cfg *configs.Config) (func() error, error) {
	// Logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info(fmt.Sprintf("starting with config: %+v", cfg))

	// State
	state, err := progress.New(cfg.StorePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create state: %w", err)
	}

	// Rate limiter
	bucket := make(chan struct{}, cfg.MaxParallelIn)

	// HTTP Client
	httpClient := httpclient.New(cfg.HTTPClientTimeout)

	// HTTP Server
	mux := http.NewServeMux()
	linksPath, err := url.JoinPath(cfg.HTTPBasePath, "/links")
	if err != nil {
		return nil, fmt.Errorf("failed to join path: %w", err)
	}

	handlePost := middleware.NewRateLimiter(
		bucket,
		middleware.NewLogger(
			logger,
			handlers.NewPost(
				cfg.MaxLinksPerIn,
				state,
				scraper.New(logger, cfg.MaxParallelOutPerIn, httpClient),
				store.New(logger, cfg.StorePath),
			),
		),
	)

	mux.HandleFunc(linksPath, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errMsg := fmt.Sprintf("Sorry, only %s method is supported for this path.", http.MethodPost)
			http.Error(w, errMsg, http.StatusMethodNotAllowed)
			return
		}

		handlePost(w, r)
	})

	handleGet := middleware.NewLogger(
		logger,
		handlers.NewGet(linksPath, cfg.StorePath, state),
	)
	handleDelete := middleware.NewLogger(
		logger,
		handlers.NewDelete(linksPath, state),
	)
	mux.HandleFunc(fmt.Sprintf("%s/", linksPath), func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGet(w, r)
		case http.MethodDelete:
			handleDelete(w, r)
		default:
			errMsg := fmt.Sprintf("Sorry, only %s and %s methods are supported for this path.", http.MethodGet, http.MethodDelete)
			http.Error(w, errMsg, http.StatusMethodNotAllowed)
			return
		}
	})

	docsPath, err := url.JoinPath(cfg.HTTPBasePath, "/docs")
	if err != nil {
		return nil, fmt.Errorf("failed to join path: %w", err)
	}
	handleDocs := handlers.NewDocs(docsPath)
	mux.HandleFunc(fmt.Sprintf("%s/", docsPath), func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			errMsg := fmt.Sprintf("Sorry, only %s method is supported for this path.", http.MethodGet)
			http.Error(w, errMsg, http.StatusMethodNotAllowed)
			return
		}

		handleDocs(w, r)
	})

	srv := http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler: mux,
	}
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	logger.Info(fmt.Sprintf("http server listening on port %d", cfg.HTTPPort))

	// GC old files
	filesCleaner := cleaner.New(cfg.StoreTimeout, cfg.StorePath, logger)
	logger.Info("files cleaner started")

	return func() error {
		logger.Info("graceful shutdown started")

		if err := srv.Shutdown(context.Background()); err != nil {
			return fmt.Errorf("failed to shutdown HTTP server: %w", err)
		}
		logger.Info("http server stopped")

		filesCleaner.Shutdown()

		for !state.IsEmpty() {
			time.Sleep(cfg.GracefulShutdownStep)
		}

		logger.Info("graceful shutdown finished")

		return nil
	}, nil
}

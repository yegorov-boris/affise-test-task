package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"
	"yegorov-boris/affise-test-task/configs"
	"yegorov-boris/affise-test-task/internal/handlers"
	"yegorov-boris/affise-test-task/internal/middleware"
	"yegorov-boris/affise-test-task/internal/services/cleaner"
	"yegorov-boris/affise-test-task/internal/services/progress"
	"yegorov-boris/affise-test-task/internal/services/scraper"
	"yegorov-boris/affise-test-task/internal/services/store"
)

func main() {
	cfg := new(configs.Config)
	if err := cfg.Parse(); err != nil {
		log.Fatalf("failed to parse config: %s", err)
	}
	if err := cfg.Validate(); err != nil {
		log.Fatalf("failed to validate config: %s", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	state, err := progress.New(cfg.StorePath)
	if err != nil {
		log.Fatalf("failed to create state: %s", err)
	}

	bucket := make(chan struct{}, cfg.MaxParallelIn)

	handlePost := middleware.NewRateLimiter(
		bucket,
		middleware.NewLogger(
			logger,
			handlers.NewPost(
				cfg.MaxLinksPerIn,
				state,
				scraper.New(logger, cfg.MaxParallelOutPerIn, cfg.HTTPClientTimeout),
				store.New(logger, cfg.StorePath),
			),
		),
	)
	http.HandleFunc(cfg.HTTPBasePath, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errMsg := fmt.Sprintf("Sorry, only %s method is supported for this path.", http.MethodPost)
			http.Error(w, errMsg, http.StatusMethodNotAllowed)
			return
		}

		handlePost(w, r)
	})

	handleGet := middleware.NewLogger(
		logger,
		handlers.NewGet(cfg.HTTPBasePath, cfg.StorePath, state),
	)
	handleDelete := middleware.NewLogger(
		logger,
		handlers.NewDelete(cfg.HTTPBasePath, cfg.StorePath, state),
	)
	http.HandleFunc(fmt.Sprintf("%s/{id:[0-9]+}", cfg.HTTPBasePath), func(w http.ResponseWriter, r *http.Request) {
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

	srv := http.Server{
		Addr: fmt.Sprintf(":%d", cfg.HTTPPort),
	}
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	filesCleaner := cleaner.New(cfg.StoreTimeout, cfg.StorePath, logger)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	logger.Info("graceful shutdown started")

	if err := srv.Shutdown(context.Background()); err != nil {
		logger.Error(err.Error())
	}

	for i := 0; i < int(cfg.MaxParallelIn); i++ {
		bucket <- struct{}{}
	}

	filesCleaner.Shutdown()

	for !state.IsEmpty() {
		time.Sleep(cfg.GracefulShutdownStep)
	}

	logger.Info("graceful shutdown finished")
}

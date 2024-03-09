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
	"yegorov-boris/affise-test-task/internal/services/cleaner"
	"yegorov-boris/affise-test-task/internal/services/progress"
	"yegorov-boris/affise-test-task/internal/services/ratelimiter"
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

	ctx, cancel := context.WithCancel(context.Background())
	rateLimiterPost := ratelimiter.New(ctx, cfg.MaxParallelIn)
	rateLimiterDefault := ratelimiter.New(ctx, 0)
	state := new(progress.State)

	http.HandleFunc(cfg.HTTPBasePath, handlers.New(rateLimiterPost, logger, handlers.NewPost()))

	http.HandleFunc(fmt.Sprintf("%s/{id:[0-9]+}", cfg.HTTPBasePath), func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGet := handlers.New(
				rateLimiterDefault,
				logger,
				handlers.NewGet(cfg.HTTPBasePath, cfg.StorePath, state),
			)
			handleGet(w, r)
		case http.MethodDelete:
			handleDelete := handlers.New(rateLimiterDefault, logger, handlers.NewDelete())
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

	cleaner.Start(ctx, cfg.StoreTimeout, cfg.StorePath, logger)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	logger.Info("graceful shutdown started")
	cancel()
	for !rateLimiterPost.IsEmpty() {
		time.Sleep(cfg.GracefulShutdownStep)
	}
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Fatal(err)
	}
	logger.Info("graceful shutdown finished")
}

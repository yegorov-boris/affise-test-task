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

	rateLimiterPost := ratelimiter.New(cfg.MaxParallelIn)
	rateLimiterDefault := ratelimiter.New(0)

	http.HandleFunc(cfg.HTTPBasePath, handlers.New(rateLimiterPost, logger, handlers.NewPost()))

	http.HandleFunc(fmt.Sprintf("%s/{id:[0-9]+}", cfg.HTTPBasePath), func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGet := handlers.New(rateLimiterDefault, logger, handlers.NewGet())
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

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	logger.Info("graceful shutdown started")
	rateLimiterPost.Stop()
	rateLimiterDefault.Stop()
	for !rateLimiterPost.IsEmpty() {
		time.Sleep(cfg.GracefulShutdownStep)
	}
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Fatal(err)
	}
	logger.Info("graceful shutdown finished")
	//

	//maxLinksErrMsg := fmt.Sprintf("Maximum %d links per request are allowed.", MaxLinks)

	//
	//http.HandleFunc("/api/v1/links", func(w http.ResponseWriter, r *http.Request) {
	//	defer func() {
	//		select {
	//		case <-bucketIn:
	//		default:
	//		}
	//	}()
	//
	//	select {
	//	case bucketIn <- struct{}{}:
	//	default:
	//		http.Error(w, "Sorry, your request can not be currently served. Please, try again a bit later.", http.StatusTooManyRequests)
	//		return
	//	}
	//
	//	if r.Method != http.MethodPost {
	//		http.Error(w, "Sorry, only POST method is supported.", http.StatusMethodNotAllowed)
	//		return
	//	}
	//
	//	var data []byte
	//
	//	if _, err := r.Body.Read(data); err != nil {
	//		if errClose := r.Body.Close(); errClose != nil {
	//			if errLog := logger.Output(2, errClose.Error()); errLog != nil {
	//				log.Fatal(errLog)
	//			}
	//		}
	//		http.Error(w, "Failed to read request body.", http.StatusInternalServerError)
	//		if errLog := logger.Output(2, err.Error()); errLog != nil {
	//			log.Fatal(errLog)
	//		}
	//		return
	//	}
	//	if errClose := r.Body.Close(); errClose != nil {
	//		if errLog := logger.Output(2, errClose.Error()); errLog != nil {
	//			log.Fatal(errLog)
	//		}
	//	}
	//
	//	var links []string
	//
	//	if err := json.Unmarshal(data, &links); err != nil {
	//		http.Error(w, "Request body should be a JSON encoded array of strings.", http.StatusBadRequest)
	//		return
	//	}
	//
	//	linksCount := len(links)
	//	if linksCount > MaxLinks {
	//		http.Error(w, maxLinksErrMsg, http.StatusBadRequest)
	//		return
	//	}
	//
	//	uniqueLinks := make(map[string]struct{})
	//	for _, link := range links {
	//		uniqueLinks[link] = struct{}{}
	//	}
	//
	//	uniqueLinksCount := len(uniqueLinks)
	//	bucketOut := make(chan struct{}, MaxParallelOutPerIn)
	//	outputsCh := make(chan Output)
	//	ctx, cancel := context.WithTimeout(r.Context(), MaxTimeOut)
	//	defer cancel()
	//
	//	var (
	//		wg      sync.WaitGroup
	//		outputs []Output
	//	)
	//
	//	wg.Add(uniqueLinksCount + 1)
	//	go func() {
	//		defer wg.Done()
	//		for {
	//			select {
	//			case output := <-outputsCh:
	//				outputs = append(outputs, output)
	//				if len(outputs) == uniqueLinksCount {
	//					return
	//				}
	//			case <-ctx.Done():
	//				return
	//			}
	//		}
	//	}()
	//	for link := range uniqueLinks {
	//		go func(link string) {
	//			defer wg.Done()
	//			defer func() {
	//				select {
	//				case <-bucketOut:
	//				default:
	//				}
	//			}()
	//			select {
	//			case bucketOut <- struct{}{}:
	//				// todo: send http request, log error, sync.Once(cancel)
	//				outputsCh <- Output{}
	//			case <-ctx.Done():
	//				return
	//			}
	//		}(link)
	//	}
	//	wg.Wait()
	//
	//	if len(outputs) != uniqueLinksCount {
	//		http.Error(w, "Failed to process links.", http.StatusInternalServerError)
	//	}
	//
	//	if _, err := fmt.Fprintf(w, "Hello"); err != nil {
	//		if errLog := logger.Output(2, err.Error()); errLog != nil {
	//			log.Fatal(errLog)
	//		}
	//	}
	//})
	//
	//
}

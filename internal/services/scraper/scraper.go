package scraper

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"
	"yegorov-boris/affise-test-task/internal/models"
)

type Scraper struct {
	logger              *slog.Logger
	maxParallelOutPerIn uint32
	httpClientTimeout   time.Duration
}

func New(
	logger *slog.Logger,
	maxParallelOutPerIn uint32,
	httpClientTimeout time.Duration,
) *Scraper {
	return &Scraper{
		logger:              logger,
		maxParallelOutPerIn: maxParallelOutPerIn,
		httpClientTimeout:   httpClientTimeout,
	}
}

func (s *Scraper) Scrap(ctx context.Context, input models.Input) []models.Output {
	var wg sync.WaitGroup

	linksCount := len(input)
	results := make([]models.Output, linksCount)
	bucket := make(chan struct{}, s.maxParallelOutPerIn)
	ctx, cancel := context.WithTimeout(ctx, s.httpClientTimeout)
	defer cancel()

	wg.Add(linksCount)
	for i, link := range input {
		go func(i int, link string) {
			defer wg.Done()

			select {
			case bucket <- struct{}{}:
				defer func() {
					<-bucket
				}()

				output, err := s.get(ctx, link)
				if err != nil {
					cancel()
					s.logger.Error(fmt.Sprintf("failed to get response for link %q: %s", link, err))

					return
				}

				results[i] = output
			case <-ctx.Done():
			}
		}(i, link)
	}
	wg.Wait()

	return results
}

func (s *Scraper) get(ctx context.Context, link string) (models.Output, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, link, nil)
	if err != nil {
		return models.Output{}, fmt.Errorf("failed to build http request: %w", err)
	}

	client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		return models.Output{}, fmt.Errorf("failed to get response: %w", err)
	}

	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return models.Output{}, fmt.Errorf("failed to read response body: %w", err)
	}

	return models.Output{
		URL:        link,
		StatusCode: res.StatusCode,
		Body:       b,
	}, nil
}

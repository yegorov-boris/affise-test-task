package scraper

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"yegorov-boris/affise-test-task/internal/contracts"
	"yegorov-boris/affise-test-task/internal/models"
)

type Scraper struct {
	logger              *slog.Logger
	maxParallelOutPerIn uint32
	httpClient          contracts.HTTPClient
}

func New(
	logger *slog.Logger,
	maxParallelOutPerIn uint32,
	httpClient contracts.HTTPClient,
) *Scraper {
	return &Scraper{
		logger:              logger,
		maxParallelOutPerIn: maxParallelOutPerIn,
		httpClient:          httpClient,
	}
}

func (s *Scraper) Scrap(ctx context.Context, input models.Input) []models.Output {
	var wg sync.WaitGroup

	linksCount := len(input)
	results := make([]models.Output, linksCount)
	bucket := make(chan struct{}, s.maxParallelOutPerIn)
	c, cancel := context.WithCancel(ctx)
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

				output, err := s.httpClient.Get(c, link)
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

	for _, result := range results {
		if result.StatusCode == 0 {
			return nil
		}
	}

	return results
}

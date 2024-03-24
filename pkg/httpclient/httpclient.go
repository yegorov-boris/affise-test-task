package httpclient

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"
	"yegorov-boris/affise-test-task/internal/models"
)

type Client struct {
	timeout time.Duration
}

func New(timeout time.Duration) *Client {
	return &Client{
		timeout: timeout,
	}
}

func (c *Client) Get(ctx context.Context, link string) (models.Output, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, link, nil)
	if err != nil {
		return models.Output{}, fmt.Errorf("failed to build http request: %w", err)
	}

	client := &http.Client{
		Timeout: c.timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

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

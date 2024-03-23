package scraper

import (
	"context"
	"errors"
	"log/slog"
	"reflect"
	"sync"
	"testing"
	"time"
	"yegorov-boris/affise-test-task/internal/models"
)

type (
	httpClientMock struct {
		expected      map[string]httpClientMockResponse
		maxParallel   uint32
		parallelCount uint32
		m             sync.Mutex
	}

	httpClientMockResponse struct {
		output models.Output
		err    error
	}
)

func (c *httpClientMock) Get(ctx context.Context, link string) (models.Output, error) {
	c.m.Lock()
	c.parallelCount++
	if c.parallelCount > c.maxParallel {
		c.maxParallel = c.parallelCount
	}
	c.m.Unlock()
	done := make(chan struct{})
	defer func() {
		c.m.Lock()
		c.parallelCount--
		c.m.Unlock()
		close(done)
	}()

	response, ok := c.expected[link]
	if !ok {
		return models.Output{}, errors.New("unexpected link")
	}
	if response.err != nil {
		time.Sleep(50 * time.Millisecond)

		return response.output, response.err
	}
	time.Sleep(100 * time.Millisecond)

	return response.output, response.err
}

func TestScraper_Scrap(t *testing.T) {
	logger := slog.Default()
	maxParallelOutPerIn := uint32(2)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	input := []string{
		"https://example1.com",
		"https://example2.com",
		"https://example3.com",
	}
	outputs := map[string]httpClientMockResponse{
		"https://example1.com": {
			output: models.Output{
				URL:        "https://example1.com",
				StatusCode: 200,
				Body:       []byte("<html><body>some html</body></html>"),
			},
			err: nil,
		},
		"https://example2.com": {
			output: models.Output{
				URL:        "https://example2.com",
				StatusCode: 200,
				Body:       []byte("some text"),
			},
			err: nil,
		},
		"https://example3.com": {
			output: models.Output{
				URL:        "https://example3.com",
				StatusCode: 200,
				Body:       []byte(`{"foo": "bar"}`),
			},
			err: nil,
		},
	}
	clientMock := func() *httpClientMock {
		return &httpClientMock{
			expected: outputs,
		}
	}
	type fields struct {
		logger              *slog.Logger
		maxParallelOutPerIn uint32
		httpClient          *httpClientMock
	}
	type args struct {
		ctx   context.Context
		input models.Input
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []models.Output
	}{
		{
			name: "should get request for every link and throttle outgoing requests",
			fields: fields{
				logger:              logger,
				maxParallelOutPerIn: maxParallelOutPerIn,
				httpClient:          clientMock(),
			},
			args: args{
				ctx:   ctx,
				input: input[:3],
			},
			want: []models.Output{
				outputs["https://example1.com"].output,
				outputs["https://example2.com"].output,
				outputs["https://example3.com"].output,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(tt.fields.logger, tt.fields.maxParallelOutPerIn, tt.fields.httpClient)
			if got, _ := s.Scrap(tt.args.ctx, tt.args.input); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Scrap() = %v, want %v", got, tt.want)
			}
			if tt.fields.httpClient.maxParallel > tt.fields.maxParallelOutPerIn {
				t.Errorf(
					"Throttling failed: expected no more than %d concurrent outgoing requests - get %d",
					tt.fields.maxParallelOutPerIn,
					tt.fields.httpClient.maxParallel,
				)
			}
		})
	}
}

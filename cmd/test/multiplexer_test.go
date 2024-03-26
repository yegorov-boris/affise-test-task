package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
	"yegorov-boris/affise-test-task/configs"
	"yegorov-boris/affise-test-task/internal/models"
	"yegorov-boris/affise-test-task/internal/multiplexer"
	"yegorov-boris/affise-test-task/pkg/dotenv"
)

func Test_Multiplexer(t *testing.T) {
	if err := dotenv.Load("../../.env"); err != nil {
		log.Fatal(err)
	}
	cfg := new(configs.ConfigTest)
	if err := cfg.Parse(); err != nil {
		log.Fatal(err)
	}
	bodies := map[string][]byte{
		"1": []byte("<html>foo</html>"),
		"2": []byte(`{"foo": "bar"}`),
		"3": {},
		"4": nil,
	}
	http.HandleFunc("/example/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		lastPart := parts[len(parts)-1]
		body, ok := bodies[lastPart]
		if !ok {
			log.Fatal("fake body not found")
		}
		if lastPart == "3" {
			time.Sleep(50 * time.Millisecond)
		}
		if lastPart == "4" {
			time.Sleep(200 * time.Millisecond)
		}
		if _, err := w.Write(body); err != nil {
			log.Fatal(err)
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

	testAPI := fmt.Sprintf("http://%s:%d", cfg.HTTPHost, cfg.HTTPPort)
	testLink := func(path string) string {
		return fmt.Sprintf("%s/example/%s", testAPI, path)
	}

	type wantResponse struct {
		postStatusCode int
		getStatusCode  int
		getBody        []models.Output
	}
	tests := []struct {
		name              string
		links             [][]string
		rateLimit         uint32
		httpClientTimeout time.Duration
		cancel            bool
		shutdown          bool
		want              []wantResponse
	}{
		{
			name: "should POST a list of links and then GET outputs",
			links: [][]string{
				{
					testLink("1"),
					testLink("2"),
				},
			},
			rateLimit:         100,
			httpClientTimeout: 100 * time.Millisecond,
			want: []wantResponse{
				{
					postStatusCode: http.StatusAccepted,
					getStatusCode:  http.StatusOK,
					getBody: []models.Output{
						{
							URL:        testLink("1"),
							StatusCode: http.StatusOK,
							Body:       string(bodies["1"]),
						},
						{
							URL:        testLink("2"),
							StatusCode: http.StatusOK,
							Body:       string(bodies["2"]),
						},
					},
				},
			},
		},
		{
			name: "should limit POST requests rate",
			links: [][]string{
				{
					testLink("3"),
				},
				{
					testLink("1"),
				},
			},
			rateLimit:         1,
			httpClientTimeout: 100 * time.Millisecond,
			want: []wantResponse{
				{
					postStatusCode: http.StatusAccepted,
					getStatusCode:  http.StatusOK,
					getBody: []models.Output{
						{
							URL:        testLink("3"),
							StatusCode: http.StatusOK,
							Body:       string(bodies["3"]),
						},
					},
				},
				{
					postStatusCode: http.StatusTooManyRequests,
				},
			},
		},
		{
			name: "should respond with error when one of outgoing requests is stopped by timeout",
			links: [][]string{
				{
					testLink("1"),
					testLink("4"),
				},
			},
			rateLimit:         100,
			httpClientTimeout: 100 * time.Millisecond,
			want: []wantResponse{
				{
					postStatusCode: http.StatusAccepted,
					getStatusCode:  http.StatusOK,
					getBody:        nil,
				},
			},
		},
		{
			name: "should stop outgoing requests when cancelled by a client",
			links: [][]string{
				{
					testLink("1"),
					testLink("3"),
				},
			},
			rateLimit:         100,
			httpClientTimeout: 100 * time.Millisecond,
			cancel:            true,
			want: []wantResponse{
				{
					postStatusCode: http.StatusAccepted,
					getStatusCode:  http.StatusOK,
					getBody:        nil,
				},
			},
		},
		{
			name: "should shutdown gracefully",
			links: [][]string{
				{
					testLink("3"),
				},
			},
			rateLimit:         100,
			httpClientTimeout: 100 * time.Millisecond,
			shutdown:          true,
			want: []wantResponse{
				{
					postStatusCode: http.StatusAccepted,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				wg sync.WaitGroup
			)

			mainConfig, err := configs.New()
			if err != nil {
				t.Error(err)
				return
			}
			mainConfig.MaxParallelIn = tt.rateLimit
			mainConfig.HTTPClientTimeout = tt.httpClientTimeout
			mainConfig.StorePath = "../../store"
			shutdown, err := multiplexer.Run(mainConfig)
			if err != nil {
				t.Error(err)
				return
			}

			var errShutdown error

			shutdownOnce := sync.OnceFunc(func() {
				errShutdown = shutdown()
				if errShutdown != nil {
					t.Errorf("graceful shutdown failed: %s", errShutdown)
					errShutdown = fmt.Errorf("graceful shutdown failed: %s", errShutdown)
				}
			})
			defer shutdownOnce()

			multiplexerAPI, err := url.JoinPath(
				fmt.Sprintf("http://%s:%d", cfg.MultiplexerHost, mainConfig.HTTPPort),
				mainConfig.HTTPBasePath,
				"/links",
			)
			if err != nil {
				t.Error(err)
				return
			}

			// wait for http server to start
			time.Sleep(10 * time.Millisecond)

			n := len(tt.links)
			output := make([][]models.Output, n)
			errs := make([]error, n)
			wg.Add(n)
			for i, input := range tt.links {
				go func(i int, input []string) {
					defer wg.Done()
					b, err := json.Marshal(input)
					if err != nil {
						errs[i] = err
						return
					}
					res, err := http.Post(multiplexerAPI, "application/json", bytes.NewBuffer(b))
					if err != nil {
						errs[i] = err
						return
					}

					if res.StatusCode != tt.want[i].postStatusCode {
						errs[i] = fmt.Errorf("Expected status code %d, got %d", tt.want[i].postStatusCode, res.StatusCode)
						return
					}
					body, err := io.ReadAll(res.Body)
					if err != nil {
						errs[i] = err
						return
					}
					if res.StatusCode != http.StatusAccepted {
						return
					}
					id, err := strconv.ParseUint(string(body), 10, 64)
					if err != nil {
						errs[i] = fmt.Errorf("Invalid response body: %s", err)
						return
					}

					if tt.shutdown {
						shutdownOnce()
						if errShutdown != nil {
							errs[i] = errShutdown
							return
						}
						time.Sleep(2 * tt.httpClientTimeout)
						if _, err := os.Open(fmt.Sprintf("../../store/%d.json", id)); err != nil {
							errs[i] = fmt.Errorf("response bodies were not written to a file: %s", err)
						}
						return
					}

					if tt.cancel {
						req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/%d", multiplexerAPI, id), nil)
						if err != nil {
							errs[i] = err
							return
						}
						res, err := http.DefaultClient.Do(req)
						if err != nil {
							errs[i] = fmt.Errorf("Failed to cancel: %s", err)
							return
						}
						if res.StatusCode != http.StatusNoContent {
							errs[i] = fmt.Errorf("Expected status %d, got %d", http.StatusNoContent, res.StatusCode)
							return
						}
					}

					time.Sleep(2 * tt.httpClientTimeout)
					res, err = http.Get(fmt.Sprintf("%s/%d", multiplexerAPI, id))
					if err != nil {
						errs[i] = err
						return
					}
					body, err = io.ReadAll(res.Body)
					if err != nil {
						errs[i] = err
						return
					}
					if string(body) == "Your request is in progress. Please, try a bit later." {
						output[i] = nil
						return
					}
					json.Unmarshal(body, &output[i])
				}(i, input)
				time.Sleep(10 * time.Millisecond)
			}
			wg.Wait()
			for _, err := range errs {
				if err != nil {
					t.Error(err)
					return
				}
			}
			for i, w := range tt.want {
				if !reflect.DeepEqual(output[i], w.getBody) {
					t.Errorf("Expected %v\nGot %v", w.getBody, output[i])
					return
				}
			}
		})
	}
}

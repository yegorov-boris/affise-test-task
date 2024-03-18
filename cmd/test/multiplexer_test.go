package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"net/url"
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

	randomBytes := func() []byte {
		n := rand.Uint32N(255)
		result := make([]byte, n)
		for i := 0; i < int(n); i++ {
			result[i] = uint8(rand.Uint32N(255))
		}

		return result
	}
	bodies := map[string][]byte{
		"1": randomBytes(),
		"2": randomBytes(),
		"3": nil,
	}
	http.HandleFunc("/example/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		lastPart := parts[len(parts)-1]
		body, ok := bodies[lastPart]
		if !ok {
			log.Fatal("fake body not found")
		}
		if lastPart == "3" {
			time.Sleep(100 * time.Millisecond)
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
		getBody        []models.Output
	}
	tests := []struct {
		name      string
		links     [][]string
		rateLimit uint32
		want      []wantResponse
	}{
		{
			name: "should POST a list of links and then GET outputs",
			links: [][]string{
				{
					testLink("1"),
					testLink("2"),
				},
			},
			rateLimit: 100,
			want: []wantResponse{
				{
					postStatusCode: http.StatusAccepted,
					getBody: []models.Output{
						{
							URL:        testLink("1"),
							StatusCode: http.StatusOK,
							Body:       bodies["1"],
						},
						{
							URL:        testLink("2"),
							StatusCode: http.StatusOK,
							Body:       bodies["2"],
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
			rateLimit: 1,
			want: []wantResponse{
				{
					postStatusCode: http.StatusAccepted,
					getBody:        nil,
				},
				{
					postStatusCode: http.StatusTooManyRequests,
					getBody:        nil,
				},
			},
		},
		//{
		//	name: "should respond with error when one of outgoing requests is stopped by timeout",
		//},
		//{
		//	name: "should stop outgoing requests when cancelled by a client",
		//},
		//{
		//	name: "should shutdown gracefully",
		//},
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
			mainConfig.StorePath = "../../store"
			shutdown, err := multiplexer.Run(mainConfig)
			if err != nil {
				t.Error(err)
				return
			}
			//time.Sleep(100 * time.Millisecond)

			defer func() {
				if err := shutdown(); err != nil {
					t.Errorf("graceful shutdown failed: %s", err)
				}
			}()

			multiplexerAPI, err := url.JoinPath(
				fmt.Sprintf("http://%s:%d", cfg.MultiplexerHost, mainConfig.HTTPPort),
				mainConfig.HTTPBasePath,
			)
			if err != nil {
				t.Error(err)
				return
			}

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
					id, err := strconv.ParseUint(string(body), 10, 64)
					if err != nil {
						errs[i] = fmt.Errorf("Invalid response body: %s", err)
						return
					}

					time.Sleep(50 * time.Millisecond)
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
					if err := json.Unmarshal(body, &output[i]); err != nil {
						errs[i] = err
					}
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

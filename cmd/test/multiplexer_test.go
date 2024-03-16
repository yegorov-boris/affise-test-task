package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"testing"
	"time"
	"yegorov-boris/affise-test-task/configs"
	"yegorov-boris/affise-test-task/internal/models"
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

	http.HandleFunc("/example", func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprintf(w, "example"); err != nil {
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

	multiplexerAPI, err := url.JoinPath(
		fmt.Sprintf("http://%s:%d", cfg.MultiplexerHost, cfg.MultiplexerPort),
		cfg.MultiplexerBasePath,
	)
	if err != nil {
		log.Fatal(err)
	}
	testAPI := fmt.Sprintf("http://%s:%d", cfg.HTTPHost, cfg.HTTPPort)
	testLink := func(path string) string {
		return fmt.Sprintf("%s/%s", testAPI, path)
	}
	tests := []struct {
		name  string
		links []string
		want  []models.Output
	}{
		{
			name: "should POST a list of links and then GET outputs",
			links: []string{
				testLink("example"),
			},
			want: []models.Output{
				{
					URL:        testLink("example"),
					StatusCode: http.StatusOK,
					Body:       []byte("example"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output []models.Output

			b, err := json.Marshal(tt.links)
			if err != nil {
				t.Error(err)
			}
			res, err := http.Post(multiplexerAPI, "application/json", bytes.NewBuffer(b))
			if err != nil {
				t.Error(err)
			}

			if res.StatusCode != http.StatusAccepted {
				t.Errorf("Expected status code %d, got %d", http.StatusAccepted, res.StatusCode)
			}
			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Error(err)
			}
			id, err := strconv.ParseUint(string(body), 10, 64)
			if err != nil {
				t.Errorf("Invalid response body: %s", err)
			}

			time.Sleep(50 * time.Millisecond)
			res, err = http.Get(fmt.Sprintf("%s/%d", multiplexerAPI, id))
			if err != nil {
				t.Error(err)
			}
			body, err = io.ReadAll(res.Body)
			if err != nil {
				t.Error(err)
			}
			if err := json.Unmarshal(body, &output); err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(output, tt.want) {
				t.Errorf("Expected %v, got %v", tt.want, output)
			}
		})
	}
}

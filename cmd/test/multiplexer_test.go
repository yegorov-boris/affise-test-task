package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"yegorov-boris/affise-test-task/configs"
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
	tests := []struct {
		links []string
	}{
		{
			links: []string{
				fmt.Sprintf("http://%s:%d/example", cfg.HTTPHost, cfg.HTTPPort),
			},
		},
	}
	for _, tt := range tests {
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
		if _, err := strconv.ParseUint(string(body), 10, 64); err != nil {
			t.Errorf("Invalid response body: %s", err)
		}
	}
}

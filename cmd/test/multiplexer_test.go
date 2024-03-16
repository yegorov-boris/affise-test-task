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
	}
	http.HandleFunc("/example/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		lastPart := parts[len(parts)-1]
		body, ok := bodies[lastPart]
		if !ok {
			log.Fatal("fake body not found")
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

	multiplexerAPI, err := url.JoinPath(
		fmt.Sprintf("http://%s:%d", cfg.MultiplexerHost, cfg.MultiplexerPort),
		cfg.MultiplexerBasePath,
	)
	if err != nil {
		log.Fatal(err)
	}
	testAPI := fmt.Sprintf("http://%s:%d", cfg.HTTPHost, cfg.HTTPPort)
	testLink := func(path string) string {
		return fmt.Sprintf("%s/example/%s", testAPI, path)
	}
	tests := []struct {
		name  string
		links []string
		want  []models.Output
	}{
		{
			name: "should POST a list of links and then GET outputs",
			links: []string{
				testLink("1"),
				testLink("2"),
			},
			want: []models.Output{
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
				t.Errorf("Expected %v\nGot %v", tt.want, output)
			}
		})
	}
}

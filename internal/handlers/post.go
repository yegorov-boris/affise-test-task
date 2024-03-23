package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"yegorov-boris/affise-test-task/internal/contracts"
	"yegorov-boris/affise-test-task/internal/models"
)

func NewPost(
	maxLinksPerIn uint32,
	state contracts.State,
	scraper contracts.Scraper,
	store contracts.Store,
) contracts.HandlerWithErr {
	return func(w http.ResponseWriter, r *http.Request) (e error) {
		var links models.Input

		callback, hasCallback := r.Context().Value("callback").(func())
		if hasCallback {
			defer func() {
				if e != nil {
					callback()
				}
			}()
		}

		data, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body.", http.StatusInternalServerError)

			return fmt.Errorf("failed to read request body: %w", err)
		}

		defer r.Body.Close()

		if err := json.Unmarshal(data, &links); err != nil {
			http.Error(w, "Request body should be a JSON encoded array of strings.", http.StatusBadRequest)

			return fmt.Errorf("failed to decode request body from JSON: %w", err)
		}

		if err := links.Validate(maxLinksPerIn); err != nil {
			errMsg := "Some strings in the request body are not valid links."
			if errors.Is(models.ErrNoLinks, err) {
				errMsg = "At least 1 link per request should be provided."
			}
			if errors.Is(models.ErrTooManyLinks, err) {
				errMsg = fmt.Sprintf("Maximum %d links per request are allowed", maxLinksPerIn)
			}

			http.Error(w, errMsg, http.StatusBadRequest)

			return fmt.Errorf("invalid request body: %w", err)
		}

		id, ctx := state.Start()
		w.WriteHeader(http.StatusAccepted)
		if _, err := fmt.Fprintf(w, "%d", id); err != nil {
			state.Finish(id)

			return fmt.Errorf("failed to write response body: %w", err)
		}

		go func() {
			store.Save(id, scraper.Scrap(ctx, links))
			state.Finish(id)
			if hasCallback {
				callback()
			}
		}()

		return nil
	}
}

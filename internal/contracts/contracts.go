package contracts

import (
	"context"
	"net/http"
	"yegorov-boris/affise-test-task/internal/models"
)

type (
	State interface {
		Start() (uint64, context.Context)
		Finish(uint64)
		Check(uint64) bool
		Cancel(uint64) bool
	}

	RateLimiter interface {
		Try() bool
		Free()
	}

	Scraper interface {
		Scrap(context.Context, models.Input) []models.Output
	}

	Store interface {
		Save(uint64, []models.Output)
	}

	Handler = func(w http.ResponseWriter, r *http.Request)

	HandlerWithErr = func(w http.ResponseWriter, r *http.Request) error
)

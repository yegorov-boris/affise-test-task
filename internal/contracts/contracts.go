package contracts

import (
	"net/http"
	"yegorov-boris/affise-test-task/internal/models"
)

type (
	Store interface {
		Load(uint64) ([]models.Output, error)
		Store(uint64, []models.Output) error
		Delete(uint64) error
	}

	State interface {
		Start(uint64)
		Finish(uint64)
		Check(uint64)
	}

	Tryer interface {
		Try() bool
	}

	Handler = func(w http.ResponseWriter, r *http.Request)

	HandlerWithErr = func(w http.ResponseWriter, r *http.Request) error
)

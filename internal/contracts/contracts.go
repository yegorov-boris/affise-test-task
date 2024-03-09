package contracts

import (
	"context"
	"net/http"
)

type (
	State interface {
		Start() (uint64, context.Context)
		Finish(uint64)
		Check(uint64) bool
		Cancel(uint64) bool
	}

	Tryer interface {
		Try() bool
	}

	Handler = func(w http.ResponseWriter, r *http.Request)

	HandlerWithErr = func(w http.ResponseWriter, r *http.Request) error
)

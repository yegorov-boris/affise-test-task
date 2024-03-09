package progress

import (
	"context"
	"sync"
	"sync/atomic"
)

type State struct {
	uid   atomic.Uint64
	state sync.Map
}

func (s *State) Start() (uint64, context.Context) {
	id := s.uid.Add(1)
	ctx, cancel := context.WithCancel(context.Background())
	s.state.Store(id, cancel)

	return id, ctx
}

func (s *State) Finish(id uint64) {
	s.state.Delete(id)
}

func (s *State) Check(id uint64) bool {
	_, ok := s.state.Load(id)

	return ok
}

func (s *State) Cancel(id uint64) bool {
	cancel, ok := s.state.Load(id)
	if !ok {
		return false
	}

	cancel.(func())()

	return true
}

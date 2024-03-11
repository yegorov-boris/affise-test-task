package progress

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

type State struct {
	uid   atomic.Uint64
	len   atomic.Int32
	state sync.Map
}

func New(storePath string) (*State, error) {
	entries, err := os.ReadDir(storePath)
	if err != nil {
		return nil, fmt.Errorf("failed to list %q: %w", storePath, err)
	}

	s := new(State)

	if len(entries) == 0 {
		return s, nil
	}

	lastFile := entries[len(entries)-1].Name()
	baseName := strings.TrimSuffix(filepath.Base(lastFile), ".json")
	lastID, err := strconv.ParseUint(baseName, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid filename in the store: %w", err)
	}

	s.uid.Store(lastID)

	return s, nil
}

func (s *State) Start() (uint64, context.Context) {
	id := s.uid.Add(1)
	ctx, cancel := context.WithCancel(context.Background())
	s.state.Store(id, cancel)
	s.len.Add(1)

	return id, ctx
}

func (s *State) Finish(id uint64) {
	s.state.Delete(id)
	s.len.Add(-1)
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

func (s *State) IsEmpty() bool {
	return s.len.Load() == 0
}

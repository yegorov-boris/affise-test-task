package progress

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
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
	var maxID uint64

	if err := os.MkdirAll(storePath, fs.ModeDir); err != nil {
		return nil, fmt.Errorf("failed to create %q: %w", storePath, err)
	}

	entries, err := os.ReadDir(storePath)
	if err != nil {
		return nil, fmt.Errorf("failed to list %q: %w", storePath, err)
	}

	s := new(State)

	for _, e := range entries {
		name := e.Name()
		if path.Ext(name) != ".json" {
			continue
		}

		baseName := strings.TrimSuffix(filepath.Base(name), ".json")
		if id, err := strconv.ParseUint(baseName, 10, 64); err == nil && id > maxID {
			maxID = id
		}
	}

	s.uid.Store(maxID)

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

	cancel.(context.CancelFunc)()

	return true
}

func (s *State) IsEmpty() bool {
	return s.len.Load() == 0
}

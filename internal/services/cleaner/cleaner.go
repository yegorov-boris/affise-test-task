package cleaner

import (
	"fmt"
	"log/slog"
	"os"
	"time"
)

type Cleaner struct {
	storeTimeout time.Duration
	storePath    string
	logger       *slog.Logger
	done         chan struct{}
}

func New(
	storeTimeout time.Duration,
	storePath string,
	logger *slog.Logger,
) *Cleaner {
	c := Cleaner{
		storeTimeout: storeTimeout,
		storePath:    storePath,
		logger:       logger,
		done:         make(chan struct{}),
	}

	ticker := time.NewTicker(storeTimeout)
	go func() {
		for {
			select {
			case <-c.done:
				return
			case <-ticker.C:
				c.do()
			}
		}
	}()

	return &c
}

func (c *Cleaner) do() {
	c.logger.Info("Cleaner started")

	entries, err := os.ReadDir(c.storePath)
	if err != nil {
		c.logger.Error(fmt.Sprintf("Cleaner failed to list %q: %s", c.storePath, err))
		return
	}

	for _, e := range entries {
		fileInfo, err := e.Info()
		if err != nil {
			c.logger.Error(fmt.Sprintf("Cleaner failed to get %q file info: %s", e.Name(), err))
			continue
		}

		if time.Since(fileInfo.ModTime()) < c.storeTimeout {
			continue
		}

		if err := os.Remove(e.Name()); err != nil {
			c.logger.Error(fmt.Sprintf("Cleaner failed to remove %q: %s", e.Name(), err))
		}
	}

	c.logger.Info("Cleaner finished")
}

func (c *Cleaner) Shutdown() {
	c.done <- struct{}{}
}

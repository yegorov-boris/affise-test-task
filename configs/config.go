package configs

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"
)

type Config struct {
	StorePath            string
	StoreTimeout         time.Duration
	HTTPPort             uint32
	HTTPBasePath         string
	HTTPClientTimeout    time.Duration
	GracefulShutdownStep time.Duration
	MaxLinksPerIn        uint32
	MaxParallelIn        uint32
	MaxParallelOutPerIn  uint32
}

func New() (*Config, error) {
	cfg := new(Config)

	if err := cfg.parse(); err != nil {
		return nil, fmt.Errorf("failed to parse config: %s", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("failed to validate config: %s", err)
	}

	return cfg, nil
}

func (c *Config) parse() error {
	var err error

	c.StorePath = os.Getenv("STORE_PATH")

	c.StoreTimeout, err = time.ParseDuration(os.Getenv("STORE_TIMEOUT"))
	if err != nil {
		return fmt.Errorf("failed to parse %q env var: %w", "STORE_TIMEOUT", err)
	}

	c.HTTPPort, err = parseUint32("HTTP_PORT")
	if err != nil {
		return fmt.Errorf("failed to parse %q env var: %w", "HTTP_PORT", err)
	}

	c.HTTPBasePath, err = url.JoinPath(os.Getenv("HTTP_BASE_PATH"), "")
	if err != nil {
		return fmt.Errorf("failed to parse %q env var: %w", "HTTP_BASE_PATH", err)
	}

	c.HTTPClientTimeout, err = time.ParseDuration(os.Getenv("HTTP_CLIENT_TIMEOUT"))
	if err != nil {
		return fmt.Errorf("failed to parse %q env var: %w", "HTTP_CLIENT_TIMEOUT", err)
	}

	c.GracefulShutdownStep, err = time.ParseDuration(os.Getenv("GRACEFUL_SHUTDOWN_STEP"))
	if err != nil {
		return fmt.Errorf("failed to parse %q env var: %w", "GRACEFUL_SHUTDOWN_STEP", err)
	}

	c.MaxLinksPerIn, err = parseUint32("MAX_LINKS_PER_IN")
	if err != nil {
		return fmt.Errorf("failed to parse %q env var: %w", "MAX_LINKS_PER_IN", err)
	}

	c.MaxParallelIn, err = parseUint32("MAX_PARALLEL_IN")
	if err != nil {
		return fmt.Errorf("failed to parse %q env var: %w", "MAX_PARALLEL_IN", err)
	}

	c.MaxParallelOutPerIn, err = parseUint32("MAX_PARALLEL_OUT_PER_IN")
	if err != nil {
		return fmt.Errorf("failed to parse %q env var: %w", "MAX_PARALLEL_OUT_PER_IN", err)
	}

	return nil
}

func (c *Config) validate() error {
	if len(c.HTTPBasePath) != 0 && c.HTTPBasePath[len(c.HTTPBasePath)-1] == byte('/') {
		return fmt.Errorf("%q parameter must not end with \"/\"", "HTTPBasePath")
	}

	if len(c.StorePath) == 0 {
		return fmt.Errorf("%q parameter must not be empty", "StorePath")
	}

	if c.MaxLinksPerIn < 1 {
		return fmt.Errorf("%q parameter must be at least 1", "MaxLinksPerIn")
	}

	if c.MaxParallelIn < 1 {
		return fmt.Errorf("%q parameter must be at least 1", "MaxParallelIn")
	}

	if c.MaxParallelOutPerIn < 1 {
		return fmt.Errorf("%q parameter must be at least 1", "MaxParallelOutPerIn")
	}

	if c.MaxParallelOutPerIn > c.MaxLinksPerIn {
		return fmt.Errorf("%q parameter must not be greater than %q parameter", "MaxParallelOutPerIn", "MaxLinksPerIn")
	}

	if c.HTTPPort < 1 || c.HTTPPort >= (1<<16) {
		return fmt.Errorf("%q parameter must be from %d to %d", "HTTPPort", 1, 1<<16-1)
	}

	return nil
}

func parseUint32(name string) (uint32, error) {
	u, err := strconv.ParseUint(os.Getenv(name), 10, 32)

	return uint32(u), err
}

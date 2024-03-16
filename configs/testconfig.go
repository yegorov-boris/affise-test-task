package configs

import (
	"fmt"
	"os"
)

type ConfigTest struct {
	HTTPHost            string
	HTTPPort            uint32
	MultiplexerHost     string
	MultiplexerPort     uint32
	MultiplexerBasePath string
}

func (c *ConfigTest) Parse() error {
	var err error

	c.HTTPHost = os.Getenv("TEST_HOST")

	c.HTTPPort, err = parseUint32("TEST_PORT")
	if err != nil {
		return fmt.Errorf("failed to parse %q env var: %w", "TEST_PORT", err)
	}

	c.MultiplexerHost = os.Getenv("MULTIPLEXER_HOST")

	c.MultiplexerPort, err = parseUint32("HTTP_PORT")
	if err != nil {
		return fmt.Errorf("failed to parse %q env var: %w", "HTTP_PORT", err)
	}

	c.MultiplexerBasePath = os.Getenv("HTTP_BASE_PATH")

	return nil
}

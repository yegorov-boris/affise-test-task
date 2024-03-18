package configs

import (
	"fmt"
	"os"
)

type ConfigTest struct {
	HTTPHost        string
	HTTPPort        uint32
	MultiplexerHost string
}

func (c *ConfigTest) Parse() error {
	var err error

	c.HTTPHost = os.Getenv("TEST_HOST")

	c.HTTPPort, err = parseUint32("TEST_PORT")
	if err != nil {
		return fmt.Errorf("failed to parse %q env var: %w", "TEST_PORT", err)
	}

	c.MultiplexerHost = os.Getenv("MULTIPLEXER_HOST")

	return nil
}

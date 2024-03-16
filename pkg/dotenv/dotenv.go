package dotenv

import (
	"fmt"
	"os"
	"strings"
)

/*
Load parses simplified .env format:
a line either is empty or has format key=value
where key and value are any string,
no interpolation,
no "newline" symbols in keys and values.
*/
func Load(name string) error {
	b, err := os.ReadFile(name)
	if err != nil {
		return fmt.Errorf("failed to read .env file: %w", err)
	}

	if err := parse(b); err != nil {
		return fmt.Errorf("failed to parse .env file: %w", err)
	}

	return nil
}

func parse(b []byte) error {
	for _, line := range strings.Split(string(b), "\n") {
		if len(line) == 0 {
			continue
		}

		pair := strings.SplitN(line, "=", 2)
		if len(pair) != 2 {
			return fmt.Errorf("unsupproted key-value pair %q in .env", line)
		}

		if err := os.Setenv(pair[0], pair[1]); err != nil {
			return fmt.Errorf("failed to set env var %q: %w", pair[0], err)
		}
	}

	return nil
}

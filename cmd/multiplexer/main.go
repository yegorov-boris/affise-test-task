package main

import (
	"log"
	"yegorov-boris/affise-test-task/configs"
	"yegorov-boris/affise-test-task/internal/multiplexer"
	"yegorov-boris/affise-test-task/pkg/dotenv"
)

func main() {
	if err := dotenv.Load(".env"); err != nil {
		log.Fatalf("failed to parse .env file: %s", err)
	}
	cfg := new(configs.Config)
	if err := cfg.Parse(); err != nil {
		log.Fatalf("failed to parse config: %s", err)
	}
	if err := cfg.Validate(); err != nil {
		log.Fatalf("failed to validate config: %s", err)
	}
	multiplexer.Run(cfg)
}

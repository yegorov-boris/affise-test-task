package main

import (
	"log"
	"os"
	"os/signal"
	"yegorov-boris/affise-test-task/configs"
	"yegorov-boris/affise-test-task/internal/multiplexer"
	"yegorov-boris/affise-test-task/pkg/dotenv"
)

func main() {
	if err := dotenv.Load(".env"); err != nil {
		log.Fatalf("failed to parse .env file: %s", err)
	}

	cfg, err := configs.New()
	if err != nil {
		log.Fatalf("failed to create config: %s", err)
	}

	shutdown, err := multiplexer.Run(cfg)
	if err != nil {
		log.Fatalf("failed to start multiplexer: %s", err)
	}

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	if err := shutdown(); err != nil {
		log.Fatalf("graceful shutdown failed: %s", err)
	}
}

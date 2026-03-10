package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"processor/internal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	setup, err := internal.NewSetup()
	if err != nil {
		log.Fatalf("Failed to initialize setup: %v", err)
	}

	err = setup.RunWorker(ctx)
	if err != nil {
		log.Fatalf("Failed to run worker: %v", err)
	}
}

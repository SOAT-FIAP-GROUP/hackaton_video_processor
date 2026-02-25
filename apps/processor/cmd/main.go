package main

import (
	"context"
	"log"
	"processor/internal"
)

func main() {
	setup, err := internal.NewSetup()
	if err != nil {
		log.Panicf("Failed to initialize setup: %v", err)
	}

	err = setup.RunWorker(context.Background())
	if err != nil {
		log.Panicf("Failed to run worker: %v", err)
	}
}

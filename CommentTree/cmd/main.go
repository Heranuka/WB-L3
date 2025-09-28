package main

import (
	"commentTree/internal/components"
	"commentTree/internal/config"
	"context"
	"log"
	"os"
	"sync"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		os.Exit(1)
	}

	wg := &sync.WaitGroup{}

	logger := components.SetupLogger()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	components, err := components.InitComponents(ctx, logger, cfg)
	if err != nil {
		log.Fatal("failed to init components")
	}

	wg.Add(1)

	go func() {
		defer wg.Done()
		if err := components.HttpServer.Run(ctx); err != nil {
			logger.Error("Failed to init components", "error", err.Error())
		}
	}()

	wg.Wait()

	if err := components.StopComponents(); err != nil {
		log.Println("failed to stop components")
	}

	logger.Info("The programm exited!")
}

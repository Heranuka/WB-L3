package main

import (
	"context"
	"log/slog"
	"sync"

	_ "wb-l3.7/docs"
	"wb-l3.7/internal/components"
	"wb-l3.7/internal/config"
)

func main() {
	logger := components.SetupLogger("local")
	config, err := config.LoadPath()
	if err != nil {
		logger.Error("error  with config in main", slog.String("error", err.Error()))
		return
	}
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	comps, err := components.InitComponents(ctx, logger, config)
	if err != nil {
		logger.Error("could not initComponents in main", slog.String("error", err.Error()))
		return
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := comps.HttpServer.Run(ctx); err != nil {
			logger.Error("error in main while running httpserver", slog.String("error", err.Error()))
		}
	}()

	wg.Wait()
}

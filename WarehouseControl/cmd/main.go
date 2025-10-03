package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wb-go/wbf/zlog"
	"golang.org/x/sync/errgroup"
	_ "wb-l3.7/docs"
	"wb-l3.7/internal/components"
	"wb-l3.7/internal/config"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	zlog.Init()

	logger := zlog.Logger

	appCtx, cancel := context.WithCancel(context.Background())
	eg, erctx := errgroup.WithContext(appCtx)

	components, err := components.InitComponents(erctx, logger, cfg)
	if err != nil {
		logger.Error().Err(err).Msg("Could not init components")
		os.Exit(1)
	}

	eg.Go(func() error {
		if err := components.HttpServer.Run(erctx); err != nil {
			if err.Error() == "http: Server closed" {
				// Это нормальное завершение сервера при Shutdown
				logger.Info().Msg("Http Server closed")
				return nil
			}
			logger.Error().Err(err).Msg("The http Server failed")
			return err
		}
		logger.Info().Msg("Http Server stopped")
		return nil
	})

	quitChan := make(chan os.Signal, 1)
	signal.Notify(quitChan, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-quitChan:
		logger.Info().Str("signal", sig.String()).Msg("Captured signal, initiating shutdown")
		cancel()
	case <-erctx.Done():
		logger.Info().Msg("Error group context cancelled")
		cancel()
	}

	// Ожидаем завершения рабочих горутин с таймаутом
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	doneChan := make(chan struct{})
	go func() {
		defer close(doneChan)
		if err = eg.Wait(); err != nil {
			logger.Error().Err(err).Msg("Application exited with error")
		} else {
			logger.Info().Msg("Application exited successfully")
		}
	}()

	select {
	case <-doneChan:
	case <-shutdownCtx.Done():
		logger.Warn().Msg("Forced shutdown after timeout")
	}

	if shutdownErr := components.Shutdown(); shutdownErr != nil {
		logger.Error().Err(shutdownErr).Msg("Error during component shutdown")
		os.Exit(1)
	}

	logger.Info().Msg("All components stopped gracefully. Flushing logs...")

	logger.Info().Msg("Graceful shutdown complete")
}

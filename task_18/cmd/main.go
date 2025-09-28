package main

import (
	"context"
	"fmt" // Для вывода ошибок
	"os"
	"os/signal"
	"sync"
	"syscall"
	"test_18/internal/components"
	"test_18/internal/config"
)

func main() {
	// 1. Загрузка конфигурации
	cfg, err := config.LoadPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1) // Здесь os.Exit приемлем, так как нет ничего, что нужно было бы корректно завершать
	}

	// 2. Инициализация логгера
	logger := components.SetupLogger(cfg.Env) // Предполагаем, что SetupLogger не возвращает ошибку, или она обрабатывается внутри

	// 3. Создание корневого контекста и функции отмены
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Важно: cancel() должен быть вызван, чтобы освободить ресурсы и остановить все горутины.

	// 4. Инициализация компонентов
	comp, err := components.InitComponents(ctx, *cfg, logger)
	if err != nil {
		logger.Error("Failed to initialize components", "error", err.Error())
		os.Exit(1) // Здесь os.Exit приемлем
	}

	// 5. Установка обработчика сигналов ОС
	// Создаем канал для получения сигналов прерывания (SIGINT, SIGTERM)
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	// 6. Запуск HTTP сервера в отдельной горутине
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("HTTP server starting...")
		if err := comp.HttpServer.Run(ctx); err != nil && err != context.Canceled {
			// Если ошибка не связана с отменой контекста (т.е. не штатное завершение),
			// логируем ее как ошибку.
			logger.Error("HTTP server stopped with error", "error", err)
			// В этом случае мы можем захотеть уведомить main, чтобы она тоже завершилась
			// Но здесь мы полагаемся на контекст, который будет отменен основным потоком.
		} else if err == context.Canceled {
			logger.Info("HTTP server stopped cleanly (context canceled).")
		}
	}()

	// 7. Ожидание сигнала остановки
	// main будет блокироваться здесь, пока не получит сигнал остановки
	<-stopChan
	logger.Info("Received stop signal, shutting down...")

	// 8. Отмена контекста
	// Это уведомит все горутины, использующие ctx, о необходимости завершения.
	cancel()

	// 9. Ожидание завершения всех горутин
	wg.Wait()
	logger.Info("All goroutines finished.")

	// 10. Корректное завершение компонентов
	if err := comp.ShutDownComponents(); err != nil {
		logger.Error("Error during shutdown", "error", err)
		os.Exit(1) // Здесь os.Exit приемлем, еслиshutdown не удался
	}

	logger.Info("Application shutdown complete.")
}

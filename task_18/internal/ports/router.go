package ports

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"test_18/internal/config"
	"test_18/internal/ports/rest"
	"test_18/internal/service"
	"test_18/pkg/e"

	"time"

	"github.com/gin-gonic/gin"
)

type Server struct {
	logger *slog.Logger
	server *http.Server
	cfg    *config.Config
}

func NewServer(ctx context.Context, cfg *config.Config, logger *slog.Logger, eventService service.Service) *Server {
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Http.Port),
		Handler:      InitRouter(ctx, logger, eventService),
		ReadTimeout:  cfg.Http.ReadTimeout,
		WriteTimeout: cfg.Http.WriteTimeout,
	}

	return &Server{
		logger: logger,
		server: server,
		cfg:    cfg,
	}
}

func InitRouter(ctx context.Context, logger *slog.Logger, eventService service.Service) *gin.Engine {
	r := gin.Default()
	h := rest.NewHandler(logger, &eventService)
	r.Use(LoggingMiddleware(logger))

	r.POST("/create_event", h.CreateEventHandler)
	r.PATCH("/update_event/:id", h.UpdateEventHandler)
	r.DELETE("/delete_event/:id", h.DeleteEventHandler)
	r.GET("/events_for_day", h.GetEventsForDayHandler)
	r.GET("/events_for_week", h.GetEventsForWeekHandler)
	r.GET("/events_for_month", h.GetEventsForMonthHandler)

	return r
}

func (s *Server) Run(ctx context.Context) error {
	errChan := make(chan error, 1)
	go func() {
		s.logger.Info("starting listening", slog.String("address", s.cfg.Http.Port))
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- e.Wrap("ports.router.go", err)
		} else {
			s.logger.Info("HTTP Server stopped gracefully")
			errChan <- nil
		}
	}()

	select {
	case <-ctx.Done():
		errChan <- ctx.Err()
		if err := s.StopHTTPServer(); err != nil {
			return err
		}
	case err := <-errChan:
		return err
	}

	return nil
}
func (s *Server) StopHTTPServer() error {
	shutDownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Http.ShutdownTimeout)
	defer cancel()
	if err := s.server.Shutdown(shutDownCtx); err != nil {
		return e.Wrap("failed to shutdown gracefully", err)
	}
	return nil
}

func LoggingMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		c.Next()
		duration := time.Since(startTime)

		logger.Info("HTTP request",
			slog.String("method", c.Request.Method),
			slog.String("uri", c.Request.RequestURI),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("latency", duration),
		)
	}
}

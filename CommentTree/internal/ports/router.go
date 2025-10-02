package ports

import (
	"commentTree/internal/config"
	"commentTree/internal/ports/rest/comments"
	render_handler "commentTree/internal/ports/rest/render"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/rs/zerolog"
	"github.com/wb-go/wbf/ginext"
)

type Server struct {
	logger zerolog.Logger
	server *http.Server
	cfg    *config.Config
}

func NewServer(ctx context.Context, logger zerolog.Logger, cfg *config.Config, service comments.CommentHandler, render render_handler.HandlerRender) *Server {
	httpHandler := comments.NewHandler(logger, cfg, service)
	rendandler := render_handler.NewHandler(logger, render)
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Http.Port),
		Handler:      InitRouter(ctx, logger, httpHandler, rendandler),
		ReadTimeout:  cfg.Http.ReadTimeout,
		WriteTimeout: cfg.Http.WriteTimeout,
	}
	return &Server{
		logger: logger,
		server: server,
		cfg:    cfg,
	}
}

func (s *Server) Run(ctx context.Context) error {
	errChan := make(chan error)

	s.logger.Info().Str("addr", s.server.Addr).Msg("Starting HTTP server")

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error().Err(err).Msg("HTTP server failed")
			errChan <- err
		} else {
			s.logger.Info().Msg("HTTP server stopped")
			errChan <- nil
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Http.ShutdownTimeout)
		defer cancel()

		s.logger.Info().Msg("Context done, initiating server shutdown")

		if err := s.server.Shutdown(shutdownCtx); err != nil {
			s.logger.Error().Err(err).Msg("HTTP server shutdown failed")
		} else {
			s.logger.Info().Msg("HTTP server shutdown completed")
		}
		return ctx.Err()

	case err := <-errChan:
		return err
	}
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}

func InitRouter(ctx context.Context, log zerolog.Logger, h *comments.Handler, ren *render_handler.Handler) *ginext.Engine {
	r := ginext.New()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:8080"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	config.AllowCredentials = true
	r.Use(cors.New(config))

	// Логгер и Recovery
	r.Use(ginext.Logger())
	r.Use(ginext.Recovery())

	// Маршруты
	r.GET("/home", ren.HomeHandler)
	r.POST("/comments", h.CreateHandler)
	r.GET("/comments", h.GetRootCommentsHandler)                      // корневые комментарии с пагинацией
	r.GET("/comments/:parent_id/children", h.GetChildCommentsHandler) // дочерние по parentId
	r.DELETE("/comments/:id", h.DeleteHandler)

	return r
}

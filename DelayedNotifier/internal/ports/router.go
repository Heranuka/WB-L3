package ports

import (
	"context"
	"delay/internal/config"
	"delay/internal/ports/noteAuth"
	"delay/internal/ports/renderH"
	"delay/internal/service"
	"delay/internal/service/render"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/rs/zerolog"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/wb-go/wbf/ginext"
)

type Server struct {
	logger zerolog.Logger
	server *ginext.Engine
	cfg    config.Config
}

func NewServer(ctx context.Context, cfg *config.Config, logger zerolog.Logger, service service.NotificationService, render render.Renderer) *Server {
	h := noteAuth.NewHandler(ctx, logger, service)
	rend := renderH.NewHandler(ctx, logger, render)
	r := InitRouter(ctx, h, rend, logger)
	return &Server{
		logger: logger,
		server: r,
		cfg:    *cfg,
	}
}

func InitRouter(ctx context.Context, h *noteAuth.Handler, rend *renderH.Handler, logger zerolog.Logger) *ginext.Engine {
	r := ginext.New()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:8080"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	config.AllowCredentials = true
	r.Use(cors.New(config))
	r.Use(ginext.Logger())
	r.Use(ginext.Recovery())
	r.GET("/", rend.HomeHandler)
	docsURL := ginSwagger.URL("http://localhost:8080/swagger/doc.json")
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, docsURL))
	notifyGroup := r.Group("/notify")
	{
		notifyGroup.POST("/create", h.CreateHanlder)
		notifyGroup.GET("/all", h.GetAllHandler)
		notifyGroup.GET("/status/:id", h.StatusHanlder)
		notifyGroup.DELETE("/cancel/:id", h.CancelHanlder)
	}

	return r
}

func (s *Server) Run(ctx context.Context) error {
	errChan := make(chan error, 1)
	srv := &http.Server{
		Addr:    ":" + s.cfg.Http.Port,
		Handler: s.server,
	}
	go func() {
		s.logger.Info().Str("port", s.cfg.Http.Port).Msg("Starting listening")
		err := srv.ListenAndServe()

		if err != nil && err != http.ErrServerClosed {
			s.logger.Error().Err(err).Msg("Server failed to start")
			errChan <- fmt.Errorf("ListenAndServe error: %w", err)
		} else {
			errChan <- nil
		}
	}()

	select {
	case <-ctx.Done():
		s.logger.Info().Str("reason", ctx.Err().Error()).Msg("Shutting down the server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			s.logger.Error().Err(err).Msg("Server forced to shutdown")
			return err
		}
		s.logger.Info().Msg("Server stopped gracefully")
		return nil
	case err := <-errChan:
		if err != nil {
			s.logger.Error().Err(err).Msg("HttpServer failed")
			return err
		}
		s.logger.Info().Msg("HttpServer stopped gracefully (without external signal).")
		return nil
	}
}

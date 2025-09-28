package ports

import (
	"commentTree/internal/config"
	rest "commentTree/internal/ports/rest/comments"
	render_handler "commentTree/internal/ports/rest/render"
	"commentTree/internal/service"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	log    *slog.Logger
	server *http.Server
	cfg    *config.Config
}

func NewServer(ctx context.Context, logger *slog.Logger, cfg *config.Config, service service.Service, render service.Render) *Server {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Http.Port),
		Handler: InitRouter(ctx, logger, service, render),
	}
	return &Server{
		log:    logger,
		server: server,
		cfg:    cfg,
	}
}

func (s *Server) Run(ctx context.Context) error {
	errResult := make(chan error, 1)
	go func() {
		s.log.Info("Starting listening:", slog.String("port", s.cfg.Http.Port))
		if err := s.server.ListenAndServe(); err != nil {
			errResult <- err
		}
	}()

	select {
	case <-ctx.Done():
		if err := s.Stop(); err != nil {
			errResult <- err
		}
	case err := <-errResult:
		return err
	}

	return nil
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}

func InitRouter(ctx context.Context, log *slog.Logger, service service.Service, render service.Render) *gin.Engine {
	h := rest.NewHandler(log, &service)
	ren := render_handler.NewHandler(log, &render)

	r := gin.Default()

	r.LoadHTMLGlob("templates/*.html")
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:8080"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	config.AllowCredentials = true

	r.Use(cors.New(config))
	r.GET("/home", ren.HomeHandler)
	r.POST("/comments", h.CreateHandler)
	r.GET("/comments/:id", h.GetByIdHandler)
	r.DELETE("/comments/:id", h.DeleteHandler)

	return r
}

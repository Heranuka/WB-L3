package ports

import (
	"context"
	"fmt"
	"net/http"
	"shortener/internal/config"

	"shortener/internal/ports/rest/linksAuth"
	"shortener/internal/ports/rest/render"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/rs/zerolog"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/wb-go/wbf/ginext"
)

type Server struct {
	logger zerolog.Logger
	server *http.Server
	cfg    *config.Config
}

func NewServer(ctx context.Context, cfg *config.Config, logger zerolog.Logger, urlservice linksAuth.ShortLinkService, clickService linksAuth.ClickService, serviceRender render.RenderHandler) *Server {
	handler := linksAuth.NewHandler(logger, urlservice, clickService)
	rend := render.NewHandler(serviceRender)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Http.Port),
		Handler: InitRouter(ctx, handler, rend),
	}
	return &Server{
		logger: logger,
		server: server,
		cfg:    cfg,
	}
}

func InitRouter(ctx context.Context, h *linksAuth.Handler, rend *render.Handler) *ginext.Engine {
	r := ginext.New()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:8080"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	config.AllowCredentials = true
	r.Use(cors.New(config))
	r.Use(ginext.Logger())
	r.Use(ginext.Recovery())

	r.GET("/", rend.Home) // Возвращает HTML через rend.Home
	docsURL := ginSwagger.URL("http://localhost:8080/swagger/doc.json")
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, docsURL))
	r.POST("/shorten", h.Create)
	r.GET("/s/:shortURL", h.Redirect)
	r.GET("/analytics/:shortURL", h.Analytics)

	return r
}

func (s *Server) Run(ctx context.Context) error {
	errResult := make(chan error)

	go func() {
		s.logger.Info().Msgf("starting listening: %s", s.server.Addr)
		errResult <- s.server.ListenAndServe()
	}()
	var err error
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err = <-errResult:
	}
	return err
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error().
			Err(err).
			Msg("handler.router.Stop: failed to shutdown httpServer")
		return fmt.Errorf("could not stop HTTP Server properly %v", err.Error())
	}
	return nil
}

package ports

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/rs/zerolog"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/wb-go/wbf/ginext"
	"wb-l3.7/internal/config"

	"wb-l3.7/internal/ports/rest/auth"
	"wb-l3.7/internal/ports/rest/items"
	"wb-l3.7/internal/ports/rest/renderer"
	"wb-l3.7/pkg/jwt"
)

type Server struct {
	logger zerolog.Logger
	server *http.Server
	cfg    *config.Config
}

func NewServer(
	config *config.Config,
	authService auth.ServiceAuth,
	itemsHistory items.HistoryStorage,
	itemsStorage items.ItemStorage,
	rendler renderer.RenderHandler,
	logger zerolog.Logger,
	manager jwt.TokenManager,
) (*Server, error) {
	logger.Info().Msg("Initializing HTTP handlers")

	httpHandler := auth.NewHandler(logger, authService)
	renderHandler := renderer.NewHandler(rendler)
	itemsHandler := items.NewItemsHandler(logger, itemsStorage, itemsHistory) // исправить!!!!

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", config.Http.Port),
		Handler:      InitRouter(httpHandler, itemsHandler, renderHandler, logger, manager),
		ReadTimeout:  config.Http.ReadTimeout,
		WriteTimeout: config.Http.WriteTimeout,
	}

	logger.Info().
		Str("addr", server.Addr).
		Msg("Server instance created")

	return &Server{
		logger: logger,
		server: server,
		cfg:    config,
	}, nil
}

func InitRouter(auth *auth.Handler, items *items.Handler, rend *renderer.Handler, logger zerolog.Logger, manager jwt.TokenManager) *ginext.Engine {
	r := ginext.New()

	r.LoadHTMLGlob("templates/*.html")
	docsURL := ginSwagger.URL("http://localhost:8080/swagger/doc.json")
	corsConfig := cors.DefaultConfig()
	// Вместо жёстко прописанного здесь, получаем AllowOrigins из конфига
	corsConfig.AllowOrigins = []string{"http://localhost:8080", "http://localhost:8080/home", "http://localhost:8080/login"} // например []string{"http://localhost:8080", "https://mydomain.com"}

	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	corsConfig.AllowCredentials = true

	r.Use(cors.New(corsConfig))

	r.Use(cors.New(corsConfig))

	logger.Info().Msg("Registering routes")

	r.GET("/", rend.RootRedirect)
	r.GET("/home", jwt.ValidateTokenMiddleware(manager), rend.Home)

	r.GET("/login", rend.Loginpage)
	r.GET("/register", rend.Registerpage)
	r.POST("/user/register", auth.Register)
	r.POST("/user/login", auth.Login)
	r.POST("/user/refresh", auth.RefreshToken)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, docsURL))

	itemsGroup := r.Group("/items")
	itemsGroup.Use(jwt.ValidateTokenMiddleware(manager))
	itemsGroup.Use(jwt.RequireRole(jwt.Admin, jwt.Manager))
	{
		itemsGroup.POST("/create", items.CreateItemHandler)
		itemsGroup.GET("/getall", items.GetItemsHandler)
		itemsGroup.GET("/history/:id", items.GetItemHistoryHandler)
		itemsGroup.PUT("/update/:id", items.UpdateItemHandler)
		itemsGroup.DELETE("/delete/:id", items.DeleteItemHandler)
	}

	viewerGroup := r.Group("/viewer")
	viewerGroup.Use(jwt.ValidateTokenMiddleware(manager))
	viewerGroup.Use(jwt.RequireRole(jwt.Viewer))

	profileGroup := r.Group("/profile")
	profileGroup.Use(jwt.ValidateTokenMiddleware(manager))
	{
		profileGroup.GET("/", items.GetUserProfileHandler)
	}

	logger.Info().
		Int("routes_count", 12).
		Msg("Routes registered successfully")

	return r
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

func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.Http.ShutdownTimeout)
	defer cancel()

	s.logger.Info().Msg("Stopping HTTP server")

	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error().Err(err).Msg("Failed to shutdown HTTP server")
	} else {
		s.logger.Info().Msg("HTTP server stopped gracefully")
	}
}

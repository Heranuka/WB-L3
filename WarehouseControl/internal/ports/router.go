package ports

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"wb-l3.7/internal/config"
	"wb-l3.7/internal/ports/rest/auth"
	"wb-l3.7/internal/ports/rest/items"
	"wb-l3.7/internal/service/render"
	"wb-l3.7/pkg/jwt"
)

type Server struct {
	logger *slog.Logger
	server *http.Server
	cfg    *config.Config
}

func NewServer(config *config.Config, authService auth.ServiceAuth, itemsHistory items.HistoryStorage, itemsStorage items.ItemStorage, serviceRender render.Render, logger *slog.Logger, manager jwt.TokenManager) (*Server, error) {
	httpHandler := auth.NewHandler(logger, authService, serviceRender)
	itemsHandler := items.NewItemsHandler(logger, itemsStorage, itemsHistory) // исправить!!!!

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", config.Http.Port),
		Handler:      InitRouter(httpHandler, itemsHandler, logger, manager),
		ReadTimeout:  config.Http.ReadTimeout,
		WriteTimeout: config.Http.WriteTimeout,
	}

	return &Server{
		logger: logger,
		server: server,
		cfg:    config,
	}, nil
}

func InitRouter(auth *auth.Handler, items *items.Handler, logger *slog.Logger, manager jwt.TokenManager) *gin.Engine {
	r := gin.Default()

	r.LoadHTMLGlob("templates/*.html")
	docsURL := ginSwagger.URL("http://localhost:8080/swagger/doc.json")
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:8080"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	config.AllowCredentials = true

	r.Use(cors.New(config))

	r.GET("/", auth.RootRedirect)
	r.GET("/login", auth.Loginpage)
	r.GET("/register", auth.Registerpage)
	r.POST("/user/register", auth.Register)    // Предполагается, что Register не требует аутентификации
	r.POST("/user/login", auth.Login)          // Предполагается, что Login не требует аутентификации
	r.POST("/user/refresh", auth.RefreshToken) // Предполагается, что Refresh не требует аутентификации
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, docsURL))

	// --- Маршруты для элементов, требующие роли ---
	itemsGroup := r.Group("/items")
	// Сначала применяем middleware для проверки токена
	itemsGroup.Use(jwt.ValidateTokenMiddleware(manager))
	// Затем применяем middleware для проверки роли
	itemsGroup.Use(jwt.RequireRole(jwt.Admin, jwt.Manager)) // Только admin или manager
	{
		// Обработчики, которые требуют наличия роли "admin" или "manager"
		itemsGroup.POST("/create", items.CreateItemHandler)
		itemsGroup.GET("/getall", items.GetItemsHandler)
		itemsGroup.PUT("/update/:id", items.UpdateItemHandler)
		itemsGroup.DELETE("/delete/:id", items.DeleteItemHandler)
	}

	// --- Маршруты, доступные только для viewer ---
	viewerGroup := r.Group("/viewer")
	viewerGroup.Use(jwt.ValidateTokenMiddleware(manager))
	viewerGroup.Use(jwt.RequireRole(jwt.Viewer)) // Только viewer
	{
		viewerGroup.GET("/home", auth.Homepage)
	}
	// --- Маршруты, доступные для всех аутентифицированных пользователей ---
	// (здесь применяется только ValidateTokenMiddleware)
	profileGroup := r.Group("/profile")
	profileGroup.Use(jwt.ValidateTokenMiddleware(manager))
	{
		profileGroup.GET("/", items.GetUserProfileHandler)
	}

	return r
}

func (s *Server) Run(ctx context.Context) error {
	errResult := make(chan error, 1)
	go func() {
		s.logger.Info(fmt.Sprintf("starting listening: %s", s.server.Addr))
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("failed to run HTTP Server", slog.String("error", err.Error()))
			errResult <- err
		} else {
			errResult <- nil
		}
		close(errResult)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.Http.ShutdownTimeout)
		defer cancel()
		if err := s.server.Shutdown(shutdownCtx); err != nil {
			s.logger.Error("failed to shutdown HTTP Server", slog.String("error", err.Error()))
		}
		return ctx.Err()
	case err := <-errResult:
		return err
	}
}

func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.Http.ShutdownTimeout)
	defer cancel()
	err := s.server.Shutdown(ctx)
	if err != nil {
		s.logger.Error("failed to shutdown HTTP Server", slog.String("error", err.Error()))
	}
}

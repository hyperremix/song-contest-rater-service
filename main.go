package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hyperremix/song-contest-rater-service/authz"
	"github.com/hyperremix/song-contest-rater-service/custommiddleware"
	"github.com/hyperremix/song-contest-rater-service/handler"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
)

func setupPool(ctx context.Context) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(os.Getenv("SONGCONTESTRATERSERVICE_DB_CONNECTION_STRING"))
	if err != nil {
		return nil, err
	}

	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute

	return pgxpool.NewWithConfig(ctx, config)
}

func main() {
	godotenv.Load(".env")
	e := echo.New()

	ctx := context.Background()

	connPool, err := setupPool(ctx)
	if err != nil {
		e.Logger.Fatal(err)
	}
	defer connPool.Close()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		connPool.Close()
	}()

	metricsGroup := e.Group("/metrics")
	metricsGroup.GET("", echoprometheus.NewHandler())

	mainGroup := e.Group("")
	mainGroup.Use(
		middleware.CORSWithConfig(middleware.CORSConfig{
			AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete, http.MethodOptions},
			AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, echo.HeaderCacheControl, echo.HeaderXRequestedWith},
		}),
		authz.RequestAuthorizer(connPool),
		custommiddleware.RequestLogger(),
		middleware.Recover(),
		middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
			LogStatus:  true,
			LogLatency: true,
			LogMethod:  true,
			LogURI:     true,
			LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
				logger := zerolog.Ctx(c.Request().Context())
				logger.Info().
					Int("status", v.Status).
					Dur("latency", v.Latency).
					Msg("response returned")

				return nil
			},
			HandleError: true,
		}),
		echoprometheus.NewMiddleware("service"),
	)

	handler.RegisterHandlerRoutes(mainGroup, connPool)

	e.Logger.Fatal(e.Start(":8080"))
}

package main

import (
	"context"
	"net/http"
	"os"

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

func main() {
	godotenv.Load(".env")
	e := echo.New()

	ctx := context.Background()

	connPool, err := pgxpool.New(ctx, os.Getenv("SONGCONTESTRATERSERVICE_DB_CONNECTION_STRING"))
	if err != nil {
		e.Logger.Fatal(err)
	}
	defer connPool.Close()

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

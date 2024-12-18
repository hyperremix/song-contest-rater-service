package main

import (
	"context"
	"net/http"
	"os"

	"github.com/hyperremix/song-contest-rater-service/authz"
	"github.com/hyperremix/song-contest-rater-service/handler"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
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

	logger := zerolog.New(os.Stdout)
	e.Use(
		middleware.CORSWithConfig(middleware.CORSConfig{
			AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete, http.MethodOptions},
			AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, echo.HeaderCacheControl, echo.HeaderXRequestedWith},
		}),
		authz.RequestAuthorizer(connPool),
		middleware.Recover(),
		middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
			LogStatus:  true,
			LogLatency: true,
			LogMethod:  true,
			LogURI:     true,
			LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
				logger.Info().
					Timestamp().
					Int("status", v.Status).
					Dur("latency", v.Latency).
					Str("method", v.Method).
					Str("URI", v.URI).
					Msg("request")

				return nil
			},
			HandleError: true,
		}),
	)

	handler.RegisterHandlerRoutes(e, connPool)

	e.Logger.Fatal(e.Start(":8080"))
}

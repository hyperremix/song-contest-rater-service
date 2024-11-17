package main

import (
	"context"
	"os"

	"github.com/hyperremix/song-contest-rater-service/authz"
	"github.com/hyperremix/song-contest-rater-service/environment"
	"github.com/hyperremix/song-contest-rater-service/handler"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal().Msgf("Error loading .env file: %s", err)
	}

	e := echo.New()

	ctx := context.Background()

	connPool, err := pgxpool.New(ctx, environment.DB_CONNECTION_STRING)
	if err != nil {
		e.Logger.Fatal(err)
	}
	defer connPool.Close()

	logger := zerolog.New(os.Stdout)
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
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
	}), authz.RequestAuthorizer(connPool))

	handler.RegisterHandlerRoutes(e, connPool)

	e.Logger.Fatal(e.Start("localhost:8080"))
}

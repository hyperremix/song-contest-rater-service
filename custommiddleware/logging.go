package custommiddleware

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

func IncomingRequestLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			correlationId := uuid.New()

			reqLogger := log.With().
				Str("method", req.Method).
				Str("uri", req.RequestURI).
				Str("correlationId", correlationId.String()).
				Logger()

			ctx := reqLogger.WithContext(req.Context())
			c.SetRequest(req.WithContext(ctx))

			reqLogger.Info().Msg("request started")

			return next(c)
		}
	}
}

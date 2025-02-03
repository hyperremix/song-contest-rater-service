package custommiddleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

func ErrorLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err != nil {
				if httpErr, ok := err.(*echo.HTTPError); ok {
					if httpErr.Code == http.StatusInternalServerError || httpErr.Code == http.StatusServiceUnavailable {
						log := zerolog.Ctx(c.Request().Context())
						log.Error().Err(err).Msg("error occurred")
					}
				}
			}
			return err
		}
	}
}

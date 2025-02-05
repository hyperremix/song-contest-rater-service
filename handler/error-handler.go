package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/hyperremix/song-contest-rater-service/mapper"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

func ErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	code := getCode(err, c)
	c.JSON(code, map[string]string{
		"code":    fmt.Sprintf("%d", code),
		"message": err.Error(),
	})
}

func getCode(err error, c echo.Context) int {
	log := zerolog.Ctx(c.Request().Context())

	if httpErr, ok := err.(*echo.HTTPError); ok {
		if httpErr.Code == http.StatusForbidden {
			log.Warn().Err(err).Msg("forbidden")
		}

		if httpErr.Code == http.StatusUnauthorized {
			log.Warn().Err(err).Msg("unauthorized")
		}
		return httpErr.Code
	}

	switch {
	case errors.Is(err, mapper.NewRequestBindingError(nil)):
		log.Warn().Err(err).Msg("bad request")
		return http.StatusBadRequest
	case errors.Is(err, mapper.NewResponseBindingError(nil)):
		log.Warn().Err(err).Msg("bad request")
		return http.StatusBadRequest
	case errors.Is(err, pgx.ErrNoRows):
		log.Warn().Err(err).Msg("not found")
		return http.StatusNotFound

	case errors.Is(err, pgx.ErrTxClosed),
		errors.Is(err, pgx.ErrTxCommitRollback):
		log.Error().Err(err).Msg("service unavailable")
		return http.StatusServiceUnavailable
	default:
		log.Error().Err(err).Msg("internal server error")
		return http.StatusInternalServerError
	}
}

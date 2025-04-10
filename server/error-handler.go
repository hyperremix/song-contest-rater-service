package server

import (
	"errors"
	"fmt"

	"connectrpc.com/connect"
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
	c.JSON(int(code), map[string]string{
		"code":    fmt.Sprintf("%d", code),
		"message": err.Error(),
	})
}

func getCode(err error, c echo.Context) connect.Code {
	log := zerolog.Ctx(c.Request().Context())

	if connectErr, ok := err.(*connect.Error); ok {
		if connectErr.Code() == connect.CodePermissionDenied {
			log.Warn().Err(err).Msg("forbidden")
		}

		if connectErr.Code() == connect.CodeUnauthenticated {
			log.Warn().Err(err).Msg("unauthorized")
		}

		return connectErr.Code()
	}

	switch {
	case errors.Is(err, mapper.NewRequestBindingError(nil)):
		log.Warn().Err(err).Msg("bad request")
		return connect.CodeFailedPrecondition
	case errors.Is(err, mapper.NewResponseBindingError(nil)):
		log.Warn().Err(err).Msg("bad request")
		return connect.CodeFailedPrecondition
	case errors.Is(err, pgx.ErrNoRows):
		log.Warn().Err(err).Msg("not found")
		return connect.CodeNotFound

	case errors.Is(err, pgx.ErrTxClosed),
		errors.Is(err, pgx.ErrTxCommitRollback):
		log.Error().Err(err).Msg("service unavailable")
		return connect.CodeUnavailable
	default:
		log.Error().Err(err).Msg("internal server error")
		return connect.CodeInternal
	}
}

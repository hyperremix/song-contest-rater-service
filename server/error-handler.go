package server

import (
	"errors"

	"connectrpc.com/connect"
	"github.com/hyperremix/song-contest-rater-service/mapper"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

// ErrorHandler handles errors for ConnectRPC requests
// ConnectRPC handles its own error serialization, so we just need to log errors
func ErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	log := zerolog.Ctx(c.Request().Context())

	// Log the error with appropriate level based on error type
	if connectErr, ok := err.(*connect.Error); ok {
		switch connectErr.Code() {
		case connect.CodePermissionDenied:
			log.Warn().Err(err).Msg("forbidden")
		case connect.CodeUnauthenticated:
			log.Warn().Err(err).Msg("unauthorized")
		case connect.CodeNotFound:
			log.Warn().Err(err).Msg("not found")
		case connect.CodeInvalidArgument:
			log.Warn().Err(err).Msg("invalid argument")
		case connect.CodeFailedPrecondition:
			log.Warn().Err(err).Msg("failed precondition")
		case connect.CodeUnavailable:
			log.Error().Err(err).Msg("service unavailable")
		default:
			log.Error().Err(err).Msg("internal server error")
		}
	} else {
		// Handle non-ConnectRPC errors
		switch {
		case errors.Is(err, mapper.NewRequestBindingError(nil)):
			log.Warn().Err(err).Msg("bad request")
		case errors.Is(err, mapper.NewResponseBindingError(nil)):
			log.Warn().Err(err).Msg("bad request")
		case errors.Is(err, pgx.ErrNoRows):
			log.Warn().Err(err).Msg("not found")
		case errors.Is(err, pgx.ErrTxClosed),
			errors.Is(err, pgx.ErrTxCommitRollback):
			log.Error().Err(err).Msg("service unavailable")
		default:
			log.Error().Err(err).Msg("internal server error")
		}
	}

	// Let ConnectRPC handle the error response
	// Don't manually set HTTP status codes or JSON responses
}

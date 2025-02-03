package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/hyperremix/song-contest-rater-service/mapper"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
)

func ErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	code := getCode(err)
	c.JSON(code, map[string]string{
		"code":    fmt.Sprintf("%d", code),
		"message": err.Error(),
	})
}

func getCode(err error) int {
	if httpErr, ok := err.(*echo.HTTPError); ok {
		return httpErr.Code
	}

	switch {
	case errors.Is(err, mapper.NewRequestBindingError(nil)):
		return http.StatusBadRequest
	case errors.Is(err, mapper.NewResponseBindingError(nil)):
		return http.StatusBadRequest
	case errors.Is(err, pgx.ErrNoRows):
		return http.StatusNotFound

	case errors.Is(err, pgx.ErrTxClosed),
		errors.Is(err, pgx.ErrTxCommitRollback):
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

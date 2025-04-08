package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/hyperremix/song-contest-rater-service/mapper"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func TestErrorHandler(t *testing.T) {
	tests := []struct {
		name              string
		err               error
		responseCommitted bool
		expectedCode      int
		expectedMsg       string
	}{
		{
			name:              "Already committed response",
			err:               errors.New("some error"),
			responseCommitted: true,
			expectedCode:      http.StatusOK,
			expectedMsg:       "",
		},
		{
			name:              "HTTP error",
			err:               echo.NewHTTPError(http.StatusBadRequest, "bad request"),
			responseCommitted: false,
			expectedCode:      http.StatusBadRequest,
			expectedMsg:       "code=400, message=bad request",
		},
		{
			name:              "Request binding error",
			err:               mapper.NewRequestBindingError(errors.New("binding failed")),
			responseCommitted: false,
			expectedCode:      http.StatusBadRequest,
			expectedMsg:       "could not bind request",
		},
		{
			name:              "Response binding error",
			err:               mapper.NewResponseBindingError(errors.New("binding failed")),
			responseCommitted: false,
			expectedCode:      http.StatusBadRequest,
			expectedMsg:       "could not bind response",
		},
		{
			name:              "No rows error",
			err:               pgx.ErrNoRows,
			responseCommitted: false,
			expectedCode:      http.StatusNotFound,
			expectedMsg:       "no rows in result set",
		},
		{
			name:              "Transaction closed error",
			err:               pgx.ErrTxClosed,
			responseCommitted: false,
			expectedCode:      http.StatusServiceUnavailable,
			expectedMsg:       "tx is closed",
		},
		{
			name:              "Unknown error",
			err:               errors.New("unknown error"),
			responseCommitted: false,
			expectedCode:      http.StatusInternalServerError,
			expectedMsg:       "unknown error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if tt.responseCommitted {
				c.NoContent(http.StatusOK)
			}

			ErrorHandler(tt.err, c)

			if tt.responseCommitted {
				assert.Equal(t, tt.expectedCode, rec.Code)
				assert.Empty(t, rec.Body.String())
				return
			}

			assert.Equal(t, tt.expectedCode, rec.Code)

			var response errorResponse
			err := json.Unmarshal(rec.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, strconv.Itoa(tt.expectedCode), response.Code)
			assert.Equal(t, tt.expectedMsg, response.Message)
		})
	}
}

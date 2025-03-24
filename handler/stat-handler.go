package handler

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/hyperremix/song-contest-rater-service/authz"
	"github.com/hyperremix/song-contest-rater-service/db"
	"github.com/hyperremix/song-contest-rater-service/mapper"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type StatHandler struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

func NewStatHandler(pool *pgxpool.Pool) *StatHandler {
	return &StatHandler{
		queries: db.New(pool),
		pool:    pool,
	}
}

func registerStatRoutes(e *echo.Group, connPool *pgxpool.Pool) {
	h := NewStatHandler(connPool)

	e.GET("/stats/users", h.listUserStats)
	e.GET("/stats/users/me", h.getUserStats)
	e.GET("/stats/global", h.getGlobalStats)
}

func (h *StatHandler) listUserStats(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()

	usersStats, err := h.queries.ListUserStats(ctx)
	if err != nil {
		return err
	}

	globalStats, err := h.queries.GetGlobalStats(ctx)
	if err != nil {
		return err
	}

	users, err := h.queries.ListUsers(ctx)
	if err != nil {
		return err
	}

	response, err := mapper.FromDbUserStatListToResponse(usersStats, globalStats, users)
	if err != nil {
		return err
	}

	return echoCtx.JSON(http.StatusOK, response)
}

func (h *StatHandler) getUserStats(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()
	authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)
	userId, err := mapper.FromProtoToDbId(authUser.UserID)
	if err != nil {
		return err
	}

	userStats, err := h.queries.GetStatsByUserId(ctx, userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return echoCtx.JSON(http.StatusOK, mapper.EmptyUserStatsResponse())
		}
		return err
	}

	globalStats, err := h.queries.GetGlobalStats(ctx)
	if err != nil {
		return err
	}

	response, err := mapper.FromDbUserStatsToResponse(userStats, globalStats, &authUser.DbUser)
	if err != nil {
		return err
	}

	return echoCtx.JSON(http.StatusOK, response)
}

func (h *StatHandler) getGlobalStats(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()

	globalStats, err := h.queries.GetGlobalStats(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return echoCtx.JSON(http.StatusOK, mapper.EmptyGlobalStatsResponse())
		}
		return err
	}

	response, err := mapper.FromDbGlobalStatsToResponse(globalStats)
	if err != nil {
		return err
	}

	return echoCtx.JSON(http.StatusOK, response)
}

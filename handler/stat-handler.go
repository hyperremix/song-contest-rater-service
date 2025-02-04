package handler

import (
	"net/http"

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

	e.GET("/stats/users/:id", h.getUserStats)
	e.GET("/stats/global", h.getGlobalStats)
}

func (h *StatHandler) getUserStats(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()

	id := echoCtx.Param("id")
	userId, err := mapper.FromProtoToDbId(id)
	if err != nil {
		return err
	}

	userStats, err := h.queries.GetStatsByUserId(ctx, userId)
	if err != nil {
		return err
	}

	globalStats, err := h.queries.GetGlobalStats(ctx)
	if err != nil {
		return err
	}

	response, err := mapper.FromDbUserStatsToResponse(userStats, globalStats)
	if err != nil {
		return err
	}

	return echoCtx.JSON(http.StatusOK, response)
}

func (h *StatHandler) getGlobalStats(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()

	globalStats, err := h.queries.GetGlobalStats(ctx)
	if err != nil {
		return err
	}

	response, err := mapper.FromDbGlobalStatsToResponse(globalStats)
	if err != nil {
		return err
	}

	return echoCtx.JSON(http.StatusOK, response)
}

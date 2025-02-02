package handler

import (
	"github.com/hyperremix/song-contest-rater-service/sse"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type singleObjectRequest struct {
	Id string `param:"id"`
}

func RegisterHandlerRoutes(e *echo.Group, connPool *pgxpool.Pool) {
	registerActRoutes(e, connPool)
	registerCompetitionRoutes(e, connPool)
	registerRatingRoutes(e, connPool)
	registerUserRoutes(e, connPool)
	registerParticipationRoutes(e, connPool)
}

var broker = sse.NewBroker()

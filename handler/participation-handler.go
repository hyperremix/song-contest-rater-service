package handler

import (
	"net/http"

	"github.com/hyperremix/song-contest-rater-service/authz"
	"github.com/hyperremix/song-contest-rater-service/db"
	"github.com/hyperremix/song-contest-rater-service/mapper"
	"github.com/hyperremix/song-contest-rater-service/permission"
	pb "github.com/hyperremix/song-contest-rater-service/protos/songcontestrater"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type ParticipationHandler struct {
	queries  *db.Queries
	connPool *pgxpool.Pool
}

func NewParticipationHandler(connPool *pgxpool.Pool) *ParticipationHandler {
	return &ParticipationHandler{
		queries:  db.New(connPool),
		connPool: connPool,
	}
}

func registerParticipationRoutes(e *echo.Group, connPool *pgxpool.Pool) {
	h := NewParticipationHandler(connPool)
	e.POST("/participations", h.createParticipation)
	e.DELETE("/participations", h.deleteParticipation)
}

func (h *ParticipationHandler) createParticipation(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()
	authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)
	if err := authUser.CheckHasPermission(permission.WriteParticipations); err != nil {
		return err
	}

	var request pb.CreateParticipationRequest
	if err := echoCtx.Bind(&request); err != nil {
		return err
	}

	insertParams, err := mapper.FromCreateRequestToInsertCompetitionAct(&request)
	if err != nil {
		return err
	}

	err = h.queries.InsertCompetitionAct(ctx, insertParams)
	if err != nil {
		return err
	}

	return echoCtx.NoContent(http.StatusCreated)
}

func (h *ParticipationHandler) deleteParticipation(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()
	authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)
	if err := authUser.CheckHasPermission(permission.WriteParticipations); err != nil {
		return err
	}

	var request mapper.DeleteParticipationRequest
	if err := echoCtx.Bind(&request); err != nil {
		return err
	}

	deleteParams, err := mapper.FromDeleteRequestToDeleteCompetitionAct(&request)
	if err != nil {
		return err
	}

	err = h.queries.DeleteCompetitionAct(ctx, deleteParams)
	if err != nil {
		return err
	}

	return echoCtx.NoContent(http.StatusNoContent)
}

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
	"github.com/rs/zerolog"
)

type ActHandler struct {
	queries  *db.Queries
	connPool *pgxpool.Pool
}

func NewActHandler(connPool *pgxpool.Pool) *ActHandler {
	return &ActHandler{
		queries:  db.New(connPool),
		connPool: connPool,
	}
}

func registerActRoutes(e *echo.Group, connPool *pgxpool.Pool) {
	h := NewActHandler(connPool)

	e.GET("/acts", h.listActs)
	e.GET("/acts/:id", h.getAct)
	e.GET("/competitions/:competitionId/acts/:id", h.getCompetitionAct)
	e.POST("/acts", h.createAct)
	e.PUT("/acts/:id", h.updateAct)
	e.DELETE("/acts/:id", h.deleteAct)
}

func (h *ActHandler) listActs(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()
	log := zerolog.Ctx(ctx)

	acts, err := h.queries.ListActs(ctx)
	if err != nil {
		log.Error().Err(err).Msg("could not list acts")
		return err
	}

	response, err := mapper.FromDbActListToResponse(acts, make([]db.Rating, 0), make([]db.User, 0))
	if err != nil {
		log.Error().Err(err).Msg("could not map to response")
		return err
	}

	return echoCtx.JSON(http.StatusOK, response)
}

func (h *ActHandler) getAct(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()
	log := zerolog.Ctx(ctx)

	var request singleObjectRequest
	if err := echoCtx.Bind(&request); err != nil {
		log.Error().Err(err).Msg("could not bind request")
		return echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
	}

	id, err := mapper.FromProtoToDbId(request.Id)
	if err != nil {
		log.Error().Err(err).Msg("could not bind request")
		return echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
	}

	act, err := h.queries.GetActById(ctx, id)
	if err != nil {
		log.Error().Err(err).Msg("could not get act")
		return err
	}

	ratings, err := h.queries.ListRatingsByActId(ctx, act.ID)
	if err != nil {
		log.Error().Err(err).Msg("could not get ratings")
		return err
	}

	users, err := h.queries.ListUsersByActId(ctx, act.ID)
	if err != nil {
		log.Error().Err(err).Msg("could not get users")
		return err
	}

	response, err := mapper.FromDbActToResponse(act, ratings, users)
	if err != nil {
		log.Error().Err(err).Msg("could not map to response")
		return err
	}

	return echoCtx.JSON(http.StatusOK, response)
}

type competitionActRequest struct {
	CompetitionId string `param:"competitionId"`
	Id            string `param:"id"`
}

func (h *ActHandler) getCompetitionAct(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()
	log := zerolog.Ctx(ctx)

	var request competitionActRequest
	if err := echoCtx.Bind(&request); err != nil {
		log.Error().Err(err).Msg("could not bind request")
		return echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
	}

	id, err := mapper.FromProtoToDbId(request.Id)
	if err != nil {
		log.Error().Err(err).Msg("could not bind request")
		return echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
	}

	competitionId, err := mapper.FromProtoToDbId(request.CompetitionId)
	if err != nil {
		log.Error().Err(err).Msg("could not bind request")
		return echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
	}

	act, err := h.queries.GetActById(ctx, id)
	if err != nil {
		log.Error().Err(err).Msg("could not get act")
		return err
	}

	ratings, err := h.queries.ListRatingsByCompetitionAndAcId(ctx, db.ListRatingsByCompetitionAndAcIdParams{CompetitionID: competitionId, ActID: act.ID})
	if err != nil {
		log.Error().Err(err).Msg("could not get ratings")
		return err
	}

	users, err := h.queries.ListUsersByActId(ctx, act.ID)
	if err != nil {
		log.Error().Err(err).Msg("could not get users")
		return err
	}

	response, err := mapper.FromDbActToResponse(act, ratings, users)
	if err != nil {
		log.Error().Err(err).Msg("could not map to response")
		return err
	}

	return echoCtx.JSON(http.StatusOK, response)
}

func (h *ActHandler) createAct(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()
	log := zerolog.Ctx(ctx)
	authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)
	if err := authUser.CheckHasPermission(permission.WriteActs); err != nil {
		return err
	}

	var request pb.CreateActRequest
	if err := echoCtx.Bind(&request); err != nil {
		log.Error().Err(err).Msg("could not bind request")
		return echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
	}

	params, err := mapper.FromCreateRequestToInsertAct(&request)
	if err != nil {
		log.Error().Err(err).Msg("could not map request to insert params")
		return echo.NewHTTPError(http.StatusBadRequest, "could not map request to insert params")
	}

	act, err := h.queries.InsertAct(ctx, params)
	if err != nil {
		log.Error().Err(err).Msg("could not insert act")
		return err
	}

	response, err := mapper.FromDbActToResponse(act, make([]db.Rating, 0), make([]db.User, 0))
	if err != nil {
		log.Error().Err(err).Msg("could not map to response")
		return err
	}

	return echoCtx.JSON(http.StatusCreated, response)
}

func (h *ActHandler) updateAct(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()
	log := zerolog.Ctx(ctx)
	authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)
	if err := authUser.CheckHasPermission(permission.WriteActs); err != nil {
		log.Error().Err(err).Msg("user does not have permission to write acts")
		return err
	}

	var request pb.UpdateActRequest
	if err := echoCtx.Bind(&request); err != nil {
		log.Error().Err(err).Msg("could not bind request")
		return echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
	}

	paramId := echoCtx.Param("id")

	if request.Id != paramId {
		log.Error().Msg("id in request does not match id in path")
		return echo.NewHTTPError(http.StatusBadRequest, "id in request does not match id in path")
	}

	updateParams, err := mapper.FromUpdateRequestToUpdateAct(&request)
	if err != nil {
		log.Error().Err(err).Msg("could not map request to update params")
		return echo.NewHTTPError(http.StatusBadRequest, "could not map request to update params")
	}

	act, err := h.queries.UpdateAct(ctx, updateParams)
	if err != nil {
		log.Error().Err(err).Msg("could not update act")
		return err
	}

	response, err := mapper.FromDbActToResponse(act, make([]db.Rating, 0), make([]db.User, 0))
	if err != nil {
		log.Error().Err(err).Msg("could not map to response")
		return err
	}

	return echoCtx.JSON(http.StatusOK, response)
}

func (h *ActHandler) deleteAct(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()
	log := zerolog.Ctx(ctx)
	authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)
	if err := authUser.CheckHasPermission(permission.WriteActs); err != nil {
		log.Error().Err(err).Msg("user does not have permission to write acts")
		return err
	}

	var request singleObjectRequest
	if err := echoCtx.Bind(&request); err != nil {
		log.Error().Err(err).Msg("could not bind request")
		return echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
	}

	id, err := mapper.FromProtoToDbId(request.Id)
	if err != nil {
		log.Error().Err(err).Msg("could not bind request")
		return echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
	}

	act, err := h.queries.DeleteActById(ctx, id)
	if err != nil {
		log.Error().Err(err).Msg("could not delete act")
		return err
	}

	response, err := mapper.FromDbActToResponse(act, make([]db.Rating, 0), make([]db.User, 0))
	if err != nil {
		log.Error().Err(err).Msg("could not map to response")
		return err
	}

	return echoCtx.JSON(http.StatusOK, response)
}

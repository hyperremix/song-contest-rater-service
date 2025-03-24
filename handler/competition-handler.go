package handler

import (
	"net/http"

	"github.com/hyperremix/song-contest-rater-service/authz"
	"github.com/hyperremix/song-contest-rater-service/db"
	"github.com/hyperremix/song-contest-rater-service/mapper"
	pb "github.com/hyperremix/song-contest-rater-service/protos/songcontestrater"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type CompetitionHandler struct {
	queries  *db.Queries
	connPool *pgxpool.Pool
}

func NewCompetitionHandler(connPool *pgxpool.Pool) *CompetitionHandler {
	return &CompetitionHandler{
		queries:  db.New(connPool),
		connPool: connPool,
	}
}

func registerCompetitionRoutes(e *echo.Group, connPool *pgxpool.Pool) {
	h := NewCompetitionHandler(connPool)

	e.GET("/competitions", h.listCompetitions)
	e.GET("/competitions/:id", h.getCompetition)
	e.POST("/competitions", h.createCompetition)
	e.PUT("/competitions/:id", h.updateCompetition)
	e.DELETE("/competitions/:id", h.deleteCompetition)
}

func (h *CompetitionHandler) listCompetitions(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()
	competitions, err := h.queries.ListCompetitions(ctx)
	if err != nil {
		return err
	}

	response, err := mapper.FromDbCompetitionListToResponse(competitions)
	if err != nil {
		return err
	}

	return echoCtx.JSON(http.StatusOK, response)
}

func (h *CompetitionHandler) getCompetition(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()

	var request singleObjectRequest
	if err := echoCtx.Bind(&request); err != nil {
		return err
	}

	id, err := mapper.FromProtoToDbId(request.Id)
	if err != nil {
		return err
	}

	competition, err := h.queries.GetCompetitionById(ctx, id)
	if err != nil {
		return err
	}

	ratings, err := h.queries.ListRatingsByCompetitionId(ctx, competition.ID)
	if err != nil {
		return err
	}

	acts, err := h.queries.ListActsByCompetitionId(ctx, competition.ID)
	if err != nil {
		return err
	}

	users, err := h.queries.ListUsersByCompetitionId(ctx, competition.ID)
	if err != nil {
		return err
	}

	response, err := mapper.FromDbToCompetitionWithActsAndUsersResponse(competition, ratings, acts, users)
	if err != nil {
		return err
	}

	return echoCtx.JSON(http.StatusOK, response)
}

func (h *CompetitionHandler) createCompetition(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()
	authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)
	if err := authUser.CheckIsAdmin(); err != nil {
		return err
	}

	var request pb.CreateCompetitionRequest
	if err := echoCtx.Bind(&request); err != nil {
		return err
	}

	insertParams := mapper.FromCreateRequestToInsertCompetition(&request)

	competition, err := h.queries.InsertCompetition(ctx, insertParams)
	if err != nil {
		return err
	}

	response, err := mapper.FromDbCompetitionToResponse(competition)
	if err != nil {
		return err
	}

	return echoCtx.JSON(http.StatusCreated, response)
}

func (h *CompetitionHandler) updateCompetition(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()
	authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)
	if err := authUser.CheckIsAdmin(); err != nil {
		return err
	}

	var request pb.UpdateCompetitionRequest
	if err := echoCtx.Bind(&request); err != nil {
		return err
	}

	paramId := echoCtx.Param("id")

	if request.Id != paramId {
		return echo.NewHTTPError(http.StatusBadRequest, "id in request does not match id in path")
	}

	updateParams, err := mapper.FromUpdateRequestToUpdateCompetition(&request)
	if err != nil {
		return err
	}

	competition, err := h.queries.UpdateCompetition(ctx, updateParams)
	if err != nil {
		return err
	}

	response, err := mapper.FromDbCompetitionToResponse(competition)
	if err != nil {
		return err
	}

	return echoCtx.JSON(http.StatusOK, response)
}

func (h *CompetitionHandler) deleteCompetition(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()
	authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)
	if err := authUser.CheckIsAdmin(); err != nil {
		return err
	}

	var request singleObjectRequest
	if err := echoCtx.Bind(&request); err != nil {
		return err
	}

	id, err := mapper.FromProtoToDbId(request.Id)
	if err != nil {
		return err
	}

	competition, err := h.queries.DeleteCompetitionById(ctx, id)
	if err != nil {
		return err
	}

	response, err := mapper.FromDbCompetitionToResponse(competition)
	if err != nil {
		return err
	}

	return echoCtx.JSON(http.StatusOK, response)
}

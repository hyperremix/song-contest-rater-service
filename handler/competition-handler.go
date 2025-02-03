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
	e.POST("/competitions", createCompetition(connPool))
	e.PUT("/competitions/:id", updateCompetition(connPool))
	e.DELETE("/competitions/:id", deleteCompetition(connPool))
}

func (h *CompetitionHandler) listCompetitions(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()
	log := zerolog.Ctx(ctx)

	competitions, err := h.queries.ListCompetitions(ctx)
	if err != nil {
		log.Error().Err(err).Msg("could not list competitions")
		return err
	}

	response, err := mapper.FromDbCompetitionListToResponse(competitions)
	if err != nil {
		log.Error().Err(err).Msg("could not map to response")
		return err
	}

	return echoCtx.JSON(http.StatusOK, response)
}

func (h *CompetitionHandler) getCompetition(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()
	log := zerolog.Ctx(ctx)

	var request singleObjectRequest
	if err := echoCtx.Bind(&request); err != nil {
		log.Error().Err(err).Msg("could not bind request")
		return echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
	}

	id, err := mapper.FromProtoToDbId(request.Id)
	if err != nil {
		log.Error().Err(err).Msg("could not map id")
		return echo.NewHTTPError(http.StatusBadRequest, "could not map id")
	}

	competition, err := h.queries.GetCompetitionById(ctx, id)
	if err != nil {
		log.Error().Err(err).Msg("could not get competition")
		return err
	}

	ratings, err := h.queries.ListRatingsByCompetitionId(ctx, competition.ID)
	if err != nil {
		log.Error().Err(err).Msg("could not get ratings")
		return err
	}

	acts, err := h.queries.ListActsByCompetitionId(ctx, competition.ID)
	if err != nil {
		log.Error().Err(err).Msg("could not get acts")
		return err
	}

	users, err := h.queries.ListUsersByCompetitionId(ctx, competition.ID)
	if err != nil {
		log.Error().Err(err).Msg("could not get users")
		return err
	}

	response, err := mapper.FromDbToCompetitionWithActsAndUsersResponse(competition, ratings, acts, users)
	if err != nil {
		log.Error().Err(err).Msg("could not map to response")
		return err
	}

	return echoCtx.JSON(http.StatusOK, response)
}

func createCompetition(connPool *pgxpool.Pool) echo.HandlerFunc {
	return func(echoCtx echo.Context) error {
		ctx := echoCtx.Request().Context()
		log := zerolog.Ctx(ctx)
		authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)
		if err := authUser.CheckHasPermission(permission.WriteCompetitions); err != nil {
			log.Error().Err(err).Msg("user does not have permission to write competitions")
			return err
		}

		conn, err := connPool.Acquire(ctx)
		if err != nil {
			log.Error().Err(err).Msg("could not acquire connection")
			return err
		}
		defer conn.Release()

		queries := db.New(conn)

		var request pb.CreateCompetitionRequest
		if err := echoCtx.Bind(&request); err != nil {
			log.Error().Err(err).Msg("could not bind request")
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		insertParams := mapper.FromCreateRequestToInsertCompetition(&request)

		competition, err := queries.InsertCompetition(ctx, insertParams)
		if err != nil {
			log.Error().Err(err).Msg("could not insert competition")
			return err
		}

		response, err := mapper.FromDbCompetitionToResponse(competition)
		if err != nil {
			log.Error().Err(err).Msg("could not map to response")
			return err
		}

		return echoCtx.JSON(http.StatusCreated, response)
	}
}

func updateCompetition(connPool *pgxpool.Pool) echo.HandlerFunc {
	return func(echoCtx echo.Context) error {
		ctx := echoCtx.Request().Context()
		log := zerolog.Ctx(ctx)
		authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)
		if err := authUser.CheckHasPermission(permission.WriteCompetitions); err != nil {
			log.Error().Err(err).Msg("user does not have permission to write competitions")
			return err
		}

		conn, err := connPool.Acquire(ctx)
		if err != nil {
			log.Error().Err(err).Msg("could not acquire connection")
			return err
		}
		defer conn.Release()

		queries := db.New(conn)

		var request pb.UpdateCompetitionRequest
		if err := echoCtx.Bind(&request); err != nil {
			log.Error().Err(err).Msg("could not bind request")
			return echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
		}

		paramId := echoCtx.Param("id")

		if request.Id != paramId {
			log.Error().Msg("id in request does not match id in path")
			return echo.NewHTTPError(http.StatusBadRequest, "id in request does not match id in path")
		}

		updateParams, err := mapper.FromUpdateRequestToUpdateCompetition(&request)
		if err != nil {
			log.Error().Err(err).Msg("could not map request to update params")
			return echo.NewHTTPError(http.StatusBadRequest, "could not map request to update params")
		}

		competition, err := queries.UpdateCompetition(ctx, updateParams)
		if err != nil {
			log.Error().Err(err).Msg("could not update competition")
			return err
		}

		response, err := mapper.FromDbCompetitionToResponse(competition)
		if err != nil {
			log.Error().Err(err).Msg("could not map to response")
			return err
		}

		return echoCtx.JSON(http.StatusOK, response)
	}
}

func deleteCompetition(connPool *pgxpool.Pool) echo.HandlerFunc {
	return func(echoCtx echo.Context) error {
		ctx := echoCtx.Request().Context()
		log := zerolog.Ctx(ctx)
		authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)
		if err := authUser.CheckHasPermission(permission.WriteCompetitions); err != nil {
			log.Error().Err(err).Msg("user does not have permission to write competitions")
			return err
		}

		conn, err := connPool.Acquire(ctx)
		if err != nil {
			log.Error().Err(err).Msg("could not acquire connection")
			return err
		}
		defer conn.Release()

		queries := db.New(conn)

		var request singleObjectRequest
		if err := echoCtx.Bind(&request); err != nil {
			log.Error().Err(err).Msg("could not bind request")
			return echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
		}

		id, err := mapper.FromProtoToDbId(request.Id)
		if err != nil {
			log.Error().Err(err).Msg("could not map id")
			return echo.NewHTTPError(http.StatusBadRequest, "could not map id")
		}

		competition, err := queries.DeleteCompetitionById(ctx, id)
		if err != nil {
			log.Error().Err(err).Msg("could not delete competition")
			return err
		}

		response, err := mapper.FromDbCompetitionToResponse(competition)
		if err != nil {
			log.Error().Err(err).Msg("could not map to response")
			return err
		}

		return echoCtx.JSON(http.StatusOK, response)
	}
}

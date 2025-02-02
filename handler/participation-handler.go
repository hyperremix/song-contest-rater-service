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

func registerParticipationRoutes(e *echo.Echo, connPool *pgxpool.Pool) {
	e.POST("/participations", createParticipation(connPool))
	e.DELETE("/participations", deleteParticipation(connPool))
}

func createParticipation(connPool *pgxpool.Pool) echo.HandlerFunc {
	return func(echoCtx echo.Context) error {
		ctx := echoCtx.Request().Context()
		log := zerolog.Ctx(ctx)
		authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)
		if err := authUser.CheckHasPermission(permission.WriteParticipations); err != nil {
			log.Error().Err(err).Msg("user does not have permission to write participations")
			return err
		}

		conn, err := connPool.Acquire(ctx)
		if err != nil {
			log.Error().Err(err).Msg("could not acquire connection")
			return err
		}
		defer conn.Release()

		queries := db.New(conn)

		var request pb.CreateParticipationRequest
		if err := echoCtx.Bind(&request); err != nil {
			log.Error().Err(err).Msg("could not bind request")
			return echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
		}

		insertParams, err := mapper.FromCreateRequestToInsertCompetitionAct(&request)
		if err != nil {
			log.Error().Err(err).Msg("could not map request to insert params")
			return err
		}

		err = queries.InsertCompetitionAct(ctx, insertParams)
		if err != nil {
			log.Error().Err(err).Msg("could not insert participation")
			return err
		}

		return echoCtx.NoContent(http.StatusCreated)
	}
}

func deleteParticipation(connPool *pgxpool.Pool) echo.HandlerFunc {
	return func(echoCtx echo.Context) error {
		ctx := echoCtx.Request().Context()
		log := zerolog.Ctx(ctx)
		authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)
		if err := authUser.CheckHasPermission(permission.WriteParticipations); err != nil {
			log.Error().Err(err).Msg("user does not have permission to write participations")
			return err
		}

		conn, err := connPool.Acquire(ctx)
		if err != nil {
			log.Error().Err(err).Msg("could not acquire connection")
			return err
		}
		defer conn.Release()

		queries := db.New(conn)

		var request mapper.DeleteParticipationRequest
		if err := echoCtx.Bind(&request); err != nil {
			log.Error().Err(err).Msg("could not bind request")
			return echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
		}

		deleteParams, err := mapper.FromDeleteRequestToDeleteCompetitionAct(&request)
		if err != nil {
			log.Error().Err(err).Msg("could not map request to delete params")
			return err
		}

		err = queries.DeleteCompetitionAct(ctx, deleteParams)
		if err != nil {
			log.Error().Err(err).Msg("could not delete participation")
			return err
		}

		return echoCtx.NoContent(http.StatusNoContent)
	}
}

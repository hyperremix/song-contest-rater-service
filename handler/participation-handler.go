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

func registerParticipationRoutes(e *echo.Echo, connPool *pgxpool.Pool) {
	e.POST("/participations", createParticipation(connPool))
	e.DELETE("/participations", deleteParticipation(connPool))
}

func createParticipation(connPool *pgxpool.Pool) echo.HandlerFunc {
	return func(echoCtx echo.Context) error {
		ctx := echoCtx.Request().Context()
		authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)
		if err := authUser.CheckHasPermission(permission.WriteParticipations); err != nil {
			return err
		}

		conn, err := connPool.Acquire(ctx)
		if err != nil {
			return err
		}
		defer conn.Release()

		queries := db.New(conn)

		var request pb.CreateParticipationRequest
		if err := echoCtx.Bind(&request); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
		}

		insertParams, err := mapper.FromCreateRequestToInsertCompetitionAct(&request)
		if err != nil {
			return err
		}

		err = queries.InsertCompetitionAct(ctx, insertParams)
		if err != nil {
			return err
		}

		return echoCtx.NoContent(http.StatusCreated)
	}
}

func deleteParticipation(connPool *pgxpool.Pool) echo.HandlerFunc {
	return func(echoCtx echo.Context) error {
		ctx := echoCtx.Request().Context()
		authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)
		if err := authUser.CheckHasPermission(permission.WriteParticipations); err != nil {
			return err
		}

		conn, err := connPool.Acquire(ctx)
		if err != nil {
			return err
		}
		defer conn.Release()

		queries := db.New(conn)

		var request mapper.DeleteParticipationRequest
		if err := echoCtx.Bind(&request); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
		}

		deleteParams, err := mapper.FromDeleteRequestToDeleteCompetitionAct(&request)
		if err != nil {
			return err
		}

		err = queries.DeleteCompetitionAct(ctx, deleteParams)
		if err != nil {
			return err
		}

		return echoCtx.NoContent(http.StatusNoContent)
	}
}

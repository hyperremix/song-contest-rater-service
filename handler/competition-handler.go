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

func registerCompetitionRoutes(e *echo.Echo, connPool *pgxpool.Pool) {
	e.GET("/competitions", listCompetitions(connPool))
	e.GET("/competitions/:id", getCompetition(connPool))
	e.POST("/competitions", createCompetition(connPool))
	e.PUT("/competitions/:id", updateCompetition(connPool))
	e.DELETE("/competitions/:id", deleteCompetition(connPool))
}

func listCompetitions(connPool *pgxpool.Pool) echo.HandlerFunc {
	return func(echoCtx echo.Context) error {
		ctx := echoCtx.Request().Context()
		conn, err := connPool.Acquire(ctx)
		if err != nil {
			return err
		}
		defer conn.Release()

		queries := db.New(conn)

		competitions, err := queries.ListCompetitions(ctx)
		if err != nil {
			return err
		}

		response, err := mapper.FromDbCompetitionListToResponse(competitions)
		if err != nil {
			return err
		}

		return echoCtx.JSON(http.StatusOK, response)
	}
}
func getCompetition(connPool *pgxpool.Pool) echo.HandlerFunc {
	return func(echoCtx echo.Context) error {
		ctx := echoCtx.Request().Context()
		conn, err := connPool.Acquire(ctx)
		if err != nil {
			return err
		}
		defer conn.Release()

		queries := db.New(conn)

		var request singleObjectRequest
		if err := echoCtx.Bind(&request); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
		}

		id, err := mapper.FromProtoToDbId(request.Id)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "could not map id")
		}

		competition, err := queries.GetCompetitionById(ctx, id)
		if err != nil {
			return err
		}

		ratings, err := queries.ListRatingsByCompetitionId(ctx, competition.ID)
		if err != nil {
			return err
		}

		acts, err := queries.ListActsByCompetitionId(ctx, competition.ID)
		if err != nil {
			return err
		}

		users, err := queries.ListUsersByCompetitionId(ctx, competition.ID)
		if err != nil {
			return err
		}

		response, err := mapper.FromDbToCompetitionWithRatingsActsAndUsersResponse(competition, ratings, acts, users)
		if err != nil {
			return err
		}

		return echoCtx.JSON(http.StatusOK, response)
	}
}
func createCompetition(connPool *pgxpool.Pool) echo.HandlerFunc {
	return func(echoCtx echo.Context) error {
		ctx := echoCtx.Request().Context()
		authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)
		if err := authUser.CheckHasPermission(permission.WriteActs); err != nil {
			return err
		}

		conn, err := connPool.Acquire(ctx)
		if err != nil {
			return err
		}
		defer conn.Release()

		queries := db.New(conn)

		var request pb.CreateCompetitionRequest
		if err := echoCtx.Bind(&request); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
		}

		insertParams := mapper.FromCreateRequestToInsertCompetition(&request)

		competition, err := queries.InsertCompetition(ctx, insertParams)
		if err != nil {
			return err
		}

		response, err := mapper.FromDbCompetitionToResponse(competition)
		if err != nil {
			return err
		}

		return echoCtx.JSON(http.StatusOK, response)
	}
}

func updateCompetition(connPool *pgxpool.Pool) echo.HandlerFunc {
	return func(echoCtx echo.Context) error {
		ctx := echoCtx.Request().Context()
		authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)
		if err := authUser.CheckHasPermission(permission.WriteActs); err != nil {
			return err
		}

		conn, err := connPool.Acquire(ctx)
		if err != nil {
			return err
		}
		defer conn.Release()

		queries := db.New(conn)

		var request pb.UpdateCompetitionRequest
		if err := echoCtx.Bind(&request); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
		}

		var params singleObjectRequest
		if err := echoCtx.Bind(&params); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
		}

		if request.Id != params.Id {
			return echo.NewHTTPError(http.StatusBadRequest, "id in request does not match id in path")
		}

		updateParams, err := mapper.FromUpdateRequestToUpdateCompetition(&request)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "could not map request to update params")
		}

		competition, err := queries.UpdateCompetition(ctx, updateParams)
		if err != nil {
			return err
		}

		response, err := mapper.FromDbCompetitionToResponse(competition)
		if err != nil {
			return err
		}

		return echoCtx.JSON(http.StatusOK, response)
	}
}
func deleteCompetition(connPool *pgxpool.Pool) echo.HandlerFunc {
	return func(echoCtx echo.Context) error {
		ctx := echoCtx.Request().Context()
		authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)
		if err := authUser.CheckHasPermission(permission.WriteActs); err != nil {
			return err
		}

		conn, err := connPool.Acquire(ctx)
		if err != nil {
			return err
		}
		defer conn.Release()

		queries := db.New(conn)

		var request singleObjectRequest
		if err := echoCtx.Bind(&request); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
		}

		id, err := mapper.FromProtoToDbId(request.Id)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "could not map id")
		}

		competition, err := queries.DeleteCompetitionById(ctx, id)
		if err != nil {
			return err
		}

		response, err := mapper.FromDbCompetitionToResponse(competition)
		if err != nil {
			return err
		}

		return echoCtx.JSON(http.StatusOK, response)
	}
}

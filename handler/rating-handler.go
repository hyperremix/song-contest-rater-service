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

func registerRatingRoutes(e *echo.Echo, connPool *pgxpool.Pool) {
	e.GET("/ratings", listRatings(connPool))
	e.GET("/acts/:id/ratings", listActRatings(connPool))
	e.GET("/users/:id/ratings", listUserRatings(connPool))
	e.GET("/ratings/:id", getRating(connPool))
	e.POST("/ratings", createRating(connPool))
	e.PUT("/ratings/:id", updateRating(connPool))
	e.DELETE("/ratings/:id", deleteRating(connPool))
}

func listRatings(connPool *pgxpool.Pool) echo.HandlerFunc {
	return func(echoCtx echo.Context) error {
		ctx := echoCtx.Request().Context()
		conn, err := connPool.Acquire(ctx)
		if err != nil {
			return err
		}
		defer conn.Release()

		queries := db.New(conn)

		ratings, err := queries.ListRatings(ctx)
		if err != nil {
			return err
		}

		response, err := mapper.FromDbRatingListToResponse(ratings)
		if err != nil {
			return err
		}

		return echoCtx.JSON(http.StatusOK, response)
	}
}

func listActRatings(connPool *pgxpool.Pool) echo.HandlerFunc {
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

		actId, err := mapper.FromProtoToDbId(request.Id)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
		}

		ratings, err := queries.ListRatingsByActId(ctx, actId)
		if err != nil {
			return err
		}

		response, err := mapper.FromDbRatingListToResponse(ratings)
		if err != nil {
			return err
		}

		return echoCtx.JSON(http.StatusOK, response)
	}
}

func listUserRatings(connPool *pgxpool.Pool) echo.HandlerFunc {
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

		userId, err := mapper.FromProtoToDbId(request.Id)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
		}

		ratings, err := queries.ListRatingsByUserId(ctx, userId)
		if err != nil {
			return err
		}

		response, err := mapper.FromDbRatingListToResponse(ratings)
		if err != nil {
			return err
		}

		return echoCtx.JSON(http.StatusOK, response)
	}
}

func getRating(connPool *pgxpool.Pool) echo.HandlerFunc {
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
			return echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
		}

		rating, err := queries.GetRatingById(ctx, id)
		if err != nil {
			return err
		}

		response, err := mapper.FromDbRatingToResponse(rating)
		if err != nil {
			return err
		}

		return echoCtx.JSON(http.StatusOK, response)
	}
}

func createRating(connPool *pgxpool.Pool) echo.HandlerFunc {
	return func(echoCtx echo.Context) error {
		ctx := echoCtx.Request().Context()
		authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)

		conn, err := connPool.Acquire(ctx)
		if err != nil {
			return err
		}
		defer conn.Release()

		queries := db.New(conn)

		var request pb.CreateRatingRequest
		if err := echoCtx.Bind(&request); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
		}

		insertRatingParams, err := mapper.FromCreateRequestToInsertRating(&request, authUser.UserID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "could not map request")
		}

		rating, err := queries.InsertRating(ctx, insertRatingParams)
		if err != nil {
			return err
		}

		response, err := mapper.FromDbRatingToResponse(rating)
		if err != nil {
			return err
		}

		return echoCtx.JSON(http.StatusOK, response)
	}
}

func updateRating(connPool *pgxpool.Pool) echo.HandlerFunc {
	return func(echoCtx echo.Context) error {
		ctx := echoCtx.Request().Context()
		authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)
		conn, err := connPool.Acquire(ctx)
		if err != nil {
			return err
		}
		defer conn.Release()

		queries := db.New(conn)

		var request pb.UpdateRatingRequest
		if err := echoCtx.Bind(&request); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
		}

		var params singleObjectRequest
		if err := echoCtx.Bind(&params); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
		}

		if params.Id != request.Id {
			return echo.NewHTTPError(http.StatusBadRequest, "id mismatch")
		}

		id, err := mapper.FromProtoToDbId(request.Id)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
		}

		existingRating, err := queries.GetRatingById(ctx, id)
		if err != nil {
			return err
		}

		if err := authUser.CheckIsOwner(existingRating); err != nil {
			return err
		}

		updateRatingParams, err := mapper.FromUpdateRequestToUpdateRating(&request)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "could not map request")
		}

		rating, err := queries.UpdateRating(ctx, updateRatingParams)
		if err != nil {
			return err
		}

		response, err := mapper.FromDbRatingToResponse(rating)
		if err != nil {
			return err
		}

		return echoCtx.JSON(http.StatusOK, response)
	}
}

func deleteRating(connPool *pgxpool.Pool) echo.HandlerFunc {
	return func(echoCtx echo.Context) error {
		ctx := echoCtx.Request().Context()
		authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)
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
			return echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
		}

		existingRating, err := queries.GetRatingById(ctx, id)
		if err != nil {
			return err
		}

		if err := authUser.CheckIsOwner(existingRating); err != nil {
			return err
		}

		rating, err := queries.DeleteRatingById(ctx, id)
		if err != nil {
			return err
		}

		response, err := mapper.FromDbRatingToResponse(rating)
		if err != nil {
			return err
		}

		return echoCtx.JSON(http.StatusOK, response)
	}
}

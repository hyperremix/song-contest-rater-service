package handler

import (
	"errors"
	"net/http"

	"github.com/hyperremix/song-contest-rater-service/authz"
	"github.com/hyperremix/song-contest-rater-service/db"
	"github.com/hyperremix/song-contest-rater-service/mapper"
	"github.com/hyperremix/song-contest-rater-service/permission"
	pb "github.com/hyperremix/song-contest-rater-service/protos/songcontestrater"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

func registerUserRoutes(e *echo.Group, connPool *pgxpool.Pool) {
	e.GET("/users", listUsers(connPool))
	e.GET("/users/:id", getUser(connPool))
	e.GET("/users/me", getAuthUser(connPool))
	e.POST("/users", createUser(connPool))
	e.PUT("/users/:id", updateUser(connPool))
	e.DELETE("/users/:id", deleteUser(connPool))
}

func listUsers(connPool *pgxpool.Pool) echo.HandlerFunc {
	return func(echoCtx echo.Context) error {
		ctx := echoCtx.Request().Context()
		log := zerolog.Ctx(ctx)
		conn, err := connPool.Acquire(ctx)
		if err != nil {
			log.Error().Err(err).Msg("could not acquire connection")
			return err
		}
		defer conn.Release()

		queries := db.New(conn)

		users, err := queries.ListUsers(ctx)
		if err != nil {
			log.Error().Err(err).Msg("could not list users")
			return err
		}

		response, err := mapper.FromDbUserListToResponse(users)
		if err != nil {
			log.Error().Err(err).Msg("could not map to response")
			return err
		}

		return echoCtx.JSON(http.StatusOK, response)
	}
}

func getUser(connPool *pgxpool.Pool) echo.HandlerFunc {
	return func(echoCtx echo.Context) error {
		ctx := echoCtx.Request().Context()
		log := zerolog.Ctx(ctx)
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

		user, err := queries.GetUserById(ctx, id)
		if err != nil {
			log.Error().Err(err).Msg("could not get user")
			return err
		}

		response, err := mapper.FromDbUserToResponse(user)
		if err != nil {
			log.Error().Err(err).Msg("could not map to response")
			return err
		}

		return echoCtx.JSON(http.StatusOK, response)
	}
}

func getAuthUser(connPool *pgxpool.Pool) echo.HandlerFunc {
	return func(echoCtx echo.Context) error {
		authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)
		ctx := echoCtx.Request().Context()
		log := zerolog.Ctx(ctx)
		conn, err := connPool.Acquire(ctx)
		if err != nil {
			log.Error().Err(err).Msg("could not acquire connection")
			return err
		}
		defer conn.Release()

		queries := db.New(conn)

		user, err := queries.GetUserBySub(ctx, authUser.Sub)
		if errors.Is(err, pgx.ErrNoRows) {
			log.Error().Err(err).Msg("user not found")
			return echo.NewHTTPError(http.StatusNotFound, "user not found")
		}

		if err != nil {
			log.Error().Err(err).Msg("could not get user")
			return err
		}

		response, err := mapper.FromDbUserToResponse(user)
		if err != nil {
			log.Error().Err(err).Msg("could not map to response")
			return err
		}

		return echoCtx.JSON(http.StatusOK, response)
	}
}

func createUser(connPool *pgxpool.Pool) echo.HandlerFunc {
	return func(echoCtx echo.Context) error {
		ctx := echoCtx.Request().Context()
		log := zerolog.Ctx(ctx)
		authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)
		conn, err := connPool.Acquire(ctx)
		if err != nil {
			log.Error().Err(err).Msg("could not acquire connection")
			return err
		}
		defer conn.Release()

		queries := db.New(conn)

		var request pb.CreateUserRequest
		if err := echoCtx.Bind(&request); err != nil {
			log.Error().Err(err).Msg("could not bind request")
			return echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
		}

		insertParams, err := mapper.FromCreateRequestToInsertUser(&request, authUser.Sub)
		if err != nil {
			log.Error().Err(err).Msg("could not map request")
			return echo.NewHTTPError(http.StatusBadRequest, "could not map request")
		}

		user, err := queries.InsertUser(ctx, insertParams)
		if err != nil {
			log.Error().Err(err).Msg("could not insert user")
			return err
		}

		response, err := mapper.FromDbUserToResponse(user)
		if err != nil {
			log.Error().Err(err).Msg("could not map to response")
			return err
		}

		return echoCtx.JSON(http.StatusCreated, response)
	}
}

func updateUser(connPool *pgxpool.Pool) echo.HandlerFunc {
	return func(echoCtx echo.Context) error {
		ctx := echoCtx.Request().Context()
		log := zerolog.Ctx(ctx)
		conn, err := connPool.Acquire(ctx)
		if err != nil {
			log.Error().Err(err).Msg("could not acquire connection")
			return err
		}
		defer conn.Release()

		queries := db.New(conn)

		var request pb.UpdateUserRequest
		if err := echoCtx.Bind(&request); err != nil {
			log.Error().Err(err).Msg("could not bind request")
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		paramId := echoCtx.Param("id")

		if request.Id != paramId {
			log.Error().Msg("id mismatch")
			return echo.NewHTTPError(http.StatusBadRequest, "id mismatch")
		}

		authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)
		if err := authUser.CheckHasPermission(permission.WriteUsers); err != nil && authUser.UserID != request.Id {
			log.Error().Err(err).Msg("user does not have permission to write users or is not the user to be updated")
			return err
		}

		updateParams, err := mapper.FromUpdateRequestToUpdateUser(&request)
		if err != nil {
			log.Error().Err(err).Msg("could not map request to update params")
			return echo.NewHTTPError(http.StatusBadRequest, "could not map request to update params")
		}

		user, err := queries.UpdateUser(ctx, updateParams)
		if err != nil {
			log.Error().Err(err).Msg("could not update user")
			return err
		}

		response, err := mapper.FromDbUserToResponse(user)
		if err != nil {
			log.Error().Err(err).Msg("could not map to response")
			return err
		}

		return echoCtx.JSON(http.StatusOK, response)
	}
}

func deleteUser(connPool *pgxpool.Pool) echo.HandlerFunc {
	return func(echoCtx echo.Context) error {
		ctx := echoCtx.Request().Context()
		log := zerolog.Ctx(ctx)
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

		authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)
		if err := authUser.CheckHasPermission(permission.WriteUsers); err != nil && authUser.UserID != request.Id {
			log.Error().Err(err).Msg("user does not have permission to write users or is not the user to be deleted")
			return err
		}

		user, err := queries.DeleteUserById(ctx, id)
		if err != nil {
			log.Error().Err(err).Msg("could not delete user")
			return err
		}

		response, err := mapper.FromDbUserToResponse(user)
		if err != nil {
			log.Error().Err(err).Msg("could not map to response")
			return err
		}

		return echoCtx.JSON(http.StatusOK, response)
	}
}

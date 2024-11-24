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

func registerUserRoutes(e *echo.Echo, connPool *pgxpool.Pool) {
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
		conn, err := connPool.Acquire(ctx)
		if err != nil {
			return err
		}
		defer conn.Release()

		queries := db.New(conn)

		users, err := queries.ListUsers(ctx)
		if err != nil {
			return err
		}

		response, err := mapper.FromDbUserListToResponse(users)
		if err != nil {
			return err
		}

		return echoCtx.JSON(http.StatusOK, response)
	}
}

func getUser(connPool *pgxpool.Pool) echo.HandlerFunc {
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

		user, err := queries.GetUserById(ctx, id)
		if err != nil {
			return err
		}

		response, err := mapper.FromDbUserToResponse(user)
		if err != nil {
			return err
		}

		return echoCtx.JSON(http.StatusOK, response)
	}
}

func getAuthUser(connPool *pgxpool.Pool) echo.HandlerFunc {
	return func(echoCtx echo.Context) error {
		authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)
		ctx := echoCtx.Request().Context()
		conn, err := connPool.Acquire(ctx)
		if err != nil {
			return err
		}
		defer conn.Release()

		queries := db.New(conn)

		user, err := queries.GetUserBySub(ctx, authUser.Sub)
		if err != nil {
			return err
		}

		response, err := mapper.FromDbUserToResponse(user)
		if err != nil {
			return err
		}

		return echoCtx.JSON(http.StatusOK, response)
	}
}

func createUser(connPool *pgxpool.Pool) echo.HandlerFunc {
	return func(echoCtx echo.Context) error {
		ctx := echoCtx.Request().Context()
		authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)
		conn, err := connPool.Acquire(ctx)
		if err != nil {
			return err
		}
		defer conn.Release()

		queries := db.New(conn)

		var request pb.CreateUserRequest
		if err := echoCtx.Bind(&request); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
		}

		insertParams, err := mapper.FromCreateRequestToInsertUser(&request, authUser.Sub)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "could not map request")
		}

		user, err := queries.InsertUser(ctx, insertParams)
		if err != nil {
			return err
		}

		response, err := mapper.FromDbUserToResponse(user)
		if err != nil {
			return err
		}

		return echoCtx.JSON(http.StatusCreated, response)
	}
}

func updateUser(connPool *pgxpool.Pool) echo.HandlerFunc {
	return func(echoCtx echo.Context) error {
		ctx := echoCtx.Request().Context()
		conn, err := connPool.Acquire(ctx)
		if err != nil {
			return err
		}
		defer conn.Release()

		queries := db.New(conn)

		var request pb.UpdateUserRequest
		if err := echoCtx.Bind(&request); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
		}

		var params singleObjectRequest
		if err := echoCtx.Bind(&params); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
		}

		if request.Id != params.Id {
			return echo.NewHTTPError(http.StatusBadRequest, "id mismatch")
		}

		authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)
		if err := authUser.CheckHasPermission(permission.WriteUsers); err != nil && authUser.UserID != request.Id {
			return err
		}

		updateParams, err := mapper.FromUpdateRequestToUpdateUser(&request)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "could not map request to update params")
		}

		user, err := queries.UpdateUser(ctx, updateParams)
		if err != nil {
			return err
		}

		response, err := mapper.FromDbUserToResponse(user)
		if err != nil {
			return err
		}

		return echoCtx.JSON(http.StatusOK, response)
	}
}

func deleteUser(connPool *pgxpool.Pool) echo.HandlerFunc {
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

		authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)
		if err := authUser.CheckHasPermission(permission.WriteUsers); err != nil && authUser.UserID != request.Id {
			return err
		}

		user, err := queries.DeleteUserById(ctx, id)
		if err != nil {
			return err
		}

		response, err := mapper.FromDbUserToResponse(user)
		if err != nil {
			return err
		}

		return echoCtx.JSON(http.StatusOK, response)
	}
}

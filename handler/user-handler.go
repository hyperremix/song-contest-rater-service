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

type UserHandler struct {
	queries  *db.Queries
	connPool *pgxpool.Pool
}

func NewUserHandler(connPool *pgxpool.Pool) *UserHandler {
	return &UserHandler{
		queries:  db.New(connPool),
		connPool: connPool,
	}
}

func registerUserRoutes(e *echo.Group, connPool *pgxpool.Pool) {
	h := NewUserHandler(connPool)

	e.GET("/users", h.listUsers)
	e.GET("/users/:id", h.getUser)
	e.GET("/users/me", h.getAuthUser)
	e.POST("/users", h.createUser)
	e.PUT("/users/:id", h.updateUser)
	e.DELETE("/users/:id", h.deleteUser)
}

func (h *UserHandler) listUsers(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()
	log := zerolog.Ctx(ctx)

	users, err := h.queries.ListUsers(ctx)
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

func (h *UserHandler) getUser(echoCtx echo.Context) error {
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

	user, err := h.queries.GetUserById(ctx, id)
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

func (h *UserHandler) getAuthUser(echoCtx echo.Context) error {
	authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)
	ctx := echoCtx.Request().Context()
	log := zerolog.Ctx(ctx)

	user, err := h.queries.GetUserBySub(ctx, authUser.Sub)
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

func (h *UserHandler) createUser(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()
	log := zerolog.Ctx(ctx)
	authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)

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

	user, err := h.queries.InsertUser(ctx, insertParams)
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

func (h *UserHandler) updateUser(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()
	log := zerolog.Ctx(ctx)

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

	user, err := h.queries.UpdateUser(ctx, updateParams)
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

func (h *UserHandler) deleteUser(echoCtx echo.Context) error {
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

	authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)
	if err := authUser.CheckHasPermission(permission.WriteUsers); err != nil && authUser.UserID != request.Id {
		log.Error().Err(err).Msg("user does not have permission to write users or is not the user to be deleted")
		return err
	}

	user, err := h.queries.DeleteUserById(ctx, id)
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

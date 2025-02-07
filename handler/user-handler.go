package handler

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/hyperremix/song-contest-rater-service/authz"
	"github.com/hyperremix/song-contest-rater-service/db"
	"github.com/hyperremix/song-contest-rater-service/mapper"
	"github.com/hyperremix/song-contest-rater-service/permission"
	pb "github.com/hyperremix/song-contest-rater-service/protos/songcontestrater"
	"github.com/hyperremix/song-contest-rater-service/s3"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	queries  *db.Queries
	connPool *pgxpool.Pool
	s3Client *s3.Client
}

func NewUserHandler(connPool *pgxpool.Pool) *UserHandler {
	return &UserHandler{
		queries:  db.New(connPool),
		connPool: connPool,
		s3Client: s3.New(),
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
	e.POST("/users/me/profile-picture-presigned-url", h.getProfilePicturePresignedURL)
}

func (h *UserHandler) listUsers(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()

	users, err := h.queries.ListUsers(ctx)
	if err != nil {
		return err
	}

	response, err := mapper.FromDbUserListToResponse(users)
	if err != nil {
		return err
	}

	return echoCtx.JSON(http.StatusOK, response)
}

func (h *UserHandler) getUser(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()

	var request singleObjectRequest
	if err := echoCtx.Bind(&request); err != nil {
		return err
	}

	id, err := mapper.FromProtoToDbId(request.Id)
	if err != nil {
		return err
	}

	user, err := h.queries.GetUserById(ctx, id)
	if err != nil {
		return err
	}

	response, err := mapper.FromDbUserToResponse(user)
	if err != nil {
		return err
	}

	return echoCtx.JSON(http.StatusOK, response)
}

func (h *UserHandler) getAuthUser(echoCtx echo.Context) error {
	authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)
	ctx := echoCtx.Request().Context()

	user, err := h.queries.GetUserBySub(ctx, authUser.Sub)
	if errors.Is(err, pgx.ErrNoRows) {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	if err != nil {
		return err
	}

	response, err := mapper.FromDbUserToResponse(user)
	if err != nil {
		return err
	}

	return echoCtx.JSON(http.StatusOK, response)
}

func (h *UserHandler) createUser(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()
	authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)

	var request pb.CreateUserRequest
	if err := echoCtx.Bind(&request); err != nil {
		return err
	}

	insertParams, err := mapper.FromCreateRequestToInsertUser(&request, authUser.Sub)
	if err != nil {
		return err
	}

	user, err := h.queries.InsertUser(ctx, insertParams)
	if err != nil {
		return err
	}

	response, err := mapper.FromDbUserToResponse(user)
	if err != nil {
		return err
	}

	return echoCtx.JSON(http.StatusCreated, response)
}

func (h *UserHandler) updateUser(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()

	var request pb.UpdateUserRequest
	if err := echoCtx.Bind(&request); err != nil {
		return err
	}

	paramId := echoCtx.Param("id")

	if request.Id != paramId {
		return echo.NewHTTPError(http.StatusBadRequest, "id mismatch")
	}

	authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)
	if err := authUser.CheckHasPermission(permission.WriteUsers); err != nil && authUser.UserID != request.Id {
		return err
	}

	updateParams, err := mapper.FromUpdateRequestToUpdateUser(&request)
	if err != nil {
		return err
	}

	user, err := h.queries.UpdateUser(ctx, updateParams)
	if err != nil {
		return err
	}

	response, err := mapper.FromDbUserToResponse(user)
	if err != nil {
		return err
	}

	return echoCtx.JSON(http.StatusOK, response)
}

func (h *UserHandler) deleteUser(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()

	var request singleObjectRequest
	if err := echoCtx.Bind(&request); err != nil {
		return err
	}

	id, err := mapper.FromProtoToDbId(request.Id)
	if err != nil {
		return err
	}

	authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)
	if err := authUser.CheckHasPermission(permission.WriteUsers); err != nil && authUser.UserID != request.Id {
		return err
	}

	user, err := h.queries.DeleteUserById(ctx, id)
	if err != nil {
		return err
	}

	response, err := mapper.FromDbUserToResponse(user)
	if err != nil {
		return err
	}

	return echoCtx.JSON(http.StatusOK, response)
}

func (h *UserHandler) getProfilePicturePresignedURL(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()
	authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)

	var request pb.GetPresignedURLRequest
	if err := echoCtx.Bind(&request); err != nil {
		return err
	}

	if !strings.HasPrefix(request.ContentType, "image/") {
		return echo.NewHTTPError(http.StatusBadRequest, "content type must be an image")
	}

	filename := fmt.Sprintf("%s%d%s", authUser.UserID, time.Now().Unix(), filepath.Ext(request.FileName))

	s3Response, err := h.s3Client.GetPresignedURL(ctx, filename, request.ContentType)
	if err != nil {
		return err
	}

	updateParams, err := mapper.FromProfilePictureToUpdateUserImageUrl(authUser.UserID, s3Response.ImageURL)
	if err != nil {
		return err
	}

	_, err = h.queries.UpdateUserImageUrl(ctx, updateParams)
	if err != nil {
		return err
	}

	response := &pb.GetPresignedURLResponse{
		PresignedUrl: s3Response.PresignedURL,
		ImageUrl:     s3Response.ImageURL,
	}

	return echoCtx.JSON(http.StatusOK, response)
}

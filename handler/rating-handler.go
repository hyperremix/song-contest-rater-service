package handler

import (
	"net/http"
	"time"

	"github.com/hyperremix/song-contest-rater-service/authz"
	"github.com/hyperremix/song-contest-rater-service/db"
	"github.com/hyperremix/song-contest-rater-service/mapper"
	pb "github.com/hyperremix/song-contest-rater-service/protos/songcontestrater"
	"github.com/hyperremix/song-contest-rater-service/sse"
	"github.com/hyperremix/song-contest-rater-service/stat"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type RatingHandler struct {
	queries     *db.Queries
	pool        *pgxpool.Pool
	statService *stat.Service
}

func NewRatingHandler(pool *pgxpool.Pool) *RatingHandler {
	return &RatingHandler{
		queries:     db.New(pool),
		pool:        pool,
		statService: stat.NewService(pool),
	}
}

func registerRatingRoutes(e *echo.Group, connPool *pgxpool.Pool) {
	h := NewRatingHandler(connPool)

	e.GET("/ratings", h.listRatings)
	e.GET("/users/:id/ratings", h.listUserRatings)
	e.GET("/ratings/:id", h.getRating)
	e.POST("/ratings", h.createRating)
	e.PUT("/ratings/:id", h.updateRating)
	e.DELETE("/ratings/:id", h.deleteRating)
	e.GET("/ratings/events", h.streamRatings)
}

func (h *RatingHandler) listRatings(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()

	ratings, err := h.queries.ListRatings(ctx)
	if err != nil {
		return err
	}

	response, err := mapper.FromDbRatingListToResponse(ratings, make([]db.User, 0))
	if err != nil {
		return err
	}

	return echoCtx.JSON(http.StatusOK, response)
}

func (h *RatingHandler) listUserRatings(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()

	var request singleObjectRequest
	if err := echoCtx.Bind(&request); err != nil {
		return err
	}

	userId, err := mapper.FromProtoToDbId(request.Id)
	if err != nil {
		return err
	}

	ratings, err := h.queries.ListRatingsByUserId(ctx, userId)
	if err != nil {
		return err
	}

	response, err := mapper.FromDbRatingListToResponse(ratings, make([]db.User, 0))
	if err != nil {
		return err
	}

	return echoCtx.JSON(http.StatusOK, response)
}

func (h *RatingHandler) getRating(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()

	var request singleObjectRequest
	if err := echoCtx.Bind(&request); err != nil {
		return err
	}

	id, err := mapper.FromProtoToDbId(request.Id)
	if err != nil {
		return err
	}

	rating, err := h.queries.GetRatingById(ctx, id)
	if err != nil {
		return err
	}

	user, err := h.queries.GetUserById(ctx, rating.UserID)
	if err != nil {
		return err
	}

	response, err := mapper.FromDbRatingToResponse(rating, &user)
	if err != nil {
		return err
	}

	return echoCtx.JSON(http.StatusOK, response)
}

func (h *RatingHandler) createRating(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()
	authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)

	var request pb.CreateRatingRequest
	if err := echoCtx.Bind(&request); err != nil {
		return err
	}

	insertRatingParams, err := mapper.FromCreateRequestToInsertRating(&request, authUser.UserID)
	if err != nil {
		return err
	}

	competition, err := h.queries.GetCompetitionById(ctx, insertRatingParams.CompetitionID)
	if err != nil {
		return err
	}

	if competition.StartTime.Time.After(time.Now()) {
		return echo.NewHTTPError(http.StatusBadRequest, "competition has not started yet")
	}

	rating, err := h.queries.InsertRating(ctx, insertRatingParams)
	if err != nil {
		return err
	}

	response, err := mapper.FromDbRatingToResponse(rating, &authUser.User)
	if err != nil {
		return err
	}

	event, err := sse.NewEvent(sse.EventOptions{
		ID:    response.Id,
		Event: "createRating",
		Data:  response,
		Retry: 10000,
	})
	if err != nil {
		return err
	}

	broker.BroadcastEvent(authUser.UserID, event)
	h.statService.UpsertRatingStats(ctx, response)
	return echoCtx.JSON(http.StatusCreated, response)
}

func (h *RatingHandler) updateRating(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()
	authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)

	var request pb.UpdateRatingRequest
	if err := echoCtx.Bind(&request); err != nil {
		return err
	}

	paramId := echoCtx.Param("id")
	if paramId != request.Id {
		return echo.NewHTTPError(http.StatusBadRequest, "id mismatch")
	}

	id, err := mapper.FromProtoToDbId(request.Id)
	if err != nil {
		return err
	}

	existingRating, err := h.queries.GetRatingById(ctx, id)
	if err != nil {
		return err
	}

	if err := authUser.CheckIsOwner(existingRating); err != nil {
		return err
	}

	updateRatingParams, err := mapper.FromUpdateRequestToUpdateRating(&request)
	if err != nil {
		return err
	}

	rating, err := h.queries.UpdateRating(ctx, updateRatingParams)
	if err != nil {
		return err
	}

	response, err := mapper.FromDbRatingToResponse(rating, &authUser.User)
	if err != nil {
		return err
	}

	event, err := sse.NewEvent(sse.EventOptions{
		ID:    response.Id,
		Event: "updateRating",
		Data:  response,
		Retry: 10000,
	})
	if err != nil {
		return err
	}

	broker.BroadcastEvent(authUser.UserID, event)
	return echoCtx.JSON(http.StatusOK, response)
}

func (h *RatingHandler) deleteRating(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()
	authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)

	var request singleObjectRequest
	if err := echoCtx.Bind(&request); err != nil {
		return err
	}

	id, err := mapper.FromProtoToDbId(request.Id)
	if err != nil {
		return err
	}

	existingRating, err := h.queries.GetRatingById(ctx, id)
	if err != nil {
		return err
	}

	if err := authUser.CheckIsOwner(existingRating); err != nil {
		return err
	}

	rating, err := h.queries.DeleteRatingById(ctx, id)
	if err != nil {
		return err
	}

	response, err := mapper.FromDbRatingToResponse(rating, nil)
	if err != nil {
		return err
	}

	event, err := sse.NewEvent(sse.EventOptions{
		ID:    response.Id,
		Event: "deleteRating",
		Data:  response,
		Retry: 10000,
	})
	if err != nil {
		return err
	}

	broker.BroadcastEvent(authUser.UserID, event)
	return echoCtx.JSON(http.StatusOK, response)
}

func (h *RatingHandler) streamRatings(echoCtx echo.Context) error {
	echoCtx.Response().Header().Set(echo.HeaderContentType, "text/event-stream")
	echoCtx.Response().Header().Set(echo.HeaderCacheControl, "no-cache")
	echoCtx.Response().Header().Set(echo.HeaderConnection, "keep-alive")
	authUser := echoCtx.Get(authz.AuthUserContextKey).(*authz.AuthUser)

	ticker := time.NewTicker(15 * time.Second)
	ch := make(chan sse.Event)
	broker.AddUserChan(authUser.UserID, ch)

	defer ticker.Stop()
	defer broker.RemoveUserChan(authUser.UserID, ch)
	for {
		select {
		case <-echoCtx.Request().Context().Done():
			return nil
		case e := <-ch:
			if err := e.MarshalTo(echoCtx.Response().Writer); err != nil {
				return err
			}
			echoCtx.Response().Flush()
		case <-ticker.C:
			event := sse.Event{
				Event:   []byte("ping"),
				Retry:   []byte("10000"),
				Comment: []byte("keep-alive"),
			}
			broker.BroadcastEvent(authUser.UserID, event)
		}
	}
}

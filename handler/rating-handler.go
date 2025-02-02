package handler

import (
	"net/http"
	"time"

	"github.com/hyperremix/song-contest-rater-service/authz"
	"github.com/hyperremix/song-contest-rater-service/db"
	"github.com/hyperremix/song-contest-rater-service/mapper"
	pb "github.com/hyperremix/song-contest-rater-service/protos/songcontestrater"
	"github.com/hyperremix/song-contest-rater-service/sse"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type RatingHandler struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

func NewRatingHandler(pool *pgxpool.Pool) *RatingHandler {
	return &RatingHandler{
		queries: db.New(pool),
		pool:    pool,
	}
}

func registerRatingRoutes(e *echo.Group, connPool *pgxpool.Pool) {
	h := NewRatingHandler(connPool)

	e.GET("/ratings", h.listRatings)
	e.GET("/users/:id/ratings", listUserRatings(connPool))
	e.GET("/ratings/:id", getRating(connPool))
	e.POST("/ratings", createRating(connPool))
	e.PUT("/ratings/:id", updateRating(connPool))
	e.DELETE("/ratings/:id", deleteRating(connPool))
	e.GET("/ratings/events", streamRatings())
}

func (h *RatingHandler) listRatings(echoCtx echo.Context) error {
	ctx := echoCtx.Request().Context()
	log := zerolog.Ctx(ctx)

	ratings, err := h.queries.ListRatings(ctx)
	if err != nil {
		log.Error().Err(err).Msg("could not list ratings")
		return err
	}

	response, err := mapper.FromDbRatingListToResponse(ratings, make([]db.User, 0))
	if err != nil {
		log.Error().Err(err).Msg("could not map to response")
		return err
	}

	return echoCtx.JSON(http.StatusOK, response)
}

func listUserRatings(connPool *pgxpool.Pool) echo.HandlerFunc {
	return func(echoCtx echo.Context) error {
		ctx := echoCtx.Request().Context()
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

		userId, err := mapper.FromProtoToDbId(request.Id)
		if err != nil {
			log.Error().Err(err).Msg("could not bind request")
			return echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
		}

		ratings, err := queries.ListRatingsByUserId(ctx, userId)
		if err != nil {
			log.Error().Err(err).Msg("could not list ratings")
			return err
		}

		response, err := mapper.FromDbRatingListToResponse(ratings, make([]db.User, 0))
		if err != nil {
			log.Error().Err(err).Msg("could not map to response")
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
			log.Error().Err(err).Msg("could not bind request")
			return echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
		}

		rating, err := queries.GetRatingById(ctx, id)
		if err != nil {
			log.Error().Err(err).Msg("could not get rating")
			return err
		}

		user, err := queries.GetUserById(ctx, rating.UserID)
		if err != nil {
			log.Error().Err(err).Msg("could not get user")
			return err
		}

		response, err := mapper.FromDbRatingToResponse(rating, &user)
		if err != nil {
			log.Error().Err(err).Msg("could not map to response")
			return err
		}

		return echoCtx.JSON(http.StatusOK, response)
	}
}

func createRating(connPool *pgxpool.Pool) echo.HandlerFunc {
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

		var request pb.CreateRatingRequest
		if err := echoCtx.Bind(&request); err != nil {
			log.Error().Err(err).Msg("could not bind request")
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		insertRatingParams, err := mapper.FromCreateRequestToInsertRating(&request, authUser.UserID)
		if err != nil {
			log.Error().Err(err).Msg("could not map request to insert params")
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		competition, err := queries.GetCompetitionById(ctx, insertRatingParams.CompetitionID)
		if err != nil {
			log.Error().Err(err).Msg("could not get competition")
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		if competition.StartTime.Time.After(time.Now()) {
			log.Error().Msg("competition has not started yet")
			return echo.NewHTTPError(http.StatusBadRequest, "competition has not started yet")
		}

		rating, err := queries.InsertRating(ctx, insertRatingParams)
		if err != nil {
			log.Error().Err(err).Msg("could not insert rating")
			return err
		}

		response, err := mapper.FromDbRatingToResponse(rating, &authUser.User)
		if err != nil {
			log.Error().Err(err).Msg("could not map rating")
			return err
		}

		event, err := sse.NewEvent(sse.EventOptions{
			ID:    response.Id,
			Event: "createRating",
			Data:  response,
			Retry: 10000,
		})
		if err != nil {
			log.Error().Err(err).Msg("could not create event")
			return err
		}

		broker.BroadcastEvent(event)
		return echoCtx.JSON(http.StatusCreated, response)
	}
}

func updateRating(connPool *pgxpool.Pool) echo.HandlerFunc {
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

		var request pb.UpdateRatingRequest
		if err := echoCtx.Bind(&request); err != nil {
			log.Error().Err(err).Msg("could not bind request")
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		paramId := echoCtx.Param("id")
		if paramId != request.Id {
			log.Error().Msg("id mismatch")
			return echo.NewHTTPError(http.StatusBadRequest, "id mismatch")
		}

		id, err := mapper.FromProtoToDbId(request.Id)
		if err != nil {
			log.Error().Err(err).Msg("could not bind id")
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		existingRating, err := queries.GetRatingById(ctx, id)
		if err != nil {
			log.Error().Err(err).Msg("could not get rating")
			return err
		}

		if err := authUser.CheckIsOwner(existingRating); err != nil {
			return err
		}

		updateRatingParams, err := mapper.FromUpdateRequestToUpdateRating(&request)
		if err != nil {
			log.Error().Err(err).Msg("could not bind update params")
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}

		rating, err := queries.UpdateRating(ctx, updateRatingParams)
		if err != nil {
			log.Error().Err(err).Msg("could not update rating")
			return err
		}

		response, err := mapper.FromDbRatingToResponse(rating, &authUser.User)
		if err != nil {
			log.Error().Err(err).Msg("could not map rating")
			return err
		}

		event, err := sse.NewEvent(sse.EventOptions{
			ID:    response.Id,
			Event: "updateRating",
			Data:  response,
			Retry: 10000,
		})
		if err != nil {
			log.Error().Err(err).Msg("could not create event")
			return err
		}

		broker.BroadcastEvent(event)
		return echoCtx.JSON(http.StatusOK, response)
	}
}

func deleteRating(connPool *pgxpool.Pool) echo.HandlerFunc {
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

		var request singleObjectRequest
		if err := echoCtx.Bind(&request); err != nil {
			log.Error().Err(err).Msg("could not bind request")
			return echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
		}

		id, err := mapper.FromProtoToDbId(request.Id)
		if err != nil {
			log.Error().Err(err).Msg("could not bind request")
			return echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
		}

		existingRating, err := queries.GetRatingById(ctx, id)
		if err != nil {
			log.Error().Err(err).Msg("could not get rating")
			return err
		}

		if err := authUser.CheckIsOwner(existingRating); err != nil {
			log.Error().Err(err).Msg("user is not owner")
			return err
		}

		rating, err := queries.DeleteRatingById(ctx, id)
		if err != nil {
			log.Error().Err(err).Msg("could not delete rating")
			return err
		}

		response, err := mapper.FromDbRatingToResponse(rating, nil)
		if err != nil {
			log.Error().Err(err).Msg("could not map to response")
			return err
		}

		event, err := sse.NewEvent(sse.EventOptions{
			ID:    response.Id,
			Event: "deleteRating",
			Data:  response,
			Retry: 10000,
		})
		if err != nil {
			log.Error().Err(err).Msg("could not create event")
			return err
		}

		broker.BroadcastEvent(event)
		return echoCtx.JSON(http.StatusOK, response)
	}
}

func streamRatings() echo.HandlerFunc {
	return func(echoCtx echo.Context) error {
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
				broker.BroadcastEvent(event)
			}
		}
	}
}

package server

import (
	"context"
	"time"

	pb "buf.build/gen/go/hyperremix/song-contest-rater-protos/protocolbuffers/go/songcontestrater/v5"
	"connectrpc.com/connect"
	"github.com/hyperremix/song-contest-rater-service/authz"
	"github.com/hyperremix/song-contest-rater-service/db"
	"github.com/hyperremix/song-contest-rater-service/mapper"
	"github.com/hyperremix/song-contest-rater-service/stat"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RatingServer struct {
	queries     *db.Queries
	pool        *pgxpool.Pool
	statService *stat.Service
}

func NewRatingServer(pool *pgxpool.Pool) *RatingServer {
	return &RatingServer{
		queries:     db.New(pool),
		pool:        pool,
		statService: stat.NewService(pool),
	}
}

func (s *RatingServer) ListRatings(ctx context.Context, request *connect.Request[pb.ListRatingsRequest]) (*connect.Response[pb.ListRatingsResponse], error) {
	ratings, err := s.queries.ListRatings(ctx)
	if err != nil {
		return nil, err
	}

	response, err := mapper.FromDbRatingListToResponse(ratings, make([]db.User, 0))
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&pb.ListRatingsResponse{Ratings: response}), nil
}

func (s *RatingServer) ListUserRatings(ctx context.Context, request *connect.Request[pb.ListUserRatingsRequest]) (*connect.Response[pb.ListUserRatingsResponse], error) {
	userId, err := mapper.FromProtoToDbId(request.Msg.UserId)
	if err != nil {
		return nil, err
	}

	ratings, err := s.queries.ListRatingsByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}

	response, err := mapper.FromDbRatingListToResponse(ratings, make([]db.User, 0))
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&pb.ListUserRatingsResponse{Ratings: response}), nil
}

func (s *RatingServer) GetRating(ctx context.Context, request *connect.Request[pb.GetRatingRequest]) (*connect.Response[pb.GetRatingResponse], error) {
	id, err := mapper.FromProtoToDbId(request.Msg.Id)
	if err != nil {
		return nil, err
	}

	rating, err := s.queries.GetRatingById(ctx, id)
	if err != nil {
		return nil, err
	}

	user, err := s.queries.GetUserById(ctx, rating.UserID)
	if err != nil {
		return nil, err
	}

	response, err := mapper.FromDbRatingToResponse(rating, &user)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&pb.GetRatingResponse{Rating: response}), nil
}

func (s *RatingServer) CreateRating(ctx context.Context, request *connect.Request[pb.CreateRatingRequest]) (*connect.Response[pb.CreateRatingResponse], error) {
	authUser := ctx.Value(authz.AuthUserContextKey).(*authz.AuthUser)

	insertRatingParams, err := mapper.FromCreateRequestToInsertRating(request.Msg, authUser.UserID)
	if err != nil {
		return nil, err
	}

	contest, err := s.queries.GetContestById(ctx, insertRatingParams.ContestID)
	if err != nil {
		return nil, err
	}

	if contest.StartTime.Time.After(time.Now()) {
		return nil, status.Errorf(codes.InvalidArgument, "contest has not started yet")
	}

	rating, err := s.queries.InsertRating(ctx, insertRatingParams)
	if err != nil {
		return nil, err
	}

	response, err := mapper.FromDbRatingToResponse(rating, &authUser.DbUser)
	if err != nil {
		return nil, err
	}

	broker.BroadcastEvent(authUser.UserID, response)
	s.statService.AddRatingToStats(ctx, response)
	return connect.NewResponse(&pb.CreateRatingResponse{Rating: response}), nil
}

func (s *RatingServer) UpdateRating(ctx context.Context, request *connect.Request[pb.UpdateRatingRequest]) (*connect.Response[pb.UpdateRatingResponse], error) {
	authUser := ctx.Value(authz.AuthUserContextKey).(*authz.AuthUser)

	id, err := mapper.FromProtoToDbId(request.Msg.Id)
	if err != nil {
		return nil, err
	}

	existingRating, err := s.queries.GetRatingById(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := authUser.CheckIsOwner(existingRating); err != nil {
		return nil, err
	}

	updateRatingParams, err := mapper.FromUpdateRequestToUpdateRating(request.Msg)
	if err != nil {
		return nil, err
	}

	rating, err := s.queries.UpdateRating(ctx, updateRatingParams)
	if err != nil {
		return nil, err
	}

	response, err := mapper.FromDbRatingToResponse(rating, &authUser.DbUser)
	if err != nil {
		return nil, err
	}

	broker.BroadcastEvent(authUser.UserID, response)
	s.statService.UpdateRatingInStats(ctx, response)
	return connect.NewResponse(&pb.UpdateRatingResponse{Rating: response}), nil
}

func (s *RatingServer) DeleteRating(ctx context.Context, request *connect.Request[pb.DeleteRatingRequest]) (*connect.Response[pb.DeleteRatingResponse], error) {
	authUser := ctx.Value(authz.AuthUserContextKey).(*authz.AuthUser)

	id, err := mapper.FromProtoToDbId(request.Msg.Id)
	if err != nil {
		return nil, err
	}

	existingRating, err := s.queries.GetRatingById(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := authUser.CheckIsOwner(existingRating); err != nil {
		return nil, err
	}

	rating, err := s.queries.DeleteRatingById(ctx, id)
	if err != nil {
		return nil, err
	}

	response, err := mapper.FromDbRatingToResponse(rating, nil)
	if err != nil {
		return nil, err
	}

	broker.BroadcastEvent(authUser.UserID, response)
	s.statService.RemoveRatingFromStats(ctx, response)
	return connect.NewResponse(&pb.DeleteRatingResponse{Rating: response}), nil
}

func (h *RatingServer) StreamRatings(ctx context.Context, request *connect.Request[pb.StreamRatingsRequest], stream *connect.ServerStream[pb.StreamRatingsResponse]) error {
	authUser := ctx.Value(authz.AuthUserContextKey).(*authz.AuthUser)

	ch := make(chan *connect.Response[pb.StreamRatingsResponse])
	broker.AddUserChan(authUser.UserID, ch)

	defer broker.RemoveUserChan(authUser.UserID, ch)
	for {
		select {
		case <-ctx.Done():
			return nil
		case e := <-ch:
			stream.Send(e.Msg)
		}
	}
}

package server

import (
	"context"
	"time"

	pb "github.com/hyperremix/song-contest-rater-protos/v4"
	"github.com/hyperremix/song-contest-rater-service/authz"
	"github.com/hyperremix/song-contest-rater-service/db"
	"github.com/hyperremix/song-contest-rater-service/mapper"
	"github.com/hyperremix/song-contest-rater-service/stat"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type RatingServer struct {
	pb.UnimplementedRatingServer
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

func (s *RatingServer) ListRatings(ctx context.Context, request *emptypb.Empty) (*pb.ListRatingsResponse, error) {
	ratings, err := s.queries.ListRatings(ctx)
	if err != nil {
		return nil, err
	}

	response, err := mapper.FromDbRatingListToResponse(ratings, make([]db.User, 0))
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *RatingServer) ListUserRatings(ctx context.Context, request *pb.ListUserRatingsRequest) (*pb.ListRatingsResponse, error) {
	userId, err := mapper.FromProtoToDbId(request.UserId)
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

	return response, nil
}

func (s *RatingServer) GetRating(ctx context.Context, request *pb.GetRatingRequest) (*pb.RatingResponse, error) {
	id, err := mapper.FromProtoToDbId(request.Id)
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

	return response, nil
}

func (s *RatingServer) CreateRating(ctx context.Context, request *pb.CreateRatingRequest) (*pb.RatingResponse, error) {
	authUser := ctx.Value(authz.AuthUserContextKey{}).(*authz.AuthUser)

	insertRatingParams, err := mapper.FromCreateRequestToInsertRating(request, authUser.UserID)
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
	return response, nil
}

func (s *RatingServer) UpdateRating(ctx context.Context, request *pb.UpdateRatingRequest) (*pb.RatingResponse, error) {
	authUser := ctx.Value(authz.AuthUserContextKey{}).(*authz.AuthUser)

	id, err := mapper.FromProtoToDbId(request.Id)
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

	updateRatingParams, err := mapper.FromUpdateRequestToUpdateRating(request)
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
	return response, nil
}

func (s *RatingServer) DeleteRating(ctx context.Context, request *pb.DeleteRatingRequest) (*pb.RatingResponse, error) {
	authUser := ctx.Value(authz.AuthUserContextKey{}).(*authz.AuthUser)

	id, err := mapper.FromProtoToDbId(request.Id)
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
	return response, nil
}

func (h *RatingServer) StreamRatings(request *emptypb.Empty, stream pb.Rating_StreamRatingsServer) error {
	ctx := stream.Context()
	authUser := ctx.Value(authz.AuthUserContextKey{}).(*authz.AuthUser)

	ch := make(chan *pb.RatingResponse)
	broker.AddUserChan(authUser.UserID, ch)

	defer broker.RemoveUserChan(authUser.UserID, ch)
	for {
		select {
		case <-ctx.Done():
			return nil
		case e := <-ch:
			stream.Send(e)
		}
	}
}

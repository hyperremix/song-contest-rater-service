package server

import (
	"context"

	"github.com/hyperremix/song-contest-rater-service/db"
	"github.com/hyperremix/song-contest-rater-service/mapper"
	pb "github.com/hyperremix/song-contest-rater-service/protos/songcontestrater"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ratingServer struct {
	pb.UnimplementedRatingServer
	connPool *pgxpool.Pool
}

func NewRatingServer(connPool *pgxpool.Pool) pb.RatingServer {
	return &ratingServer{connPool: connPool}
}

func (s *ratingServer) ListRatings(ctx context.Context, request *emptypb.Empty) (*pb.ListRatingsResponse, error) {
	conn, err := s.connPool.Acquire(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not acquire connection: %v", err)
	}
	defer conn.Release()

	queries := db.New(conn)

	ratings, err := queries.ListRatings(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not list ratings: %v", err)
	}

	response, err := mapper.FromDbRatingListToResponse(ratings)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not convert ratings to proto: %v", err)
	}

	return response, nil
}

func (s *ratingServer) ListActRatings(ctx context.Context, request *pb.ListActRatingsRequest) (*pb.ListRatingsResponse, error) {
	conn, err := s.connPool.Acquire(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not acquire connection: %v", err)
	}
	defer conn.Release()

	queries := db.New(conn)
	actId, err := mapper.FromProtoToDbId(request.ActId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "could not map act id: %v", err)
	}

	ratings, err := queries.ListRatingsByActId(ctx, actId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not list ratings: %v", err)
	}

	response, err := mapper.FromDbRatingListToResponse(ratings)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not convert ratings to proto: %v", err)
	}

	return response, nil
}

func (s *ratingServer) ListUserRatings(ctx context.Context, request *pb.ListUserRatingsRequest) (*pb.ListRatingsResponse, error) {
	conn, err := s.connPool.Acquire(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not acquire connection: %v", err)
	}
	defer conn.Release()

	queries := db.New(conn)
	userId, err := mapper.FromProtoToDbId(request.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "could not map user id: %v", err)
	}

	ratings, err := queries.ListRatingsByUserId(ctx, userId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not list ratings: %v", err)
	}

	response, err := mapper.FromDbRatingListToResponse(ratings)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not convert ratings to proto: %v", err)
	}

	return response, nil
}

func (s *ratingServer) GetRating(ctx context.Context, request *pb.GetRatingRequest) (*pb.RatingResponse, error) {
	conn, err := s.connPool.Acquire(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not acquire connection: %v", err)
	}
	defer conn.Release()

	queries := db.New(conn)
	id, err := mapper.FromProtoToDbId(request.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "could not map id: %v", err)
	}

	rating, err := queries.GetRatingById(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not get rating: %v", err)
	}

	response, err := mapper.FromDbRatingToResponse(rating)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not convert rating to proto: %v", err)
	}

	return response, nil
}

func (s *ratingServer) CreateRating(ctx context.Context, request *pb.CreateRatingRequest) (*pb.RatingResponse, error) {
	conn, err := s.connPool.Acquire(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not acquire connection: %v", err)
	}
	defer conn.Release()

	queries := db.New(conn)
	insertRatingParams, err := mapper.FromCreateRequestToInsertRating(request)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "could not map create request: %v", err)
	}

	rating, err := queries.InsertRating(ctx, insertRatingParams)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not create rating: %v", err)
	}

	response, err := mapper.FromDbRatingToResponse(rating)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not convert rating to proto: %v", err)
	}

	return response, nil
}

func (s *ratingServer) UpdateRating(ctx context.Context, request *pb.UpdateRatingRequest) (*pb.RatingResponse, error) {
	conn, err := s.connPool.Acquire(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not acquire connection: %v", err)
	}
	defer conn.Release()

	queries := db.New(conn)
	updateRatingParams, err := mapper.FromUpdateRequestToUpdateRating(request)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "could not map update request: %v", err)
	}

	rating, err := queries.UpdateRating(ctx, updateRatingParams)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not update rating: %v", err)
	}

	response, err := mapper.FromDbRatingToResponse(rating)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not convert rating to proto: %v", err)
	}

	return response, nil
}

func (s *ratingServer) DeleteRating(ctx context.Context, request *pb.DeleteRatingRequest) (*pb.RatingResponse, error) {
	conn, err := s.connPool.Acquire(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not acquire connection: %v", err)
	}
	defer conn.Release()

	queries := db.New(conn)
	id, err := mapper.FromProtoToDbId(request.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "could not map id: %v", err)
	}

	rating, err := queries.DeleteRatingById(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not delete rating: %v", err)
	}

	response, err := mapper.FromDbRatingToResponse(rating)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not convert rating to proto: %v", err)
	}

	return response, nil
}

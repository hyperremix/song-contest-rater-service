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

type competitionServer struct {
	pb.UnimplementedCompetitionServer
	connPool *pgxpool.Pool
}

func NewCompetitionServer(connPool *pgxpool.Pool) pb.CompetitionServer {
	return &competitionServer{connPool: connPool}
}

func (s *competitionServer) ListCompetitions(ctx context.Context, request *emptypb.Empty) (*pb.ListCompetitionsResponse, error) {
	conn, err := s.connPool.Acquire(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not acquire connection: %v", err)
	}
	defer conn.Release()

	queries := db.New(conn)

	competitions, err := queries.ListCompetitions(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not list competitions: %v", err)
	}

	response, err := mapper.FromDbCompetitionListToResponse(competitions)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not convert competitions to proto: %v", err)
	}

	return response, nil
}
func (s *competitionServer) GetCompetition(ctx context.Context, request *pb.GetCompetitionRequest) (*pb.CompetitionResponse, error) {
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

	competition, err := queries.GetCompetitionById(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not get competition: %v", err)
	}

	response, err := mapper.FromDbCompetitionToResponse(competition)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not convert competition to proto: %v", err)
	}

	return response, nil
}
func (s *competitionServer) CreateCompetition(ctx context.Context, request *pb.CreateCompetitionRequest) (*pb.CompetitionResponse, error) {
	conn, err := s.connPool.Acquire(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not acquire connection: %v", err)
	}
	defer conn.Release()

	queries := db.New(conn)
	insertParams := mapper.FromCreateRequestToInsertCompetition(request)

	competition, err := queries.InsertCompetition(ctx, insertParams)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not insert competition: %v", err)
	}

	response, err := mapper.FromDbCompetitionToResponse(competition)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not convert competition to proto: %v", err)
	}

	return response, nil
}
func (s *competitionServer) UpdateCompetition(ctx context.Context, request *pb.UpdateCompetitionRequest) (*pb.CompetitionResponse, error) {
	conn, err := s.connPool.Acquire(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not acquire connection: %v", err)
	}
	defer conn.Release()

	queries := db.New(conn)
	updateParams, err := mapper.FromUpdateRequestToUpdateCompetition(request)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "could not map update params: %v", err)
	}

	competition, err := queries.UpdateCompetition(ctx, updateParams)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not update competition: %v", err)
	}

	response, err := mapper.FromDbCompetitionToResponse(competition)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not convert competition to proto: %v", err)
	}

	return response, nil
}
func (s *competitionServer) DeleteCompetition(ctx context.Context, request *pb.DeleteCompetitionRequest) (*pb.CompetitionResponse, error) {
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

	competition, err := queries.DeleteCompetitionById(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not delete competition: %v", err)
	}

	response, err := mapper.FromDbCompetitionToResponse(competition)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not convert competition to proto: %v", err)
	}

	return response, nil
}

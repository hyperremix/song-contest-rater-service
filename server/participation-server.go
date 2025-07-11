package server

import (
	"context"

	pb "github.com/hyperremix/song-contest-rater-protos/v4"
	"github.com/hyperremix/song-contest-rater-service/authz"
	"github.com/hyperremix/song-contest-rater-service/db"
	"github.com/hyperremix/song-contest-rater-service/mapper"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ParticipationServer struct {
	pb.UnimplementedParticipationServer
	queries  *db.Queries
	connPool *pgxpool.Pool
}

func NewParticipationServer(connPool *pgxpool.Pool) *ParticipationServer {
	return &ParticipationServer{
		queries:  db.New(connPool),
		connPool: connPool,
	}
}

func (s *ParticipationServer) ListParticipations(ctx context.Context, request *emptypb.Empty) (*pb.ListParticipationsResponse, error) {
	authUser := ctx.Value(authz.AuthUserContextKey{}).(*authz.AuthUser)
	if err := authUser.CheckIsAdmin(); err != nil {
		return nil, err
	}

	participations, err := s.queries.ListParticipations(ctx)
	if err != nil {
		return nil, err
	}

	response, err := mapper.FromManyParticipationsToProto(participations)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *ParticipationServer) CreateParticipation(ctx context.Context, request *pb.CreateParticipationRequest) (*pb.ParticipationResponse, error) {
	authUser := ctx.Value(authz.AuthUserContextKey{}).(*authz.AuthUser)
	if err := authUser.CheckIsAdmin(); err != nil {
		return nil, err
	}

	insertParams, err := mapper.FromCreateRequestToInsertParticipation(request)
	if err != nil {
		return nil, err
	}

	participation, err := s.queries.InsertParticipation(ctx, insertParams)
	if err != nil {
		return nil, err
	}

	response, err := mapper.FromParticipationToProto(&participation)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *ParticipationServer) DeleteParticipation(ctx context.Context, request *pb.DeleteParticipationRequest) (*pb.ParticipationResponse, error) {
	authUser := ctx.Value(authz.AuthUserContextKey{}).(*authz.AuthUser)
	if err := authUser.CheckIsAdmin(); err != nil {
		return nil, err
	}

	deleteParams, err := mapper.FromDeleteRequestToDeleteParticipation(request)
	if err != nil {
		return nil, err
	}

	participation, err := s.queries.DeleteParticipation(ctx, deleteParams)
	if err != nil {
		return nil, err
	}

	response, err := mapper.FromParticipationToProto(&participation)
	if err != nil {
		return nil, err
	}

	return response, nil
}

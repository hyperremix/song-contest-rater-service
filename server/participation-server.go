package server

import (
	"context"

	pb "buf.build/gen/go/hyperremix/song-contest-rater-protos/protocolbuffers/go/songcontestrater/v5"
	"connectrpc.com/connect"
	"github.com/hyperremix/song-contest-rater-service/authz"
	"github.com/hyperremix/song-contest-rater-service/db"
	"github.com/hyperremix/song-contest-rater-service/mapper"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ParticipationServer struct {
	queries  *db.Queries
	connPool *pgxpool.Pool
}

func NewParticipationServer(connPool *pgxpool.Pool) *ParticipationServer {
	return &ParticipationServer{
		queries:  db.New(connPool),
		connPool: connPool,
	}
}

func (s *ParticipationServer) ListParticipations(ctx context.Context, request *connect.Request[pb.ListParticipationsRequest]) (*connect.Response[pb.ListParticipationsResponse], error) {
	authUser := ctx.Value(authz.AuthUserContextKey).(*authz.AuthUser)
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

	return connect.NewResponse(&pb.ListParticipationsResponse{Participations: response}), nil
}

func (s *ParticipationServer) CreateParticipation(ctx context.Context, request *connect.Request[pb.CreateParticipationRequest]) (*connect.Response[pb.CreateParticipationResponse], error) {
	authUser := ctx.Value(authz.AuthUserContextKey).(*authz.AuthUser)
	if err := authUser.CheckIsAdmin(); err != nil {
		return nil, err
	}

	insertParams, err := mapper.FromCreateRequestToInsertParticipation(request.Msg)
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

	return connect.NewResponse(&pb.CreateParticipationResponse{Participation: response}), nil
}

func (s *ParticipationServer) DeleteParticipation(ctx context.Context, request *connect.Request[pb.DeleteParticipationRequest]) (*connect.Response[pb.DeleteParticipationResponse], error) {
	authUser := ctx.Value(authz.AuthUserContextKey).(*authz.AuthUser)
	if err := authUser.CheckIsAdmin(); err != nil {
		return nil, err
	}

	deleteParams, err := mapper.FromDeleteRequestToDeleteParticipation(request.Msg)
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

	return connect.NewResponse(&pb.DeleteParticipationResponse{Participation: response}), nil
}

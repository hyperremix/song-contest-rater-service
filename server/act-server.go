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

type ActServer struct {
	queries  *db.Queries
	connPool *pgxpool.Pool
}

func NewActServer(connPool *pgxpool.Pool) *ActServer {
	return &ActServer{
		queries:  db.New(connPool),
		connPool: connPool,
	}
}

func (s *ActServer) ListActs(ctx context.Context, request *connect.Request[pb.ListActsRequest]) (*connect.Response[pb.ListActsResponse], error) {
	acts, err := s.queries.ListActs(ctx)
	if err != nil {
		return nil, err
	}

	response, err := mapper.FromDbActListToResponse(acts, make([]db.Rating, 0), make([]db.User, 0))
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&pb.ListActsResponse{Acts: response}), nil
}

func (s *ActServer) GetAct(ctx context.Context, request *connect.Request[pb.GetActRequest]) (*connect.Response[pb.GetActResponse], error) {
	id, err := mapper.FromProtoToDbId(request.Msg.Id)
	if err != nil {
		return nil, err
	}

	act, err := s.queries.GetActById(ctx, id)
	if err != nil {
		return nil, err
	}

	ratings, err := s.queries.ListRatingsByActId(ctx, act.ID)
	if err != nil {
		return nil, err
	}

	users, err := s.queries.ListUsersByActId(ctx, act.ID)
	if err != nil {
		return nil, err
	}

	response, err := mapper.FromDbActToResponse(act, ratings, users)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&pb.GetActResponse{Act: response}), nil
}

func (s *ActServer) CreateAct(ctx context.Context, request *connect.Request[pb.CreateActRequest]) (*connect.Response[pb.CreateActResponse], error) {
	authUser := ctx.Value(authz.AuthUserContextKey).(*authz.AuthUser)
	if err := authUser.CheckIsAdmin(); err != nil {
		return nil, err
	}

	params, err := mapper.FromCreateRequestToInsertAct(request.Msg)
	if err != nil {
		return nil, err
	}

	act, err := s.queries.InsertAct(ctx, params)
	if err != nil {
		return nil, err
	}

	response, err := mapper.FromDbActToResponse(act, make([]db.Rating, 0), make([]db.User, 0))
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&pb.CreateActResponse{Act: response}), nil
}

func (s *ActServer) UpdateAct(ctx context.Context, request *connect.Request[pb.UpdateActRequest]) (*connect.Response[pb.UpdateActResponse], error) {
	authUser := ctx.Value(authz.AuthUserContextKey).(*authz.AuthUser)
	if err := authUser.CheckIsAdmin(); err != nil {
		return nil, err
	}

	updateParams, err := mapper.FromUpdateRequestToUpdateAct(request.Msg)
	if err != nil {
		return nil, err
	}

	act, err := s.queries.UpdateAct(ctx, updateParams)
	if err != nil {
		return nil, err
	}

	response, err := mapper.FromDbActToResponse(act, make([]db.Rating, 0), make([]db.User, 0))
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&pb.UpdateActResponse{Act: response}), nil
}

func (s *ActServer) DeleteAct(ctx context.Context, request *connect.Request[pb.DeleteActRequest]) (*connect.Response[pb.DeleteActResponse], error) {
	authUser := ctx.Value(authz.AuthUserContextKey).(*authz.AuthUser)
	if err := authUser.CheckIsAdmin(); err != nil {
		return nil, err
	}

	id, err := mapper.FromProtoToDbId(request.Msg.Id)
	if err != nil {
		return nil, err
	}

	act, err := s.queries.DeleteActById(ctx, id)
	if err != nil {
		return nil, err
	}

	response, err := mapper.FromDbActToResponse(act, make([]db.Rating, 0), make([]db.User, 0))
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&pb.DeleteActResponse{Act: response}), nil
}

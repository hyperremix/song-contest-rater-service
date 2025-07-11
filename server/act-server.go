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

type ActServer struct {
	pb.UnimplementedActServer
	queries  *db.Queries
	connPool *pgxpool.Pool
}

func NewActServer(connPool *pgxpool.Pool) *ActServer {
	return &ActServer{
		queries:  db.New(connPool),
		connPool: connPool,
	}
}

func (s *ActServer) ListActs(ctx context.Context, request *emptypb.Empty) (*pb.ListActsResponse, error) {
	acts, err := s.queries.ListActs(ctx)
	if err != nil {
		return nil, err
	}

	response, err := mapper.FromDbActListToResponse(acts, make([]db.Rating, 0), make([]db.User, 0))
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *ActServer) GetAct(ctx context.Context, request *pb.GetActRequest) (*pb.ActResponse, error) {
	id, err := mapper.FromProtoToDbId(request.Id)
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

	return response, nil
}

func (s *ActServer) CreateAct(ctx context.Context, request *pb.CreateActRequest) (*pb.ActResponse, error) {
	authUser := ctx.Value(authz.AuthUserContextKey{}).(*authz.AuthUser)
	if err := authUser.CheckIsAdmin(); err != nil {
		return nil, err
	}

	params, err := mapper.FromCreateRequestToInsertAct(request)
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

	return response, nil
}

func (s *ActServer) UpdateAct(ctx context.Context, request *pb.UpdateActRequest) (*pb.ActResponse, error) {
	authUser := ctx.Value(authz.AuthUserContextKey{}).(*authz.AuthUser)
	if err := authUser.CheckIsAdmin(); err != nil {
		return nil, err
	}

	updateParams, err := mapper.FromUpdateRequestToUpdateAct(request)
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

	return response, nil
}

func (s *ActServer) DeleteAct(ctx context.Context, request *pb.DeleteActRequest) (*pb.ActResponse, error) {
	authUser := ctx.Value(authz.AuthUserContextKey{}).(*authz.AuthUser)
	if err := authUser.CheckIsAdmin(); err != nil {
		return nil, err
	}

	id, err := mapper.FromProtoToDbId(request.Id)
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

	return response, nil
}

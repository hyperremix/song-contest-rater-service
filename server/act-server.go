package server

import (
	"context"

	"github.com/hyperremix/song-contest-rater-service/authz"
	"github.com/hyperremix/song-contest-rater-service/db"
	"github.com/hyperremix/song-contest-rater-service/mapper"
	"github.com/hyperremix/song-contest-rater-service/permission"
	pb "github.com/hyperremix/song-contest-rater-service/protos/songcontestrater"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type actServer struct {
	pb.UnimplementedActServer
	connPool *pgxpool.Pool
}

func NewActServer(connPool *pgxpool.Pool) pb.ActServer {
	return &actServer{connPool: connPool}
}

func (s *actServer) ListActs(ctx context.Context, request *emptypb.Empty) (*pb.ListActsResponse, error) {
	conn, err := s.connPool.Acquire(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not acquire connection: %v", err)
	}
	defer conn.Release()

	queries := db.New(conn)

	acts, err := queries.ListActs(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not list acts: %v", err)
	}

	response, err := mapper.FromDbActListToResponse(acts)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not convert acts to proto: %v", err)
	}

	return response, nil
}

func (s *actServer) GetAct(ctx context.Context, request *pb.GetActRequest) (*pb.ActResponse, error) {
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

	act, err := queries.GetActById(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not get act: %v", err)
	}

	response, err := mapper.FromDbActToResponse(act)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not convert act to proto: %v", err)
	}

	return response, nil
}

func (s *actServer) CreateAct(ctx context.Context, request *pb.CreateActRequest) (*pb.ActResponse, error) {
	authUser := ctx.Value(authz.AuthUserContextKey{}).(*authz.AuthUser)
	if !authUser.HasPermission(permission.WriteActs) {
		return nil, status.Errorf(codes.PermissionDenied, "missing permission: %s", permission.WriteActs)
	}

	conn, err := s.connPool.Acquire(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not acquire connection: %v", err)
	}
	defer conn.Release()

	queries := db.New(conn)
	params, err := mapper.FromCreateRequestToInsertAct(request)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "could not map request: %v", err)
	}

	act, err := queries.InsertAct(ctx, params)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not create act: %v", err)
	}

	response, err := mapper.FromDbActToResponse(act)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not convert act to proto: %v", err)
	}

	return response, nil
}

func (s *actServer) UpdateAct(ctx context.Context, request *pb.UpdateActRequest) (*pb.ActResponse, error) {
	authUser := ctx.Value(authz.AuthUserContextKey{}).(*authz.AuthUser)
	if !authUser.HasPermission(permission.WriteActs) {
		return nil, status.Errorf(codes.PermissionDenied, "missing permission: %s", permission.WriteActs)
	}

	conn, err := s.connPool.Acquire(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not acquire connection: %v", err)
	}
	defer conn.Release()

	queries := db.New(conn)
	updateParams, err := mapper.FromUpdateRequestToUpdateAct(request)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "could not map update params: %v", err)
	}

	act, err := queries.UpdateAct(ctx, updateParams)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not update act: %v", err)
	}

	response, err := mapper.FromDbActToResponse(act)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not convert act to proto: %v", err)
	}

	return response, nil
}

func (s *actServer) DeleteAct(ctx context.Context, request *pb.DeleteActRequest) (*pb.ActResponse, error) {
	authUser := ctx.Value(authz.AuthUserContextKey{}).(*authz.AuthUser)
	if !authUser.HasPermission(permission.WriteActs) {
		return nil, status.Errorf(codes.PermissionDenied, "missing permission: %s", permission.WriteActs)
	}

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

	act, err := queries.DeleteActById(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not delete act: %v", err)
	}

	response, err := mapper.FromDbActToResponse(act)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not convert act to proto: %v", err)
	}

	return response, nil
}

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

type userServer struct {
	pb.UnimplementedUserServer
	connPool *pgxpool.Pool
}

func NewUserServer(connPool *pgxpool.Pool) pb.UserServer {
	return &userServer{connPool: connPool}
}

func (s *userServer) ListUsers(ctx context.Context, request *emptypb.Empty) (*pb.ListUsersResponse, error) {
	conn, err := s.connPool.Acquire(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not acquire connection: %v", err)
	}
	defer conn.Release()

	queries := db.New(conn)

	users, err := queries.ListUsers(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not list users: %v", err)
	}

	response, err := mapper.FromDbUserListToResponse(users)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not convert users to proto: %v", err)
	}

	return response, nil
}

func (s *userServer) GetUser(ctx context.Context, request *pb.GetUserRequest) (*pb.UserResponse, error) {
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

	user, err := queries.GetUserById(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not get user: %v", err)
	}

	response, err := mapper.FromDbUserToResponse(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not convert user to proto: %v", err)
	}

	return response, nil
}

func (s *userServer) CreateUser(ctx context.Context, request *pb.CreateUserRequest) (*pb.UserResponse, error) {
	conn, err := s.connPool.Acquire(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not acquire connection: %v", err)
	}
	defer conn.Release()

	queries := db.New(conn)
	insertParams, err := mapper.FromCreateRequestToInsertUser(request)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "could not map request: %v", err)
	}

	user, err := queries.InsertUser(ctx, insertParams)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not insert user: %v", err)
	}

	response, err := mapper.FromDbUserToResponse(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not convert user to proto: %v", err)
	}

	return response, nil
}

func (s *userServer) UpdateUser(ctx context.Context, request *pb.UpdateUserRequest) (*pb.UserResponse, error) {
	conn, err := s.connPool.Acquire(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not acquire connection: %v", err)
	}
	defer conn.Release()

	queries := db.New(conn)
	updateParams, err := mapper.FromUpdateRequestToUpdateUser(request)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "could not map request: %v", err)
	}

	user, err := queries.UpdateUser(ctx, updateParams)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not update user: %v", err)
	}

	response, err := mapper.FromDbUserToResponse(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not convert user to proto: %v", err)
	}

	return response, nil
}

func (s *userServer) DeleteUser(ctx context.Context, request *pb.DeleteUserRequest) (*pb.UserResponse, error) {
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

	user, err := queries.DeleteUserById(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not delete user: %v", err)
	}

	response, err := mapper.FromDbUserToResponse(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not convert user to proto: %v", err)
	}

	return response, nil
}

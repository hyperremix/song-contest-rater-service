package server

import (
	"context"
	"errors"

	pb "github.com/hyperremix/song-contest-rater-protos/v4"
	"github.com/hyperremix/song-contest-rater-service/authz"
	"github.com/hyperremix/song-contest-rater-service/db"
	"github.com/hyperremix/song-contest-rater-service/mapper"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UserServer struct {
	pb.UnimplementedUserServer
	queries  *db.Queries
	connPool *pgxpool.Pool
}

func NewUserServer(connPool *pgxpool.Pool) *UserServer {
	return &UserServer{
		queries:  db.New(connPool),
		connPool: connPool,
	}
}

func (s *UserServer) ListUsers(ctx context.Context, request *emptypb.Empty) (*pb.ListUsersResponse, error) {
	users, err := s.queries.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	response, err := mapper.FromDbUserListToResponse(users)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *UserServer) GetUser(ctx context.Context, request *pb.GetUserRequest) (*pb.UserResponse, error) {
	id, err := mapper.FromProtoToDbId(request.Id)
	if err != nil {
		return nil, err
	}

	user, err := s.queries.GetUserById(ctx, id)
	if err != nil {
		return nil, err
	}

	response, err := mapper.FromDbUserToResponse(user)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *UserServer) GetAuthUser(ctx context.Context, request *emptypb.Empty) (*pb.UserResponse, error) {
	authUser := ctx.Value(authz.AuthUserContextKey{}).(*authz.AuthUser)

	user, err := s.queries.GetUserBySub(ctx, authUser.ClerkUser.ID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	if err != nil {
		return nil, err
	}

	response, err := mapper.FromDbUserToResponse(user)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *UserServer) CreateUser(ctx context.Context, request *pb.CreateUserRequest) (*pb.UserResponse, error) {
	authUser := ctx.Value(authz.AuthUserContextKey{}).(*authz.AuthUser)

	insertParams, err := mapper.FromCreateRequestToInsertUser(request, authUser.ClerkUser.ID)
	if err != nil {
		return nil, err
	}

	user, err := s.queries.InsertUser(ctx, insertParams)
	if err != nil {
		return nil, err
	}

	response, err := mapper.FromDbUserToResponse(user)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *UserServer) UpdateUser(ctx context.Context, request *pb.UpdateUserRequest) (*pb.UserResponse, error) {
	authUser := ctx.Value(authz.AuthUserContextKey{}).(*authz.AuthUser)

	if err := authUser.CheckIsAdmin(); err != nil && authUser.UserID != request.Id {
		return nil, err
	}

	updateParams, err := mapper.FromUpdateRequestToUpdateUser(request)
	if err != nil {
		return nil, err
	}

	user, err := s.queries.UpdateUser(ctx, updateParams)
	if err != nil {
		return nil, err
	}

	response, err := mapper.FromDbUserToResponse(user)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *UserServer) DeleteUser(ctx context.Context, request *pb.DeleteUserRequest) (*pb.UserResponse, error) {
	authUser := ctx.Value(authz.AuthUserContextKey{}).(*authz.AuthUser)

	if err := authUser.CheckIsAdmin(); err != nil && authUser.UserID != request.Id {
		return nil, err
	}

	id, err := mapper.FromProtoToDbId(request.Id)
	if err != nil {
		return nil, err
	}

	user, err := s.queries.DeleteUserById(ctx, id)
	if err != nil {
		return nil, err
	}

	response, err := mapper.FromDbUserToResponse(user)
	if err != nil {
		return nil, err
	}

	return response, nil
}

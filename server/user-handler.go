package server

import (
	"context"
	"errors"

	pb "buf.build/gen/go/hyperremix/song-contest-rater-protos/protocolbuffers/go/songcontestrater/v5"
	"connectrpc.com/connect"
	"github.com/hyperremix/song-contest-rater-service/authz"
	"github.com/hyperremix/song-contest-rater-service/db"
	"github.com/hyperremix/song-contest-rater-service/mapper"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserServer struct {
	queries  *db.Queries
	connPool *pgxpool.Pool
}

func NewUserServer(connPool *pgxpool.Pool) *UserServer {
	return &UserServer{
		queries:  db.New(connPool),
		connPool: connPool,
	}
}

func (s *UserServer) ListUsers(ctx context.Context, request *connect.Request[pb.ListUsersRequest]) (*connect.Response[pb.ListUsersResponse], error) {
	users, err := s.queries.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	response, err := mapper.FromDbUserListToResponse(users)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&pb.ListUsersResponse{Users: response}), nil
}

func (s *UserServer) GetUser(ctx context.Context, request *connect.Request[pb.GetUserRequest]) (*connect.Response[pb.GetUserResponse], error) {
	id, err := mapper.FromProtoToDbId(request.Msg.Id)
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

	return connect.NewResponse(&pb.GetUserResponse{User: response}), nil
}

func (s *UserServer) GetAuthUser(ctx context.Context, request *connect.Request[pb.GetUserRequest]) (*connect.Response[pb.GetUserResponse], error) {
	authUser := ctx.Value(authz.AuthUserContextKey).(*authz.AuthUser)

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

	return connect.NewResponse(&pb.GetUserResponse{User: response}), nil
}

func (s *UserServer) CreateUser(ctx context.Context, request *connect.Request[pb.CreateUserRequest]) (*connect.Response[pb.CreateUserResponse], error) {
	authUser := ctx.Value(authz.AuthUserContextKey).(*authz.AuthUser)

	insertParams, err := mapper.FromCreateRequestToInsertUser(request.Msg, authUser.ClerkUser.ID)
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

	return connect.NewResponse(&pb.CreateUserResponse{User: response}), nil
}

func (s *UserServer) UpdateUser(ctx context.Context, request *connect.Request[pb.UpdateUserRequest]) (*connect.Response[pb.UpdateUserResponse], error) {
	authUser := ctx.Value(authz.AuthUserContextKey).(*authz.AuthUser)

	if err := authUser.CheckIsAdmin(); err != nil && authUser.UserID != request.Msg.Id {
		return nil, err
	}

	updateParams, err := mapper.FromUpdateRequestToUpdateUser(request.Msg)
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

	return connect.NewResponse(&pb.UpdateUserResponse{User: response}), nil
}

func (s *UserServer) DeleteUser(ctx context.Context, request *connect.Request[pb.DeleteUserRequest]) (*connect.Response[pb.DeleteUserResponse], error) {
	authUser := ctx.Value(authz.AuthUserContextKey).(*authz.AuthUser)

	if err := authUser.CheckIsAdmin(); err != nil && authUser.UserID != request.Msg.Id {
		return nil, err
	}

	id, err := mapper.FromProtoToDbId(request.Msg.Id)
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

	return connect.NewResponse(&pb.DeleteUserResponse{User: response}), nil
}

package server

import (
	"context"
	"database/sql"
	"errors"

	pb "buf.build/gen/go/hyperremix/song-contest-rater-protos/protocolbuffers/go/songcontestrater/v5"
	"connectrpc.com/connect"
	"github.com/hyperremix/song-contest-rater-service/authz"
	"github.com/hyperremix/song-contest-rater-service/db"
	"github.com/hyperremix/song-contest-rater-service/mapper"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StatServer struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

func NewStatServer(pool *pgxpool.Pool) *StatServer {
	return &StatServer{
		queries: db.New(pool),
		pool:    pool,
	}
}

func (s *StatServer) ListUserStats(ctx context.Context, request *connect.Request[pb.ListUserStatsRequest]) (*connect.Response[pb.ListUserStatsResponse], error) {
	usersStats, err := s.queries.ListUserStats(ctx)
	if err != nil {
		return nil, err
	}

	globalStats, err := s.queries.GetGlobalStats(ctx)
	if err != nil {
		return nil, err
	}

	users, err := s.queries.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	response, err := mapper.FromDbUserStatListToResponse(usersStats, globalStats, users)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&pb.ListUserStatsResponse{Stats: response}), nil
}

func (s *StatServer) GetMyStats(ctx context.Context, request *connect.Request[pb.GetMyStatsRequest]) (*connect.Response[pb.GetMyStatsResponse], error) {
	authUser := ctx.Value(authz.AuthUserContextKey).(*authz.AuthUser)
	userId, err := mapper.FromProtoToDbId(authUser.UserID)
	if err != nil {
		return nil, err
	}

	userStats, err := s.queries.GetStatsByUserId(ctx, userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return connect.NewResponse(&pb.GetMyStatsResponse{Stats: mapper.EmptyUserStatsResponse()}), nil
		}
		return nil, err
	}

	globalStats, err := s.queries.GetGlobalStats(ctx)
	if err != nil {
		return nil, err
	}

	response, err := mapper.FromDbUserStatsToResponse(userStats, globalStats, &authUser.DbUser)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&pb.GetMyStatsResponse{Stats: response}), nil
}

func (s *StatServer) GetGlobalStats(ctx context.Context, request *connect.Request[pb.GetGlobalStatsRequest]) (*connect.Response[pb.GetGlobalStatsResponse], error) {
	globalStats, err := s.queries.GetGlobalStats(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return connect.NewResponse(&pb.GetGlobalStatsResponse{Stats: mapper.EmptyGlobalStatsResponse()}), nil
		}
		return nil, err
	}

	response, err := mapper.FromDbGlobalStatsToResponse(globalStats)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&pb.GetGlobalStatsResponse{Stats: response}), nil
}

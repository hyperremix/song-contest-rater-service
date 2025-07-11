package server

import (
	"context"
	"database/sql"
	"errors"

	pb "github.com/hyperremix/song-contest-rater-protos/v4"
	"github.com/hyperremix/song-contest-rater-service/authz"
	"github.com/hyperremix/song-contest-rater-service/db"
	"github.com/hyperremix/song-contest-rater-service/mapper"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/emptypb"
)

type StatServer struct {
	pb.UnimplementedStatServer
	queries *db.Queries
	pool    *pgxpool.Pool
}

func NewStatServer(pool *pgxpool.Pool) *StatServer {
	return &StatServer{
		queries: db.New(pool),
		pool:    pool,
	}
}

func (s *StatServer) ListUserStats(ctx context.Context, request *emptypb.Empty) (*pb.ListUserStatsResponse, error) {
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

	return response, nil
}

func (s *StatServer) GetMyStats(ctx context.Context, request *emptypb.Empty) (*pb.UserStatsResponse, error) {
	authUser := ctx.Value(authz.AuthUserContextKey{}).(*authz.AuthUser)
	userId, err := mapper.FromProtoToDbId(authUser.UserID)
	if err != nil {
		return nil, err
	}

	userStats, err := s.queries.GetStatsByUserId(ctx, userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return mapper.EmptyUserStatsResponse(), nil
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

	return response, nil
}

func (s *StatServer) GetGlobalStats(ctx context.Context, request *emptypb.Empty) (*pb.GlobalStatsResponse, error) {
	globalStats, err := s.queries.GetGlobalStats(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return mapper.EmptyGlobalStatsResponse(), nil
		}
		return nil, err
	}

	response, err := mapper.FromDbGlobalStatsToResponse(globalStats)
	if err != nil {
		return nil, err
	}

	return response, nil
}

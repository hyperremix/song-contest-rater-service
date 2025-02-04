package stat

import (
	"context"
	"errors"

	"github.com/hyperremix/song-contest-rater-service/db"
	"github.com/hyperremix/song-contest-rater-service/mapper"
	pb "github.com/hyperremix/song-contest-rater-service/protos/songcontestrater"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{
		pool:    pool,
		queries: db.New(pool),
	}
}

func (s *Service) UpsertRatingStats(ctx context.Context, rating *pb.RatingResponse) error {
	err := s.upsertUserStats(ctx, rating)
	if err != nil {
		return err
	}

	return s.upsertGlobalStats(ctx, rating)
}

// TODO: How to handle stats for creating (done), updating (hard) and deleting (easy) ratings?

func (s *Service) upsertUserStats(ctx context.Context, rating *pb.RatingResponse) error {
	userId, err := mapper.FromProtoToDbId(rating.User.Id)
	if err != nil {
		return err
	}

	userStats, err := s.queries.GetStatsByUserId(ctx, userId)
	if errors.Is(err, pgx.ErrNoRows) {
		upsertParams := mapper.FromRatingToDbUserStats(rating)
		_, err = s.queries.UpsertUserStats(ctx, upsertParams)
		if err != nil {
			return err
		}

		return nil
	}

	if err != nil {
		return err
	}

	updatedUpsertParams, err := mapper.FromStatsToUpsertUserStats(rating, userStats)
	if err != nil {
		return err
	}

	_, err = s.queries.UpsertUserStats(ctx, updatedUpsertParams)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) upsertGlobalStats(ctx context.Context, rating *pb.RatingResponse) error {
	globalStats, err := s.queries.GetGlobalStats(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		upsertParams := mapper.FromRatingToDbGlobalStats(rating)
		_, err = s.queries.UpsertGlobalStats(ctx, upsertParams)
		if err != nil {
			return err
		}

		return nil
	}

	if err != nil {
		return err
	}

	updatedUpsertParams, err := mapper.FromStatsToUpsertGlobalStats(rating, globalStats)
	if err != nil {
		return err
	}

	_, err = s.queries.UpsertGlobalStats(ctx, updatedUpsertParams)
	if err != nil {
		return err
	}

	return nil
}

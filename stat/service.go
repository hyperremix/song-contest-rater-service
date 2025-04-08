package stat

import (
	"context"
	"errors"

	pb "github.com/hyperremix/song-contest-rater-protos/v4"
	"github.com/hyperremix/song-contest-rater-service/db"
	"github.com/hyperremix/song-contest-rater-service/mapper"
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

func (s *Service) AddRatingToStats(ctx context.Context, rating *pb.RatingResponse) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
	}()

	queries := s.queries.WithTx(tx)

	err = s.addToUserStats(ctx, queries, rating)
	if err != nil {
		return err
	}

	return s.addToGlobalStats(ctx, queries, rating)
}

func (s *Service) UpdateRatingInStats(ctx context.Context, rating *pb.RatingResponse) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
	}()

	queries := s.queries.WithTx(tx)

	err = s.updateUserStats(ctx, queries, rating)
	if err != nil {
		return err
	}

	return s.updateGlobalStats(ctx, queries, rating)
}

func (s *Service) RemoveRatingFromStats(ctx context.Context, rating *pb.RatingResponse) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
	}()

	queries := s.queries.WithTx(tx)

	err = s.removeFromUserStats(ctx, queries, rating)
	if err != nil {
		return err
	}

	return s.removeFromGlobalStats(ctx, queries, rating)
}

func (s *Service) addToUserStats(ctx context.Context, queries *db.Queries, rating *pb.RatingResponse) error {
	userId, err := mapper.FromProtoToDbId(rating.User.Id)
	if err != nil {
		return err
	}

	userStats, err := queries.GetStatsByUserId(ctx, userId)
	if errors.Is(err, pgx.ErrNoRows) {
		upsertParams := mapper.FromRatingToDbUserStats(rating)
		_, err = queries.UpsertUserStats(ctx, upsertParams)
		if err != nil {
			return err
		}

		return nil
	}

	if err != nil {
		return err
	}

	upsertParams, err := mapper.AddToUserStats(rating, userStats)
	if err != nil {
		return err
	}

	_, err = queries.UpsertUserStats(ctx, upsertParams)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) updateUserStats(ctx context.Context, queries *db.Queries, rating *pb.RatingResponse) error {
	userId, err := mapper.FromProtoToDbId(rating.User.Id)
	if err != nil {
		return err
	}

	userStats, err := queries.GetStatsByUserId(ctx, userId)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	ratingId, err := mapper.FromProtoToDbId(rating.Id)
	if err != nil {
		return err
	}

	oldDbRating, err := queries.GetRatingById(ctx, ratingId)
	if err != nil {
		return err
	}

	oldRating, err := mapper.FromDbRatingToResponse(oldDbRating, nil)
	if err != nil {
		return err
	}

	updatedUpsertParams, err := mapper.UpdateUserStats(rating, oldRating, userStats)
	if err != nil {
		return err
	}

	_, err = queries.UpsertUserStats(ctx, updatedUpsertParams)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) removeFromUserStats(ctx context.Context, queries *db.Queries, rating *pb.RatingResponse) error {
	userId, err := mapper.FromProtoToDbId(rating.User.Id)
	if err != nil {
		return err
	}

	userStats, err := queries.GetStatsByUserId(ctx, userId)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}

	updatedUpsertParams, err := mapper.RemoveFromUserStats(rating, userStats)
	if err != nil {
		return err
	}

	_, err = queries.UpsertUserStats(ctx, updatedUpsertParams)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) addToGlobalStats(ctx context.Context, queries *db.Queries, rating *pb.RatingResponse) error {
	globalStats, err := queries.GetGlobalStats(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		upsertParams := mapper.FromRatingToDbGlobalStats(rating)
		_, err = queries.UpsertGlobalStats(ctx, upsertParams)
		if err != nil {
			return err
		}

		return nil
	}

	if err != nil {
		return err
	}

	updatedUpsertParams, err := mapper.AddToGlobalStats(rating, globalStats)
	if err != nil {
		return err
	}

	_, err = queries.UpsertGlobalStats(ctx, updatedUpsertParams)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) updateGlobalStats(ctx context.Context, queries *db.Queries, rating *pb.RatingResponse) error {
	globalStats, err := queries.GetGlobalStats(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}

	if err != nil {
		return err
	}

	ratingId, err := mapper.FromProtoToDbId(rating.Id)
	if err != nil {
		return err
	}

	oldDbRating, err := queries.GetRatingById(ctx, ratingId)
	if err != nil {
		return err
	}

	oldRating, err := mapper.FromDbRatingToResponse(oldDbRating, nil)
	if err != nil {
		return err
	}

	updatedUpsertParams, err := mapper.UpdateGlobalStats(rating, oldRating, globalStats)
	if err != nil {
		return err
	}

	_, err = queries.UpsertGlobalStats(ctx, updatedUpsertParams)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) removeFromGlobalStats(ctx context.Context, queries *db.Queries, rating *pb.RatingResponse) error {
	globalStats, err := queries.GetGlobalStats(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}

	if err != nil {
		return err
	}

	updatedUpsertParams, err := mapper.RemoveFromGlobalStats(rating, globalStats)
	if err != nil {
		return err
	}

	_, err = queries.UpsertGlobalStats(ctx, updatedUpsertParams)
	if err != nil {
		return err
	}

	return nil
}

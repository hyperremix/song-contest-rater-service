package mapper

import (
	"github.com/hyperremix/song-contest-rater-service/db"
	pb "github.com/hyperremix/song-contest-rater-service/protos/songcontestrater"
	"github.com/jackc/pgx/v5/pgtype"
)

func FromDbUserStatsToResponse(stats db.UserStat, globalStats db.GlobalStat) (*pb.UserStatsResponse, error) {
	userId, err := FromDbToProtoId(stats.UserID)
	if err != nil {
		return nil, NewResponseBindingError(err)
	}

	userRatingAvg, err := fromNumericToFloat64(stats.RatingAvg)
	if err != nil {
		return nil, NewResponseBindingError(err)
	}

	globalRatingAvg, err := fromNumericToFloat64(globalStats.RatingAvg)
	if err != nil {
		return nil, NewResponseBindingError(err)
	}

	return &pb.UserStatsResponse{
		UserId:        userId,
		UserRatingAvg: userRatingAvg,
		TotalRatings:  stats.RatingCount.Int32,
		RatingBias:    globalRatingAvg - userRatingAvg,
		CriticType:    fromRatingBiasToCriticType(globalRatingAvg - userRatingAvg),
		CreatedAt:     fromDbToProtoTimestamp(stats.CreatedAt),
		UpdatedAt:     fromDbToProtoTimestamp(stats.UpdatedAt),
	}, nil
}

func FromDbGlobalStatsToResponse(stats db.GlobalStat) (*pb.GlobalStatsResponse, error) {
	globalRatingAvg, err := fromNumericToFloat64(stats.RatingAvg)
	if err != nil {
		return nil, NewResponseBindingError(err)
	}

	return &pb.GlobalStatsResponse{
		GlobalRatingAvg: globalRatingAvg,
		TotalRatings:    stats.RatingCount.Int32,
		CreatedAt:       fromDbToProtoTimestamp(stats.CreatedAt),
		UpdatedAt:       fromDbToProtoTimestamp(stats.UpdatedAt),
	}, nil
}

func FromRatingToDbUserStats(rating *pb.RatingResponse) db.UpsertUserStatsParams {
	userId, err := FromProtoToDbId(rating.User.Id)
	if err != nil {
		return db.UpsertUserStatsParams{}
	}

	return db.UpsertUserStatsParams{
		UserID:      userId,
		RatingAvg:   fromFloat64ToNumeric(float64(rating.Total) / 5),
		RatingCount: fromInt32ToInt4(1),
	}
}

func FromRatingToDbGlobalStats(rating *pb.RatingResponse) db.UpsertGlobalStatsParams {
	return db.UpsertGlobalStatsParams{
		RatingAvg:   fromFloat64ToNumeric(float64(rating.Total) / 5),
		RatingCount: fromInt32ToInt4(1),
	}
}

func AddToUserStats(newRating *pb.RatingResponse, userStats db.UserStat) (db.UpsertUserStatsParams, error) {
	newAvg, newCount, err := calculateAddedAverage(
		userStats.RatingAvg,
		userStats.RatingCount.Int32,
		float64(newRating.Total),
	)
	if err != nil {
		return db.UpsertUserStatsParams{}, err
	}

	return db.UpsertUserStatsParams{
		UserID:      userStats.UserID,
		RatingAvg:   newAvg,
		RatingCount: newCount,
	}, nil
}

func UpdateUserStats(newRating *pb.RatingResponse, oldRating *pb.RatingResponse, userStats db.UserStat) (db.UpsertUserStatsParams, error) {
	newAvg, err := calculateUpdatedAverage(
		userStats.RatingAvg,
		userStats.RatingCount.Int32,
		float64(oldRating.Total),
		float64(newRating.Total),
	)
	if err != nil {
		return db.UpsertUserStatsParams{}, err
	}

	return db.UpsertUserStatsParams{
		UserID:      userStats.UserID,
		RatingAvg:   newAvg,
		RatingCount: userStats.RatingCount,
	}, nil
}

func RemoveFromUserStats(rating *pb.RatingResponse, userStats db.UserStat) (db.UpsertUserStatsParams, error) {
	newAvg, newCount, err := calculateRemovedAverage(
		userStats.RatingAvg,
		userStats.RatingCount.Int32,
		float64(rating.Total),
	)
	if err != nil {
		return db.UpsertUserStatsParams{}, err
	}

	return db.UpsertUserStatsParams{
		UserID:      userStats.UserID,
		RatingAvg:   newAvg,
		RatingCount: newCount,
	}, nil
}

func AddToGlobalStats(newRating *pb.RatingResponse, globalStats db.GlobalStat) (db.UpsertGlobalStatsParams, error) {
	newAvg, newCount, err := calculateAddedAverage(
		globalStats.RatingAvg,
		globalStats.RatingCount.Int32,
		float64(newRating.Total),
	)
	if err != nil {
		return db.UpsertGlobalStatsParams{}, err
	}

	return db.UpsertGlobalStatsParams{
		RatingAvg:   newAvg,
		RatingCount: newCount,
	}, nil
}

func UpdateGlobalStats(newRating *pb.RatingResponse, oldRating *pb.RatingResponse, globalStats db.GlobalStat) (db.UpsertGlobalStatsParams, error) {
	newAvg, err := calculateUpdatedAverage(
		globalStats.RatingAvg,
		globalStats.RatingCount.Int32,
		float64(oldRating.Total),
		float64(newRating.Total),
	)
	if err != nil {
		return db.UpsertGlobalStatsParams{}, err
	}

	return db.UpsertGlobalStatsParams{
		RatingAvg:   newAvg,
		RatingCount: globalStats.RatingCount,
	}, nil
}

func RemoveFromGlobalStats(rating *pb.RatingResponse, globalStats db.GlobalStat) (db.UpsertGlobalStatsParams, error) {
	newAvg, newCount, err := calculateRemovedAverage(
		globalStats.RatingAvg,
		globalStats.RatingCount.Int32,
		float64(rating.Total),
	)
	if err != nil {
		return db.UpsertGlobalStatsParams{}, err
	}

	return db.UpsertGlobalStatsParams{
		RatingAvg:   newAvg,
		RatingCount: newCount,
	}, nil
}

func fromRatingBiasToCriticType(ratingBias float64) pb.CriticType {
	switch {
	case ratingBias <= -1.0:
		return pb.CriticType_CRITIC_TYPE_HARSH
	case ratingBias <= -0.5:
		return pb.CriticType_CRITIC_TYPE_SLIGHTLY_CRITICAL
	case ratingBias >= 1.0:
		return pb.CriticType_CRITIC_TYPE_GENEROUS
	case ratingBias >= 0.5:
		return pb.CriticType_CRITIC_TYPE_EASY_TO_PLEASE
	default:
		return pb.CriticType_CRITIC_TYPE_BALANCED
	}
}

func calculateAddedAverage(currentAvg pgtype.Numeric, count int32, newValue float64) (pgtype.Numeric, pgtype.Int4, error) {
	currentAvgFloat, err := fromNumericToFloat64(currentAvg)
	if err != nil {
		return pgtype.Numeric{}, pgtype.Int4{}, err
	}

	newCount := count + 1
	newAvg := (currentAvgFloat*float64(count) + newValue) / float64(newCount)
	return fromFloat64ToNumeric(newAvg), fromInt32ToInt4(newCount), nil
}

func calculateUpdatedAverage(dbCurrentAvg pgtype.Numeric, count int32, oldValue, newValue float64) (pgtype.Numeric, error) {
	currentAvg, err := fromNumericToFloat64(dbCurrentAvg)
	if err != nil {
		return pgtype.Numeric{}, err
	}

	return fromFloat64ToNumeric((currentAvg*float64(count) - oldValue + newValue) / float64(count)), nil
}

func calculateRemovedAverage(dbCurrentAvg pgtype.Numeric, count int32, valueToRemove float64) (pgtype.Numeric, pgtype.Int4, error) {
	currentAvg, err := fromNumericToFloat64(dbCurrentAvg)
	if err != nil {
		return pgtype.Numeric{}, pgtype.Int4{}, err
	}

	newCount := count - 1
	return fromFloat64ToNumeric((currentAvg*float64(count) - valueToRemove) / float64(newCount)), fromInt32ToInt4(newCount), nil
}

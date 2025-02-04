package mapper

import (
	"github.com/hyperremix/song-contest-rater-service/db"
	pb "github.com/hyperremix/song-contest-rater-service/protos/songcontestrater"
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

func FromStatsToUpsertUserStats(rating *pb.RatingResponse, userStats db.UserStat) (db.UpsertUserStatsParams, error) {
	newRatingCount := userStats.RatingCount.Int32 + 1
	newRatingAvg, err := fromNumericToFloat64(userStats.RatingAvg)
	if err != nil {
		return db.UpsertUserStatsParams{}, err
	}

	newRatingAvg = (newRatingAvg*float64(userStats.RatingCount.Int32) + float64(rating.Total)) / float64(newRatingCount)

	return db.UpsertUserStatsParams{
		UserID:      userStats.UserID,
		RatingAvg:   fromFloat64ToNumeric(newRatingAvg),
		RatingCount: fromInt32ToInt4(newRatingCount),
	}, nil
}

func FromStatsToUpsertGlobalStats(rating *pb.RatingResponse, globalStats db.GlobalStat) (db.UpsertGlobalStatsParams, error) {
	newRatingCount := globalStats.RatingCount.Int32 + 1
	newRatingAvg, err := fromNumericToFloat64(globalStats.RatingAvg)
	if err != nil {
		return db.UpsertGlobalStatsParams{}, err
	}

	newRatingAvg = (newRatingAvg*float64(globalStats.RatingCount.Int32) + float64(rating.Total)) / float64(newRatingCount)

	return db.UpsertGlobalStatsParams{
		RatingAvg:   fromFloat64ToNumeric(newRatingAvg),
		RatingCount: fromInt32ToInt4(newRatingCount),
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

package mapper

import (
	"github.com/hyperremix/song-contest-rater-service/db"
	pb "github.com/hyperremix/song-contest-rater-service/protos/songcontestrater"
	"github.com/jackc/pgx/v5/pgtype"
)

func FromDbRatingListToResponse(r []db.Rating) (*pb.ListRatingsResponse, error) {
	var ratings []*pb.RatingResponse

	for _, rating := range r {
		proto, err := FromDbRatingToResponse(rating)
		if err != nil {
			return nil, err
		}

		ratings = append(ratings, proto)
	}

	return &pb.ListRatingsResponse{Ratings: ratings}, nil
}

func FromDbRatingToResponse(r db.Rating) (*pb.RatingResponse, error) {
	id, err := fromDbToProtoId(r.ID)
	if err != nil {
		return nil, err
	}

	competitionId, err := fromDbToProtoId(r.CompetitionID)
	if err != nil {
		return nil, err
	}

	actId, err := fromDbToProtoId(r.ActID)
	if err != nil {
		return nil, err
	}

	userId, err := fromDbToProtoId(r.UserID)
	if err != nil {
		return nil, err
	}

	return &pb.RatingResponse{
		Id:            id,
		CompetitionId: competitionId,
		ActId:         actId,
		UserId:        userId,
		Song:          r.Song.Int32,
		Singing:       r.Singing.Int32,
		Show:          r.Show.Int32,
		Looks:         r.Looks.Int32,
		Clothes:       r.Clothes.Int32,
		CreatedAt:     mapFromDbToProtoTimestamp(r.CreatedAt),
		UpdatedAt:     mapFromDbToProtoTimestamp(r.UpdatedAt),
	}, nil
}

func FromCreateRequestToInsertRating(c *pb.CreateRatingRequest) (db.InsertRatingParams, error) {
	competitionId, err := FromProtoToDbId(c.CompetitionId)
	if err != nil {
		return db.InsertRatingParams{}, err
	}

	actId, err := FromProtoToDbId(c.ActId)
	if err != nil {
		return db.InsertRatingParams{}, err
	}

	userId, err := FromProtoToDbId(c.UserId)
	if err != nil {
		return db.InsertRatingParams{}, err
	}

	return db.InsertRatingParams{
		CompetitionID: competitionId,
		ActID:         actId,
		UserID:        userId,
		Song:          fromInt32ToInt4(c.Song),
		Singing:       fromInt32ToInt4(c.Singing),
		Show:          fromInt32ToInt4(c.Show),
		Looks:         fromInt32ToInt4(c.Looks),
		Clothes:       fromInt32ToInt4(c.Clothes),
	}, nil
}

func FromUpdateRequestToUpdateRating(c *pb.UpdateRatingRequest) (db.UpdateRatingParams, error) {
	id, err := FromProtoToDbId(c.Id)
	if err != nil {
		return db.UpdateRatingParams{}, err
	}

	return db.UpdateRatingParams{
		ID:      id,
		Song:    fromInt32ToInt4(c.Song),
		Singing: fromInt32ToInt4(c.Singing),
		Show:    fromInt32ToInt4(c.Show),
		Looks:   fromInt32ToInt4(c.Looks),
		Clothes: fromInt32ToInt4(c.Clothes),
	}, nil
}

func fromInt32ToInt4(i int32) pgtype.Int4 {
	return pgtype.Int4{
		Int32: i,
		Valid: true,
	}
}

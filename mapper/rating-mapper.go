package mapper

import (
	pb "buf.build/gen/go/hyperremix/song-contest-rater-protos/protocolbuffers/go/songcontestrater/v5"
	"github.com/hyperremix/song-contest-rater-service/db"
)

func FromDbRatingListToResponse(r []db.Rating, u []db.User) ([]*pb.Rating, error) {
	var ratings []*pb.Rating

	for _, rating := range r {
		user := getUser(u, rating.UserID)

		proto, err := FromDbRatingToResponse(rating, user)
		if err != nil {
			return nil, NewResponseBindingError(err)
		}

		ratings = append(ratings, proto)
	}

	return ratings, nil
}

func FromDbRatingToResponse(r db.Rating, u *db.User) (*pb.Rating, error) {
	id, err := FromDbToProtoId(r.ID)
	if err != nil {
		return nil, NewResponseBindingError(err)
	}

	contestId, err := FromDbToProtoId(r.ContestID)
	if err != nil {
		return nil, NewResponseBindingError(err)
	}

	actId, err := FromDbToProtoId(r.ActID)
	if err != nil {
		return nil, NewResponseBindingError(err)
	}

	var userResponse *pb.User
	if u != nil {
		userResponse, err = FromDbUserToResponse(*u)
		if err != nil {
			return nil, NewResponseBindingError(err)
		}
	}

	return &pb.Rating{
		Id:        id,
		ContestId: contestId,
		ActId:     actId,
		Song:      r.Song.Int32,
		Singing:   r.Singing.Int32,
		Show:      r.Show.Int32,
		Looks:     r.Looks.Int32,
		Clothes:   r.Clothes.Int32,
		Total:     r.Total.Int32,
		User:      userResponse,
		CreatedAt: fromDbToProtoTimestamp(r.CreatedAt),
		UpdatedAt: fromDbToProtoTimestamp(r.UpdatedAt),
	}, nil
}

func FromCreateRequestToInsertRating(c *pb.CreateRatingRequest, protoUserId string) (db.InsertRatingParams, error) {
	contestId, err := FromProtoToDbId(c.ContestId)
	if err != nil {
		return db.InsertRatingParams{}, NewRequestBindingError(err)
	}

	actId, err := FromProtoToDbId(c.ActId)
	if err != nil {
		return db.InsertRatingParams{}, NewRequestBindingError(err)
	}

	userId, err := FromProtoToDbId(protoUserId)
	if err != nil {
		return db.InsertRatingParams{}, NewRequestBindingError(err)
	}

	return db.InsertRatingParams{
		ContestID: contestId,
		ActID:     actId,
		UserID:    userId,
		Song:      fromInt32ToInt4(c.Song),
		Singing:   fromInt32ToInt4(c.Singing),
		Show:      fromInt32ToInt4(c.Show),
		Looks:     fromInt32ToInt4(c.Looks),
		Clothes:   fromInt32ToInt4(c.Clothes),
	}, nil
}

func FromUpdateRequestToUpdateRating(c *pb.UpdateRatingRequest) (db.UpdateRatingParams, error) {
	id, err := FromProtoToDbId(c.Id)
	if err != nil {
		return db.UpdateRatingParams{}, NewRequestBindingError(err)
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

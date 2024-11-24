package mapper

import (
	"sort"

	"github.com/hyperremix/song-contest-rater-service/db"
	pb "github.com/hyperremix/song-contest-rater-service/protos/songcontestrater"
	"github.com/hyperremix/song-contest-rater-service/util"
	"github.com/jackc/pgx/v5/pgtype"
)

func FromDbActListToResponse(a []db.Act, r []db.Rating, u []db.User) (*pb.ListActsResponse, error) {
	var acts []*pb.ActResponse

	for _, act := range a {
		proto, err := FromDbActToResponse(act, getActRatings(r, act.ID), u)
		if err != nil {
			return nil, err
		}

		acts = append(acts, proto)
	}

	sort.Slice(acts, func(i, j int) bool {
		return util.ManyRatingsSum(acts[i].Ratings) > util.ManyRatingsSum(acts[j].Ratings)
	})

	return &pb.ListActsResponse{Acts: acts}, nil
}

func getActRatings(r []db.Rating, actID pgtype.UUID) []db.Rating {
	var ratings []db.Rating
	for _, rating := range r {
		if rating.ActID == actID {
			ratings = append(ratings, rating)
		}
	}

	return ratings
}

func FromDbActToResponse(a db.Act, r []db.Rating, u []db.User) (*pb.ActResponse, error) {
	id, err := FromDbToProtoId(a.ID)
	if err != nil {
		return nil, err
	}

	ratingListResponse, err := FromDbRatingListToResponse(r, u)
	if err != nil {
		return nil, err
	}

	sort.Slice(ratingListResponse.Ratings, func(i, j int) bool {
		return util.RatingSum(ratingListResponse.Ratings[i]) > util.RatingSum(ratingListResponse.Ratings[j])
	})

	return &pb.ActResponse{
		Id:         id,
		ArtistName: a.ArtistName,
		SongName:   a.SongName,
		ImageUrl:   a.ImageUrl,
		Ratings:    ratingListResponse.Ratings,
		CreatedAt:  fromDbToProtoTimestamp(a.CreatedAt),
		UpdatedAt:  fromDbToProtoTimestamp(a.UpdatedAt),
	}, nil
}

func FromCreateRequestToInsertAct(c *pb.CreateActRequest) (db.InsertActParams, error) {
	return db.InsertActParams{
		ArtistName: c.ArtistName,
		SongName:   c.SongName,
		ImageUrl:   c.ImageUrl,
	}, nil
}

func FromUpdateRequestToUpdateAct(c *pb.UpdateActRequest) (db.UpdateActParams, error) {
	id, err := FromProtoToDbId(c.Id)
	if err != nil {
		return db.UpdateActParams{}, err
	}

	return db.UpdateActParams{
		ID:         id,
		ArtistName: c.ArtistName,
		SongName:   c.SongName,
		ImageUrl:   c.ImageUrl,
	}, nil
}

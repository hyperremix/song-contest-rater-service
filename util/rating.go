package util

import pb "github.com/hyperremix/song-contest-rater-service/protos/songcontestrater"

func RatingSum(rating *pb.RatingResponse) int32 {
	return rating.Song + rating.Singing + rating.Show + rating.Looks + rating.Clothes
}

func ManyRatingsSum(ratings []*pb.RatingResponse) int32 {
	var sum int32
	for _, rating := range ratings {
		sum += RatingSum(rating)
	}

	return sum
}

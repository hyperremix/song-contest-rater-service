package util

import pb "buf.build/gen/go/hyperremix/song-contest-rater-protos/protocolbuffers/go/songcontestrater/v5"

func RatingSum(rating *pb.Rating) int32 {
	return rating.Song + rating.Singing + rating.Show + rating.Looks + rating.Clothes
}

func ManyRatingsSum(ratings []*pb.Rating) int32 {
	var sum int32
	for _, rating := range ratings {
		sum += RatingSum(rating)
	}

	return sum
}

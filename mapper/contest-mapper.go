package mapper

import (
	pb "github.com/hyperremix/song-contest-rater-protos/v4"
	"github.com/hyperremix/song-contest-rater-service/db"
)

func FromDbContestListToResponse(c []db.Contest) (*pb.ListContestsResponse, error) {
	var contests []*pb.ContestResponse

	for _, contest := range c {
		proto, err := FromDbContestToResponse(contest)
		if err != nil {
			return nil, NewResponseBindingError(err)
		}

		contests = append(contests, proto)
	}

	return &pb.ListContestsResponse{Contests: contests}, nil
}

func FromDbContestToResponse(c db.Contest) (*pb.ContestResponse, error) {
	id, err := FromDbToProtoId(c.ID)
	if err != nil {
		return nil, NewResponseBindingError(err)
	}

	return &pb.ContestResponse{
		Id:        id,
		City:      c.City,
		Country:   c.Country,
		Heat:      fromDbHeatToResponse(c.Heat),
		StartTime: fromDbToProtoTimestamp(c.StartTime),
		ImageUrl:  c.ImageUrl,
		CreatedAt: fromDbToProtoTimestamp(c.CreatedAt),
		UpdatedAt: fromDbToProtoTimestamp(c.UpdatedAt),
	}, nil
}

func FromDbToContestWithActsAndUsersResponse(c db.Contest, ratings []db.Rating, contestActs []db.ListActsByContestIdRow, users []db.User) (*pb.ContestResponse, error) {
	contest, err := FromDbContestToResponse(c)
	if err != nil {
		return nil, NewResponseBindingError(err)
	}

	actListResponse, err := FromDbOrderedActListToResponse(contestActs, ratings, users)
	if err != nil {
		return nil, NewResponseBindingError(err)
	}

	return &pb.ContestResponse{
		Id:        contest.Id,
		City:      contest.City,
		Country:   contest.Country,
		Heat:      contest.Heat,
		StartTime: contest.StartTime,
		ImageUrl:  contest.ImageUrl,
		CreatedAt: contest.CreatedAt,
		UpdatedAt: contest.UpdatedAt,
		Acts:      actListResponse.Acts,
	}, nil
}

func FromCreateRequestToInsertContest(r *pb.CreateContestRequest) db.InsertContestParams {
	return db.InsertContestParams{
		City:      r.City,
		Country:   r.Country,
		Heat:      fromRequestHeatToDb(r.Heat),
		StartTime: fromProtoToDbTimestamp(r.StartTime),
		ImageUrl:  r.ImageUrl,
	}
}

func FromUpdateRequestToUpdateContest(r *pb.UpdateContestRequest) (db.UpdateContestParams, error) {
	id, err := FromProtoToDbId(r.Id)
	if err != nil {
		return db.UpdateContestParams{}, NewRequestBindingError(err)
	}

	return db.UpdateContestParams{
		ID:        id,
		City:      r.City,
		Country:   r.Country,
		Heat:      fromRequestHeatToDb(r.Heat),
		StartTime: fromProtoToDbTimestamp(r.StartTime),
		ImageUrl:  r.ImageUrl,
	}, nil
}

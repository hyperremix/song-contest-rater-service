package mapper

import (
	pb "buf.build/gen/go/hyperremix/song-contest-rater-protos/protocolbuffers/go/songcontestrater/v5"
	"github.com/hyperremix/song-contest-rater-service/db"
)

func FromDbContestListToResponse(c []db.Contest) ([]*pb.Contest, error) {
	var contests []*pb.Contest

	for _, contest := range c {
		proto, err := FromDbContestToResponse(contest)
		if err != nil {
			return nil, NewResponseBindingError(err)
		}

		contests = append(contests, proto)
	}

	return contests, nil
}

func FromDbContestToResponse(c db.Contest) (*pb.Contest, error) {
	id, err := FromDbToProtoId(c.ID)
	if err != nil {
		return nil, NewResponseBindingError(err)
	}

	return &pb.Contest{
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

func FromDbToContestWithActsAndUsersResponse(c db.Contest, ratings []db.Rating, contestActs []db.ListActsByContestIdRow, users []db.User) (*pb.Contest, error) {
	contest, err := FromDbContestToResponse(c)
	if err != nil {
		return nil, NewResponseBindingError(err)
	}

	acts, err := FromDbOrderedActListToResponse(contestActs, ratings, users)
	if err != nil {
		return nil, NewResponseBindingError(err)
	}

	return &pb.Contest{
		Id:        contest.Id,
		City:      contest.City,
		Country:   contest.Country,
		Heat:      contest.Heat,
		StartTime: contest.StartTime,
		ImageUrl:  contest.ImageUrl,
		CreatedAt: contest.CreatedAt,
		UpdatedAt: contest.UpdatedAt,
		Acts:      acts,
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

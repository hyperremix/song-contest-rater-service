package mapper

import (
	"github.com/hyperremix/song-contest-rater-service/db"
	pb "github.com/hyperremix/song-contest-rater-service/protos/songcontestrater"
)

func FromDbCompetitionListToResponse(c []db.Competition) (*pb.ListCompetitionsResponse, error) {
	var competitions []*pb.CompetitionResponse

	for _, competition := range c {
		proto, err := FromDbCompetitionToResponse(competition)
		if err != nil {
			return nil, err
		}

		competitions = append(competitions, proto)
	}

	return &pb.ListCompetitionsResponse{Competitions: competitions}, nil
}

func FromDbCompetitionToResponse(c db.Competition) (*pb.CompetitionResponse, error) {
	id, err := FromDbToProtoId(c.ID)
	if err != nil {
		return nil, err
	}

	return &pb.CompetitionResponse{
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

func FromDbToCompetitionWithActsAndUsersResponse(c db.Competition, ratings []db.Rating, competitionActs []db.ListActsByCompetitionIdRow, users []db.User) (*pb.CompetitionResponse, error) {
	competition, err := FromDbCompetitionToResponse(c)
	if err != nil {
		return nil, err
	}

	actListResponse, err := FromDbOrderedActListToResponse(competitionActs, ratings, users)
	if err != nil {
		return nil, err
	}

	return &pb.CompetitionResponse{
		Id:        competition.Id,
		City:      competition.City,
		Country:   competition.Country,
		Heat:      competition.Heat,
		StartTime: competition.StartTime,
		ImageUrl:  competition.ImageUrl,
		CreatedAt: competition.CreatedAt,
		UpdatedAt: competition.UpdatedAt,
		Acts:      actListResponse.Acts,
	}, nil
}

func FromCreateRequestToInsertCompetition(r *pb.CreateCompetitionRequest) db.InsertCompetitionParams {
	return db.InsertCompetitionParams{
		City:      r.City,
		Country:   r.Country,
		Heat:      fromRequestHeatToDb(r.Heat),
		StartTime: fromProtoToDbTimestamp(r.StartTime),
		ImageUrl:  r.ImageUrl,
	}
}

func FromUpdateRequestToUpdateCompetition(r *pb.UpdateCompetitionRequest) (db.UpdateCompetitionParams, error) {
	id, err := FromProtoToDbId(r.Id)
	if err != nil {
		return db.UpdateCompetitionParams{}, err
	}

	return db.UpdateCompetitionParams{
		ID:        id,
		City:      r.City,
		Country:   r.Country,
		Heat:      fromRequestHeatToDb(r.Heat),
		StartTime: fromProtoToDbTimestamp(r.StartTime),
		ImageUrl:  r.ImageUrl,
	}, nil
}

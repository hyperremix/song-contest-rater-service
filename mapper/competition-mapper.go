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
	id, err := fromDbToProtoId(c.ID)
	if err != nil {
		return nil, err
	}

	return &pb.CompetitionResponse{
		Id:          id,
		City:        c.City,
		Country:     c.Country,
		Description: c.Description,
		StartTime:   mapFromDbToProtoTimestamp(c.StartTime),
		ImageUrl:    c.ImageUrl,
		CreatedAt:   mapFromDbToProtoTimestamp(c.CreatedAt),
		UpdatedAt:   mapFromDbToProtoTimestamp(c.UpdatedAt),
	}, nil
}

func FromCreateRequestToInsertCompetition(r *pb.CreateCompetitionRequest) db.InsertCompetitionParams {
	return db.InsertCompetitionParams{
		City:        r.City,
		Country:     r.Country,
		Description: r.Description,
		StartTime:   mapFromProtoToDbTimestamp(r.StartTime),
		ImageUrl:    r.ImageUrl,
	}
}

func FromUpdateRequestToUpdateCompetition(r *pb.UpdateCompetitionRequest) (db.UpdateCompetitionParams, error) {
	id, err := FromProtoToDbId(r.Id)
	if err != nil {
		return db.UpdateCompetitionParams{}, err
	}

	return db.UpdateCompetitionParams{
		ID:          id,
		City:        r.City,
		Country:     r.Country,
		Description: r.Description,
		StartTime:   mapFromProtoToDbTimestamp(r.StartTime),
		ImageUrl:    r.ImageUrl,
	}, nil
}

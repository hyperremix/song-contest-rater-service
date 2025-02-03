package mapper

import (
	"github.com/hyperremix/song-contest-rater-service/db"
	pb "github.com/hyperremix/song-contest-rater-service/protos/songcontestrater"
)

func FromCreateRequestToInsertCompetitionAct(request *pb.CreateParticipationRequest) (db.InsertCompetitionActParams, error) {
	competitionId, err := FromProtoToDbId(request.CompetitionId)
	if err != nil {
		return db.InsertCompetitionActParams{}, NewRequestBindingError(err)
	}

	actId, err := FromProtoToDbId(request.ActId)
	if err != nil {
		return db.InsertCompetitionActParams{}, NewRequestBindingError(err)
	}

	return db.InsertCompetitionActParams{CompetitionID: competitionId, ActID: actId, Order: fromInt32ToInt4(request.Order)}, nil
}

type DeleteParticipationRequest struct {
	CompetitionID string `query:"competition_id"`
	ActID         string `query:"act_id"`
}

func FromDeleteRequestToDeleteCompetitionAct(request *DeleteParticipationRequest) (db.DeleteCompetitionActParams, error) {
	competitionId, err := FromProtoToDbId(request.CompetitionID)
	if err != nil {
		return db.DeleteCompetitionActParams{}, NewRequestBindingError(err)
	}

	actId, err := FromProtoToDbId(request.ActID)
	if err != nil {
		return db.DeleteCompetitionActParams{}, NewRequestBindingError(err)
	}

	return db.DeleteCompetitionActParams{CompetitionID: competitionId, ActID: actId}, nil
}

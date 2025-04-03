package mapper

import (
	pb "github.com/hyperremix/song-contest-rater-protos/v3"
	"github.com/hyperremix/song-contest-rater-service/db"
)

func FromManyCompetitionActsToProto(rows []db.CompetitionsAct) (*pb.ListParticipationsResponse, error) {
	protoResponses := make([]*pb.ParticipationResponse, len(rows))

	for i, row := range rows {
		protoResponse, err := FromCompetitionActToProto(&row)
		if err != nil {
			return nil, NewRequestBindingError(err)
		}
		protoResponses[i] = protoResponse
	}

	return &pb.ListParticipationsResponse{Participations: protoResponses}, nil
}

func FromCompetitionActToProto(row *db.CompetitionsAct) (*pb.ParticipationResponse, error) {
	competitionId, err := FromDbToProtoId(row.CompetitionID)
	if err != nil {
		return nil, NewRequestBindingError(err)
	}

	actId, err := FromDbToProtoId(row.ActID)
	if err != nil {
		return nil, NewRequestBindingError(err)
	}

	return &pb.ParticipationResponse{
		CompetitionId: competitionId,
		ActId:         actId,
		Order:         int32(row.Order.Int32),
	}, nil
}

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

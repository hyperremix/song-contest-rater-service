package mapper

import (
	pb "buf.build/gen/go/hyperremix/song-contest-rater-protos/protocolbuffers/go/songcontestrater/v5"
	"github.com/hyperremix/song-contest-rater-service/db"
)

func FromManyParticipationsToProto(rows []db.Participation) ([]*pb.Participation, error) {
	protoResponses := make([]*pb.Participation, len(rows))

	for i, row := range rows {
		protoResponse, err := FromParticipationToProto(&row)
		if err != nil {
			return nil, NewRequestBindingError(err)
		}
		protoResponses[i] = protoResponse
	}

	return protoResponses, nil
}

func FromParticipationToProto(row *db.Participation) (*pb.Participation, error) {
	contestId, err := FromDbToProtoId(row.ContestID)
	if err != nil {
		return nil, NewRequestBindingError(err)
	}

	actId, err := FromDbToProtoId(row.ActID)
	if err != nil {
		return nil, NewRequestBindingError(err)
	}

	return &pb.Participation{
		ContestId: contestId,
		ActId:     actId,
		Order:     int32(row.Order.Int32),
	}, nil
}

func FromCreateRequestToInsertParticipation(request *pb.CreateParticipationRequest) (db.InsertParticipationParams, error) {
	contestId, err := FromProtoToDbId(request.ContestId)
	if err != nil {
		return db.InsertParticipationParams{}, NewRequestBindingError(err)
	}

	actId, err := FromProtoToDbId(request.ActId)
	if err != nil {
		return db.InsertParticipationParams{}, NewRequestBindingError(err)
	}

	return db.InsertParticipationParams{ContestID: contestId, ActID: actId, Order: fromInt32ToInt4(request.Order)}, nil
}

func FromDeleteRequestToDeleteParticipation(request *pb.DeleteParticipationRequest) (db.DeleteParticipationParams, error) {
	contestId, err := FromProtoToDbId(request.ContestId)
	if err != nil {
		return db.DeleteParticipationParams{}, NewRequestBindingError(err)
	}

	actId, err := FromProtoToDbId(request.ActId)
	if err != nil {
		return db.DeleteParticipationParams{}, NewRequestBindingError(err)
	}

	return db.DeleteParticipationParams{ContestID: contestId, ActID: actId}, nil
}

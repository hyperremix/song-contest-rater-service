package mapper

import (
	"net/http"

	"github.com/hyperremix/song-contest-rater-service/db"
	pb "github.com/hyperremix/song-contest-rater-service/protos/songcontestrater"
	"github.com/labstack/echo/v4"
)

func FromCreateRequestToInsertCompetitionAct(request *pb.CreateParticipationRequest) (db.InsertCompetitionActParams, error) {
	competitionId, err := FromProtoToDbId(request.CompetitionId)
	if err != nil {
		return db.InsertCompetitionActParams{}, echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
	}

	actId, err := FromProtoToDbId(request.ActId)
	if err != nil {
		return db.InsertCompetitionActParams{}, echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
	}

	return db.InsertCompetitionActParams{CompetitionID: competitionId, ActID: actId}, nil
}

type DeleteParticipationRequest struct {
	CompetitionID string `query:"competition_id"`
	ActID         string `query:"act_id"`
}

func FromDeleteRequestToDeleteCompetitionAct(request *DeleteParticipationRequest) (db.DeleteCompetitionActParams, error) {
	competitionId, err := FromProtoToDbId(request.CompetitionID)
	if err != nil {
		return db.DeleteCompetitionActParams{}, echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
	}

	actId, err := FromProtoToDbId(request.ActID)
	if err != nil {
		return db.DeleteCompetitionActParams{}, echo.NewHTTPError(http.StatusBadRequest, "could not bind request")
	}

	return db.DeleteCompetitionActParams{CompetitionID: competitionId, ActID: actId}, nil
}

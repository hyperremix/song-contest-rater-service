package mapper

import (
	"github.com/hyperremix/song-contest-rater-service/db"
	pb "github.com/hyperremix/song-contest-rater-service/protos/songcontestrater"
)

func FromDbActListToResponse(a []db.Act) (*pb.ListActsResponse, error) {
	var acts []*pb.ActResponse

	for _, act := range a {
		proto, err := FromDbActToResponse(act)
		if err != nil {
			return nil, err
		}

		acts = append(acts, proto)
	}

	return &pb.ListActsResponse{Acts: acts}, nil
}

func FromDbActToResponse(a db.Act) (*pb.ActResponse, error) {
	id, err := fromDbToProtoId(a.ID)
	if err != nil {
		return nil, err
	}

	return &pb.ActResponse{
		Id:         id,
		ArtistName: a.ArtistName,
		SongName:   a.SongName,
		ImageUrl:   a.ImageUrl,
		CreatedAt:  mapFromDbToProtoTimestamp(a.CreatedAt),
		UpdatedAt:  mapFromDbToProtoTimestamp(a.UpdatedAt),
	}, nil
}

func FromCreateRequestToInsertAct(c *pb.CreateActRequest) (db.InsertActParams, error) {
	return db.InsertActParams{
		ArtistName: c.ArtistName,
		SongName:   c.SongName,
		ImageUrl:   c.ImageUrl,
	}, nil
}

func FromUpdateRequestToUpdateAct(c *pb.UpdateActRequest) (db.UpdateActParams, error) {
	id, err := FromProtoToDbId(c.Id)
	if err != nil {
		return db.UpdateActParams{}, err
	}

	return db.UpdateActParams{
		ID:         id,
		ArtistName: c.ArtistName,
		SongName:   c.SongName,
		ImageUrl:   c.ImageUrl,
	}, nil
}

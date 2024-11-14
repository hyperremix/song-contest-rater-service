package mapper

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func fromDbToProtoId(id pgtype.UUID) (string, error) {
	uuid, err := uuid.FromBytes(id.Bytes[:])
	if err != nil {
		return "", err
	}

	return uuid.String(), nil
}

func FromProtoToDbId(id string) (pgtype.UUID, error) {
	uuid, err := uuid.Parse(id)
	if err != nil {
		return pgtype.UUID{}, err
	}

	return pgtype.UUID{Bytes: uuid, Valid: true}, nil
}

func mapFromProtoToDbTimestamp(timestamp *timestamppb.Timestamp) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: timestamp.AsTime(), Valid: true}
}

func mapFromDbToProtoTimestamp(timestamp pgtype.Timestamptz) *timestamppb.Timestamp {
	return timestamppb.New(timestamp.Time)
}

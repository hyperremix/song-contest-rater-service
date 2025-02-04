package mapper

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func FromDbToProtoId(id pgtype.UUID) (string, error) {
	uuid, err := uuid.FromBytes(id.Bytes[:])
	if err != nil {
		return "", NewResponseBindingError(err)
	}

	return uuid.String(), nil
}

func FromProtoToDbId(id string) (pgtype.UUID, error) {
	uuid, err := uuid.Parse(id)
	if err != nil {
		return pgtype.UUID{}, NewRequestBindingError(err)
	}

	return pgtype.UUID{Bytes: uuid, Valid: true}, nil
}

func fromProtoToDbTimestamp(timestamp *timestamppb.Timestamp) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: timestamp.AsTime(), Valid: true}
}

func fromDbToProtoTimestamp(timestamp pgtype.Timestamptz) *timestamppb.Timestamp {
	return timestamppb.New(timestamp.Time)
}

func fromInt32ToInt4(i int32) pgtype.Int4 {
	return pgtype.Int4{
		Int32: i,
		Valid: true,
	}
}

func fromNumericToFloat64(n pgtype.Numeric) (float64, error) {
	f, err := n.Float64Value()
	if err != nil {
		return 0, NewResponseBindingError(err)
	}

	return f.Float64, nil
}

func fromFloat64ToNumeric(f float64) pgtype.Numeric {
	var x pgtype.Numeric
	err := x.Scan(fmt.Sprintf("%f", f))

	if err != nil {
		return pgtype.Numeric{}
	}

	return x
}

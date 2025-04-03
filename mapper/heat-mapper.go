package mapper

import (
	pb "github.com/hyperremix/song-contest-rater-protos/v3"
	"github.com/hyperremix/song-contest-rater-service/db"
)

func fromDbHeatToResponse(h db.Heat) pb.Heat {
	switch h {
	case db.HeatHEATSEMIFINAL:
		return pb.Heat_HEAT_SEMI_FINAL
	case db.HeatHEATFINAL:
		return pb.Heat_HEAT_FINAL
	case db.HeatHEAT1:
		return pb.Heat_HEAT_1
	case db.HeatHEAT2:
		return pb.Heat_HEAT_2
	case db.HeatHEAT3:
		return pb.Heat_HEAT_3
	case db.HeatHEAT4:
		return pb.Heat_HEAT_4
	case db.HeatHEAT5:
		return pb.Heat_HEAT_5
	case db.HeatHEATFINALQUALIFIER:
		return pb.Heat_HEAT_FINAL_QUALIFIER
	default:
		return pb.Heat_HEAT_UNSPECIFIED
	}
}

func fromRequestHeatToDb(h pb.Heat) db.Heat {
	switch h {
	case pb.Heat_HEAT_SEMI_FINAL:
		return db.HeatHEATSEMIFINAL
	case pb.Heat_HEAT_FINAL:
		return db.HeatHEATFINAL
	case pb.Heat_HEAT_1:
		return db.HeatHEAT1
	case pb.Heat_HEAT_2:
		return db.HeatHEAT2
	case pb.Heat_HEAT_3:
		return db.HeatHEAT3
	case pb.Heat_HEAT_4:
		return db.HeatHEAT4
	case pb.Heat_HEAT_5:
		return db.HeatHEAT5
	case pb.Heat_HEAT_FINAL_QUALIFIER:
		return db.HeatHEATFINALQUALIFIER
	default:
		return db.HeatHEATUNSPECIFIED
	}
}

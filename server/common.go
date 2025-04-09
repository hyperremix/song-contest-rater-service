package server

import (
	"net/http"

	pb "buf.build/gen/go/hyperremix/song-contest-rater-protos/connectrpc/go/songcontestrater/v5/songcontestraterv5connect"
	"connectrpc.com/grpcreflect"
	"github.com/hyperremix/song-contest-rater-service/event"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

var broker = event.NewBroker()

func RegisterHandlers(e *echo.Echo, connPool *pgxpool.Pool) {
	e.Any(connectHandler(pb.NewContestServiceHandler(NewContestServer(connPool))))
	e.Any(connectHandler(pb.NewUserServiceHandler(NewUserServer(connPool))))
	e.Any(connectHandler(pb.NewStatServiceHandler(NewStatServer(connPool))))
	e.Any(connectHandler(pb.NewRatingServiceHandler(NewRatingServer(connPool))))
	e.Any(connectHandler(pb.NewParticipationServiceHandler(NewParticipationServer(connPool))))
	e.Any(connectHandler(pb.NewActServiceHandler(NewActServer(connPool))))

	reflector := grpcreflect.NewStaticReflector("songcontestrater.v5.ContestService", "songcontestrater.v5.UserService", "songcontestrater.v5.StatService", "songcontestrater.v5.RatingService", "songcontestrater.v5.ParticipationService", "songcontestrater.v5.ActService")
	e.Any(connectHandler(grpcreflect.NewHandlerV1(reflector)))
	e.Any(connectHandler(grpcreflect.NewHandlerV1Alpha(reflector)))
}

func connectHandler(path string, handler http.Handler) (string, echo.HandlerFunc) {
	path = path + "*"
	return path, echo.WrapHandler(handler)
}

package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	clerk "github.com/clerk/clerk-sdk-go/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"

	pb "github.com/hyperremix/song-contest-rater-protos/v4"
	"github.com/hyperremix/song-contest-rater-service/authz"
	scrlogging "github.com/hyperremix/song-contest-rater-service/logging"
	"github.com/hyperremix/song-contest-rater-service/server"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func setupPool(ctx context.Context) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(os.Getenv("SONGCONTESTRATERSERVICE_DB_CONNECTION_STRING"))
	if err != nil {
		return nil, err
	}

	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute

	config.ConnConfig.ConnectTimeout = 5 * time.Second
	config.ConnConfig.RuntimeParams = map[string]string{
		"statement_timeout":                   "30000",
		"lock_timeout":                        "10000",
		"idle_in_transaction_session_timeout": "300000",
		"search_path":                         "song_contest_rater_service",
	}

	return pgxpool.NewWithConfig(ctx, config)
}

func main() {
	godotenv.Load(".env")

	ctx := context.Background()

	clerk.SetKey(os.Getenv("SONGCONTESTRATERSERVICE_CLERK_SECRET_KEY"))

	connPool, err := setupPool(ctx)
	defer connPool.Close()

	logger := zerolog.New(os.Stdout)

	opts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}

	requestAuthorizer := authz.NewRequestAuthorizer(connPool)

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logging.UnaryServerInterceptor(scrlogging.InterceptorLogger(logger), opts...),
			requestAuthorizer.UnaryServerInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			logging.StreamServerInterceptor(scrlogging.InterceptorLogger(logger), opts...),
			requestAuthorizer.StreamServerInterceptor(),
		),
	)
	reflection.Register(grpcServer)

	pb.RegisterContestServer(grpcServer, server.NewContestServer(connPool))
	pb.RegisterActServer(grpcServer, server.NewActServer(connPool))
	pb.RegisterRatingServer(grpcServer, server.NewRatingServer(connPool))
	pb.RegisterUserServer(grpcServer, server.NewUserServer(connPool))
	pb.RegisterParticipationServer(grpcServer, server.NewParticipationServer(connPool))
	pb.RegisterStatServer(grpcServer, server.NewStatServer(connPool))

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		connPool.Close()
		grpcServer.GracefulStop()

		<-ctx.Done()
	}()

	listen, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal().Msgf("%v\n", err)
	}

	log.Info().Msg("starting gRPC server... Listening on :8080")
	err = grpcServer.Serve(listen)
	if err != nil {
		log.Fatal().Msgf("%v\n", err)
	}
}

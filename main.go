package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/hyperremix/song-contest-rater-service/environment"
	"github.com/hyperremix/song-contest-rater-service/protos/songcontestrater"
	"github.com/hyperremix/song-contest-rater-service/server"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	if err := RunServer(); err != nil {
		log.Error().Msgf("%v\n", err)
	}
}

func RunServer() error {
	ctx := context.Background()

	connPool, err := pgxpool.New(ctx, environment.DB_CONNECTION_STRING)
	if err != nil {
		return err
	}
	defer connPool.Close()

	logger := zerolog.New(os.Stdout)

	opts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logging.UnaryServerInterceptor(InterceptorLogger(logger), opts...),
		),
		grpc.ChainStreamInterceptor(
			logging.StreamServerInterceptor(InterceptorLogger(logger), opts...),
		),
	)
	reflection.Register(grpcServer)

	songcontestrater.RegisterCompetitionServer(grpcServer, server.NewCompetitionServer(connPool))

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			log.Info().Msg("shutting down gRPC server...")

			grpcServer.GracefulStop()

			<-ctx.Done()
		}
	}()

	listen, err := net.Listen("tcp", ":8080")
	if err != nil {
		return err
	}

	log.Info().Msg("starting gRPC server... Listening on :8080")
	return grpcServer.Serve(listen)
}

func InterceptorLogger(l zerolog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l := l.With().Fields(fields).Logger()

		switch lvl {
		case logging.LevelDebug:
			l.Debug().Msg(msg)
		case logging.LevelInfo:
			l.Info().Msg(msg)
		case logging.LevelWarn:
			l.Warn().Msg(msg)
		case logging.LevelError:
			l.Error().Msg(msg)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}

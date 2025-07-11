package server

import (
	"context"

	pb "github.com/hyperremix/song-contest-rater-protos/v4"
	"github.com/hyperremix/song-contest-rater-service/authz"
	"github.com/hyperremix/song-contest-rater-service/db"
	"github.com/hyperremix/song-contest-rater-service/mapper"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ContestServer struct {
	pb.UnimplementedContestServer
	queries  *db.Queries
	connPool *pgxpool.Pool
}

func NewContestServer(connPool *pgxpool.Pool) *ContestServer {
	return &ContestServer{
		queries:  db.New(connPool),
		connPool: connPool,
	}
}

func (s *ContestServer) ListContests(ctx context.Context, request *emptypb.Empty) (*pb.ListContestsResponse, error) {
	contests, err := s.queries.ListContests(ctx)
	if err != nil {
		return nil, err
	}

	response, err := mapper.FromDbContestListToResponse(contests)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *ContestServer) GetContest(ctx context.Context, request *pb.GetContestRequest) (*pb.ContestResponse, error) {
	id, err := mapper.FromProtoToDbId(request.Id)
	if err != nil {
		return nil, err
	}

	contest, err := s.queries.GetContestById(ctx, id)
	if err != nil {
		return nil, err
	}

	ratings, err := s.queries.ListRatingsByContestId(ctx, contest.ID)
	if err != nil {
		return nil, err
	}

	acts, err := s.queries.ListActsByContestId(ctx, contest.ID)
	if err != nil {
		return nil, err
	}

	users, err := s.queries.ListUsersByContestId(ctx, contest.ID)
	if err != nil {
		return nil, err
	}

	response, err := mapper.FromDbToContestWithActsAndUsersResponse(contest, ratings, acts, users)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *ContestServer) CreateContest(ctx context.Context, request *pb.CreateContestRequest) (*pb.ContestResponse, error) {
	authUser := ctx.Value(authz.AuthUserContextKey{}).(*authz.AuthUser)
	if err := authUser.CheckIsAdmin(); err != nil {
		return nil, err
	}

	insertParams := mapper.FromCreateRequestToInsertContest(request)

	contest, err := s.queries.InsertContest(ctx, insertParams)
	if err != nil {
		return nil, err
	}

	response, err := mapper.FromDbContestToResponse(contest)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *ContestServer) UpdateContest(ctx context.Context, request *pb.UpdateContestRequest) (*pb.ContestResponse, error) {
	authUser := ctx.Value(authz.AuthUserContextKey{}).(*authz.AuthUser)
	if err := authUser.CheckIsAdmin(); err != nil {
		return nil, err
	}

	updateParams, err := mapper.FromUpdateRequestToUpdateContest(request)
	if err != nil {
		return nil, err
	}

	contest, err := s.queries.UpdateContest(ctx, updateParams)
	if err != nil {
		return nil, err
	}

	response, err := mapper.FromDbContestToResponse(contest)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *ContestServer) DeleteContest(ctx context.Context, request *pb.DeleteContestRequest) (*pb.ContestResponse, error) {
	authUser := ctx.Value(authz.AuthUserContextKey{}).(*authz.AuthUser)
	if err := authUser.CheckIsAdmin(); err != nil {
		return nil, err
	}

	id, err := mapper.FromProtoToDbId(request.Id)
	if err != nil {
		return nil, err
	}

	contest, err := s.queries.DeleteContestById(ctx, id)
	if err != nil {
		return nil, err
	}

	response, err := mapper.FromDbContestToResponse(contest)
	if err != nil {
		return nil, err
	}

	return response, nil
}

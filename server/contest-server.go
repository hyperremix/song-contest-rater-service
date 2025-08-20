package server

import (
	"context"

	pb "buf.build/gen/go/hyperremix/song-contest-rater-protos/protocolbuffers/go/songcontestrater/v5"
	"connectrpc.com/connect"
	"github.com/hyperremix/song-contest-rater-service/authz"
	"github.com/hyperremix/song-contest-rater-service/db"
	"github.com/hyperremix/song-contest-rater-service/mapper"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ContestServer struct {
	queries  *db.Queries
	connPool *pgxpool.Pool
}

func NewContestServer(connPool *pgxpool.Pool) *ContestServer {
	return &ContestServer{
		queries:  db.New(connPool),
		connPool: connPool,
	}
}

func (s *ContestServer) ListContests(ctx context.Context, request *connect.Request[pb.ListContestsRequest]) (*connect.Response[pb.ListContestsResponse], error) {
	contests, err := s.queries.ListContests(ctx)
	if err != nil {
		return nil, err
	}

	response, err := mapper.FromDbContestListToResponse(contests)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&pb.ListContestsResponse{Contests: response}), nil
}

func (s *ContestServer) GetContest(ctx context.Context, request *connect.Request[pb.GetContestRequest]) (*connect.Response[pb.GetContestResponse], error) {
	id, err := mapper.FromProtoToDbId(request.Msg.Id)
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

	return connect.NewResponse(&pb.GetContestResponse{Contest: response}), nil
}

func (s *ContestServer) CreateContest(ctx context.Context, request *connect.Request[pb.CreateContestRequest]) (*connect.Response[pb.CreateContestResponse], error) {
	authUser := ctx.Value(authz.AuthUserContextKey).(*authz.AuthUser)
	if err := authUser.CheckIsAdmin(); err != nil {
		return nil, err
	}

	insertParams := mapper.FromCreateRequestToInsertContest(request.Msg)

	contest, err := s.queries.InsertContest(ctx, insertParams)
	if err != nil {
		return nil, err
	}

	response, err := mapper.FromDbContestToResponse(contest)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&pb.CreateContestResponse{Contest: response}), nil
}

func (s *ContestServer) UpdateContest(ctx context.Context, request *connect.Request[pb.UpdateContestRequest]) (*connect.Response[pb.UpdateContestResponse], error) {
	authUser := ctx.Value(authz.AuthUserContextKey).(*authz.AuthUser)
	if err := authUser.CheckIsAdmin(); err != nil {
		return nil, err
	}

	updateParams, err := mapper.FromUpdateRequestToUpdateContest(request.Msg)
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

	return connect.NewResponse(&pb.UpdateContestResponse{Contest: response}), nil
}

func (s *ContestServer) DeleteContest(ctx context.Context, request *connect.Request[pb.DeleteContestRequest]) (*connect.Response[pb.DeleteContestResponse], error) {
	authUser := ctx.Value(authz.AuthUserContextKey).(*authz.AuthUser)
	if err := authUser.CheckIsAdmin(); err != nil {
		return nil, err
	}

	id, err := mapper.FromProtoToDbId(request.Msg.Id)
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

	return connect.NewResponse(&pb.DeleteContestResponse{Contest: response}), nil
}

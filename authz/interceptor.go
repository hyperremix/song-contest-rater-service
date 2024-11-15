package authz

import (
	"context"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/hyperremix/song-contest-rater-service/db"
	"github.com/hyperremix/song-contest-rater-service/mapper"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type CustomClaims struct {
	Scope string `json:"scope"`
}

type AuthUserContextKey struct{}

func (c CustomClaims) Validate(ctx context.Context) error {
	return nil
}

func UnaryServerInterceptor(connPool *pgxpool.Pool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "missing metadata")
		}

		authorization, ok := md["authorization"]
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "missing token")
		}

		authUser, err := validateAuthorization(ctx, authorization[0])
		if err != nil {
			return nil, err
		}

		conn, err := connPool.Acquire(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "could not acquire connection: %v", err)
		}
		defer conn.Release()

		queries := db.New(conn)

		user, err := queries.GetUserBySub(ctx, authUser.Sub)
		if err != nil {
			ctx = context.WithValue(ctx, AuthUserContextKey{}, authUser)
		} else {
			userID, err := mapper.FromDbToProtoId(user.ID)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "could not map user id: %v", err)
			}

			authUser.UserID = userID
			ctx = context.WithValue(ctx, AuthUserContextKey{}, authUser)
		}

		return handler(ctx, req)
	}
}

func validateAuthorization(ctx context.Context, authHeader string) (*AuthUser, error) {
	issuerURL, err := url.Parse(os.Getenv("SONGCONTESTRATERSERVICE_AUTH0_ISSUER_URL"))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not parse the issuer url: %v", err)
	}

	provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)

	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		[]string{os.Getenv("SONGCONTESTRATERSERVICE_AUTH0_AUDIENCE")},
		validator.WithCustomClaims(
			func() validator.CustomClaims {
				return &CustomClaims{}
			},
		),
		validator.WithAllowedClockSkew(time.Minute),
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not set up the jwt validator: %v", err)
	}

	tokenBase64 := strings.Split(authHeader, " ")[1]

	if _, err := jwtValidator.ValidateToken(ctx, tokenBase64); err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	var authUser AuthUser
	err = authUser.decode(tokenBase64)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	return &authUser, nil
}

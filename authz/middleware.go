package authz

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/clerk/clerk-sdk-go/v2/user"
	clerkuser "github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/hyperremix/song-contest-rater-service/db"
	"github.com/hyperremix/song-contest-rater-service/mapper"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

type RequestAuthorizer struct {
	connPool *pgxpool.Pool
	queries  *db.Queries
}

func NewRequestAuthorizer(connPool *pgxpool.Pool) *RequestAuthorizer {
	return &RequestAuthorizer{
		connPool: connPool,
		queries:  db.New(connPool),
	}
}

type CustomClaims struct {
	Scope string `json:"scope"`
}

var AuthUserContextKey = "authUser"

func (c CustomClaims) Validate(ctx context.Context) error {
	return nil
}

func (r *RequestAuthorizer) Authorize() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(echoCtx echo.Context) error {
			if echoCtx.Request().URL.Path == "/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo" {
				return next(echoCtx)
			}

			ctx := echoCtx.Request().Context()
			authorization, ok := echoCtx.Request().Header["Authorization"]
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			authUser, err := validateAuthorization(ctx, authorization[0])
			if err != nil {
				return err
			}

			userID, dbUser, err := syncDbAndClerkState(ctx, authUser, r)
			if err != nil {
				return err
			}

			authUser.UserID = userID
			authUser.DbUser = dbUser
			echoCtx.Set(AuthUserContextKey, authUser)

			return next(echoCtx)
		}
	}
}

func validateAuthorization(ctx context.Context, authHeader string) (*AuthUser, error) {
	sessionToken := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := jwt.Verify(ctx, &jwt.VerifyParams{
		Token: sessionToken,
	})
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "could not verify token")
	}

	user, err := user.Get(ctx, claims.Subject)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "could not get user from token")
	}

	var publicMetadata PublicMetadata
	if err := json.Unmarshal(user.PublicMetadata, &publicMetadata); err != nil {
		return nil, echo.NewHTTPError(http.StatusForbidden, "missing permission to access this resource")
	}

	return &AuthUser{
		ClerkUser: user,
		Metadata:  publicMetadata,
	}, nil
}

func syncDbAndClerkState(ctx context.Context, authUser *AuthUser, r *RequestAuthorizer) (string, db.User, error) {
	log := zerolog.Ctx(ctx)
	user, err := r.queries.GetUserBySub(ctx, authUser.ClerkUser.ID)
	if err != nil {
		log.Info().Msgf("user not found in db, inserting user: %s", authUser.ClerkUser.ID)
		user, err = r.queries.InsertUser(ctx, db.InsertUserParams{
			Sub:       authUser.ClerkUser.ID,
			Email:     authUser.ClerkUser.EmailAddresses[0].EmailAddress,
			Firstname: *authUser.ClerkUser.FirstName,
			Lastname:  *authUser.ClerkUser.LastName,
			ImageUrl:  *authUser.ClerkUser.ImageURL,
		})
		if err != nil {
			return "", db.User{}, err
		}
	}

	userID, err := mapper.FromDbToProtoId(user.ID)
	if err != nil {
		return "", db.User{}, err
	}

	if authUser.Metadata.ID == "" {
		log.Info().Msgf("metadata.id is not set, updating metadata: %s", authUser.ClerkUser.ID)
		metadataJSON, err := json.Marshal(map[string]interface{}{
			"id": userID,
		})
		if err != nil {
			return "", db.User{}, err
		}

		rawJSON := json.RawMessage(metadataJSON)
		_, err = clerkuser.UpdateMetadata(ctx, authUser.ClerkUser.ID, &clerkuser.UpdateMetadataParams{
			PublicMetadata: &rawJSON,
		})
		if err != nil {
			return "", db.User{}, err
		}

		authUser.Metadata.ID = userID
	}

	if user.Firstname != *authUser.ClerkUser.FirstName || user.Lastname != *authUser.ClerkUser.LastName || user.ImageUrl != *authUser.ClerkUser.ImageURL {
		log.Info().Msgf("user data has changed, updating user: %s", authUser.ClerkUser.ID)
		updateParams, err := mapper.ToUpdateUserParams(user.ID, *authUser.ClerkUser.FirstName, *authUser.ClerkUser.LastName, *authUser.ClerkUser.ImageURL)
		if err != nil {
			return "", db.User{}, err
		}

		_, err = r.queries.UpdateUser(ctx, updateParams)
		if err != nil {
			return "", db.User{}, err
		}
	}

	return userID, user, nil
}

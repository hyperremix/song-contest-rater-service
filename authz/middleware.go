package authz

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/hyperremix/song-contest-rater-service/db"
	"github.com/hyperremix/song-contest-rater-service/mapper"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

type CustomClaims struct {
	Scope string `json:"scope"`
}

var AuthUserContextKey = "authUser"

func (c CustomClaims) Validate(ctx context.Context) error {
	return nil
}

func RequestAuthorizer(connPool *pgxpool.Pool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(echoCtx echo.Context) error {
			ctx := echoCtx.Request().Context()
			authorization, ok := echoCtx.Request().Header["Authorization"]
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			authUser, err := validateAuthorization(ctx, authorization[0])
			if err != nil {
				return err
			}

			conn, err := connPool.Acquire(ctx)
			if err != nil {
				return err
			}
			defer conn.Release()

			queries := db.New(conn)

			user, err := queries.GetUserBySub(ctx, authUser.Sub)
			if err != nil {
				echoCtx.Set(AuthUserContextKey, authUser)
			} else {
				userID, err := mapper.FromDbToProtoId(user.ID)
				if err != nil {
					return err
				}

				authUser.UserID = userID
				echoCtx.Set(AuthUserContextKey, authUser)
			}

			return next(echoCtx)
		}
	}
}

func validateAuthorization(ctx context.Context, authHeader string) (*AuthUser, error) {
	issuerURL, err := url.Parse(os.Getenv("SONGCONTESTRATERSERVICE_AUTH0_ISSUER_URL"))
	if err != nil {
		return nil, err
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
		return nil, err
	}

	var tokenPayloadBase64 string
	tokenPayloadBase64, err = getJwtPayload(authHeader)
	if err != nil {
		return nil, err
	}

	if _, err := jwtValidator.ValidateToken(ctx, tokenPayloadBase64); err != nil {
		log.Error().Err(err).Msg("could not validate token")
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "could not validate token")
	}

	var authUser AuthUser
	err = authUser.decode(tokenPayloadBase64)
	if err != nil {
		log.Error().Err(err).Msg("could not decode token")
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "could not decode token")
	}

	return &authUser, nil
}

func getJwtPayload(authHeader string) (string, error) {
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header")
	}

	return strings.Split(authHeader, " ")[1], nil
}

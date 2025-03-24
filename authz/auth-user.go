package authz

import (
	"net/http"
	"reflect"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/hyperremix/song-contest-rater-service/db"
	"github.com/hyperremix/song-contest-rater-service/mapper"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

type AuthUser struct {
	UserID    string
	ClerkUser *clerk.User
	DbUser    db.User
	Metadata  PublicMetadata
}

type PublicMetadata struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}

func (u *AuthUser) CheckIsAdmin() error {
	if u.Metadata.Role != "admin" {
		return echo.NewHTTPError(http.StatusForbidden, "missing permission to access this resource")
	}

	return nil
}

func (u *AuthUser) CheckIsOwner(obj any) error {
	dbId := reflect.ValueOf(&obj).Elem().Elem().FieldByName("UserID").Interface().(pgtype.UUID)
	id, err := mapper.FromDbToProtoId(dbId)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to convert id")
	}

	if u.UserID == id {
		return nil
	}

	return echo.NewHTTPError(http.StatusForbidden, "missing permission to access this resource")
}

func (u *AuthUser) CheckIsUser(user db.User) error {
	if u.ClerkUser.ID == user.Sub {
		return nil
	}

	return echo.NewHTTPError(http.StatusForbidden, "missing permission to access this resource")
}

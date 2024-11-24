package authz

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"reflect"
	"strings"

	"github.com/hyperremix/song-contest-rater-service/db"
	"github.com/hyperremix/song-contest-rater-service/mapper"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

type AuthUser struct {
	UserID      string
	User        db.User
	Permissions []string `json:"permissions"`
	Sub         string   `json:"sub"`
}

func (u *AuthUser) CheckHasPermission(p string) error {
	for _, perm := range u.Permissions {
		if perm == p {
			return nil
		}
	}

	return echo.NewHTTPError(http.StatusForbidden, "missing permission to access this resource")
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
	if u.Sub == user.Sub {
		return nil
	}

	return echo.NewHTTPError(http.StatusForbidden, "missing permission to access this resource")
}

func (u *AuthUser) decode(s string) error {
	barr, err := base64.RawStdEncoding.DecodeString(strings.Split(s, ".")[1])
	if err != nil {
		return err
	}

	return json.Unmarshal(barr, u)
}

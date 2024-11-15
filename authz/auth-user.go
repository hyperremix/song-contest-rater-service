package authz

import (
	"encoding/base64"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/hyperremix/song-contest-rater-service/db"
)

type AuthUser struct {
	UserID      string
	Permissions []string `json:"permissions"`
	Sub         string   `json:"sub"`
}

func (u *AuthUser) HasPermission(p string) bool {
	for _, perm := range u.Permissions {
		if perm == p {
			return true
		}
	}
	return false
}

func (u *AuthUser) IsOwner(obj any) bool {
	id := reflect.ValueOf(&obj).Elem().FieldByName("UserID").String()
	return u.UserID == id
}

func (u *AuthUser) IsUser(user db.User) bool {
	return u.Sub == user.Sub
}

func (u *AuthUser) decode(s string) error {
	barr, err := base64.RawStdEncoding.DecodeString(strings.Split(s, ".")[1])
	if err != nil {
		return err
	}
	return json.Unmarshal(barr, u)
}

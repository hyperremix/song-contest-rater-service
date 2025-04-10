package authz

import (
	"errors"
	"reflect"

	"connectrpc.com/connect"
	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/hyperremix/song-contest-rater-service/db"
	"github.com/hyperremix/song-contest-rater-service/mapper"
	"github.com/jackc/pgx/v5/pgtype"
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
		return connect.NewError(connect.CodePermissionDenied, errors.New("missing permission to access this resource"))
	}

	return nil
}

func (u *AuthUser) CheckIsOwner(obj any) error {

	dbId := reflect.ValueOf(&obj).Elem().Elem().FieldByName("UserID").Interface().(pgtype.UUID)
	id, err := mapper.FromDbToProtoId(dbId)
	if err != nil {
		return connect.NewError(connect.CodeInternal, errors.New("failed to convert id"))
	}

	if u.UserID == id {
		return nil
	}

	return connect.NewError(connect.CodePermissionDenied, errors.New("missing permission to access this resource"))
}

func (u *AuthUser) CheckIsUser(user db.User) error {

	if u.ClerkUser.ID == user.Sub {
		return nil
	}

	return connect.NewError(connect.CodePermissionDenied, errors.New("missing permission to access this resource"))
}

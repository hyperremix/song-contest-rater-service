package mapper

import (
	pb "github.com/hyperremix/song-contest-rater-protos/v3"
	"github.com/hyperremix/song-contest-rater-service/db"
	"github.com/jackc/pgx/v5/pgtype"
)

func FromDbUserListToResponse(u []db.User) (*pb.ListUsersResponse, error) {
	var users []*pb.UserResponse

	for _, user := range u {
		proto, err := FromDbUserToResponse(user)
		if err != nil {
			return nil, NewResponseBindingError(err)
		}

		users = append(users, proto)
	}

	return &pb.ListUsersResponse{Users: users}, nil
}

func FromDbUserToResponse(u db.User) (*pb.UserResponse, error) {
	id, err := FromDbToProtoId(u.ID)
	if err != nil {
		return nil, NewResponseBindingError(err)
	}

	return &pb.UserResponse{
		Id:        id,
		Email:     u.Email,
		Firstname: u.Firstname,
		Lastname:  u.Lastname,
		ImageUrl:  u.ImageUrl,
		CreatedAt: fromDbToProtoTimestamp(u.CreatedAt),
		UpdatedAt: fromDbToProtoTimestamp(u.UpdatedAt),
	}, nil
}

func FromCreateRequestToInsertUser(c *pb.CreateUserRequest, sub string) (db.InsertUserParams, error) {
	return db.InsertUserParams{
		Sub:       sub,
		Email:     c.Email,
		Firstname: c.Firstname,
		Lastname:  c.Lastname,
		ImageUrl:  c.ImageUrl,
	}, nil
}

func FromUpdateRequestToUpdateUser(c *pb.UpdateUserRequest) (db.UpdateUserParams, error) {
	id, err := FromProtoToDbId(c.Id)
	if err != nil {
		return db.UpdateUserParams{}, NewRequestBindingError(err)
	}

	return db.UpdateUserParams{
		ID:        id,
		Firstname: c.Firstname,
		Lastname:  c.Lastname,
		ImageUrl:  c.ImageUrl,
	}, nil
}

func ToUpdateUserParams(id pgtype.UUID, firstname string, lastname string, imageUrl string) (db.UpdateUserParams, error) {

	return db.UpdateUserParams{
		ID:        id,
		Firstname: firstname,
		Lastname:  lastname,
		ImageUrl:  imageUrl,
	}, nil
}

func FromProfilePictureToUpdateUserImageUrl(userId string, imageUrl string) (db.UpdateUserImageUrlParams, error) {
	id, err := FromProtoToDbId(userId)
	if err != nil {
		return db.UpdateUserImageUrlParams{}, NewRequestBindingError(err)
	}

	return db.UpdateUserImageUrlParams{
		ID:       id,
		ImageUrl: imageUrl,
	}, nil
}

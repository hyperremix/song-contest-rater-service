package mapper

import (
	"github.com/hyperremix/song-contest-rater-service/db"
	pb "github.com/hyperremix/song-contest-rater-service/protos/songcontestrater"
)

func FromDbUserListToResponse(u []db.User) (*pb.ListUsersResponse, error) {
	var users []*pb.UserResponse

	for _, user := range u {
		proto, err := FromDbUserToResponse(user)
		if err != nil {
			return nil, err
		}

		users = append(users, proto)
	}

	return &pb.ListUsersResponse{Users: users}, nil
}

func FromDbUserToResponse(u db.User) (*pb.UserResponse, error) {
	id, err := fromDbToProtoId(u.ID)
	if err != nil {
		return nil, err
	}

	return &pb.UserResponse{
		Id:        id,
		Email:     u.Email,
		Firstname: u.Firstname,
		Lastname:  u.Lastname,
		ImageUrl:  u.ImageUrl,
		CreatedAt: mapFromDbToProtoTimestamp(u.CreatedAt),
		UpdatedAt: mapFromDbToProtoTimestamp(u.UpdatedAt),
	}, nil
}

func FromCreateRequestToInsertUser(c *pb.CreateUserRequest) (db.InsertUserParams, error) {
	return db.InsertUserParams{
		Email:     c.Email,
		Firstname: c.Firstname,
		Lastname:  c.Lastname,
		ImageUrl:  c.ImageUrl,
	}, nil
}

func FromUpdateRequestToUpdateUser(c *pb.UpdateUserRequest) (db.UpdateUserParams, error) {
	id, err := FromProtoToDbId(c.Id)
	if err != nil {
		return db.UpdateUserParams{}, err
	}

	return db.UpdateUserParams{
		ID:        id,
		Email:     c.Email,
		Firstname: c.Firstname,
		Lastname:  c.Lastname,
		ImageUrl:  c.ImageUrl,
	}, nil
}

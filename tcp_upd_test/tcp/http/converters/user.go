package converters

import (
	"tcp_upd_test/tcp/http/models"
	"tcp_upd_test/tcp/http/models/dto"
)

func FromUserToUserResponse(user models.User) dto.UserResponse {
	var surname string

	if user.Surname != nil {
		surname = *user.Surname
	}

	return dto.UserResponse{
		Id:      user.ID,
		Name:    user.Name,
		Surname: surname,
		Email:   user.Email,
	}
}

func FromUserToCreateUserResponse(user models.User) dto.CreateUserResponse {
	return dto.CreateUserResponse{
		Id: user.ID,
	}
}

func FromCreateUserRequestToUser(userReq dto.CreateUserRequest) models.User {

	var surname string
	if userReq.Surname != nil {
		surname = *userReq.Surname
	}

	var age int64
	if userReq.Age != nil {
		age = *userReq.Age
	}

	return models.User{
		Email:    *userReq.Email,
		Name:     *userReq.Name,
		Password: *userReq.Password,
		Surname:  &surname,
		Age:      &age,
	}
}

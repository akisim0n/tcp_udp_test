package dto

type CreateUserRequest struct {
	Name     *string `json:"name,omitempty"`
	Age      *int64  `json:"age,omitempty"`
	Surname  *string `json:"surname,omitempty"`
	Email    *string `json:"email" validate:"required"`
	Password *string `json:"password" validate:"required"`
}

type UpdateUserRequest struct {
	Name     *string `json:"name,omitempty"`
	Age      *int64  `json:"age,omitempty"`
	Surname  *string `json:"surname,omitempty"`
	Email    *string `json:"email,omitempty"`
	Password *string `json:"password,omitempty"`
}

type CreateUserResponse struct {
	Id int64 `json:"id"`
}

type UserResponse struct {
	Id      int64  `json:"id"`
	Name    string `json:"name"`
	Surname string `json:"surname"`
	Email   string `json:"email"`
}

type UserListResponse struct {
	Users []UserResponse `json:"users"`
	Total int            `json:"total"`
}

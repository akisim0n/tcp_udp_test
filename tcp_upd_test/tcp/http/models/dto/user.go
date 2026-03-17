package dto

type CreateUserRequest struct {
	Name     *string `json:"name"`
	Age      *int64  `json:"age"`
	Surname  *string `json:"surname"`
	Email    *string `json:"email"`
	Password *string `json:"password"`
}

type CreateUserResponse struct {
	Id int64 `json:"id"`
}

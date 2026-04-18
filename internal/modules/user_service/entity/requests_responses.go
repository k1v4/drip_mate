package entity

import "github.com/k1v4/drip_mate/internal/entity"

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=10"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessId entity.Role `json:"access_id"`
}

type DeleteUserResponse struct {
	IsSuccessfully bool `json:"is_successfully"`
}

type UpdateUserRequest struct {
	UserID   int64  `json:"user_id"   validate:"required"`
	Email    string `json:"email"     validate:"omitempty,email"`
	Password string `json:"password"  validate:"omitempty,min=10"`
	Username string `json:"username"  validate:"omitempty,min=3,max=50"`
	Name     string `json:"name"      validate:"omitempty,min=1,max=100"`
	Surname  string `json:"surname"   validate:"omitempty,min=1,max=100"`
	City     string `json:"city"      validate:"omitempty,min=1,max=100"`
}

type UpdateUserResponse struct {
	User User `json:"user"`
}

package entity

type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	AccessId     int    `json:"access_id"`
}

type IsAdminRequest struct {
	UserID int64 `json:"user_id"`
}

type IsAdminResponse struct {
	IsAdmin bool `json:"is_admin"`
}

type GetUserRequestById struct {
	UserID int64 `json:"user_id"`
}

type GetUserResponseById struct {
	User User `json:"user"`
}

type DeleteUserRequest struct {
	UserID int64 `json:"user_id"`
}

type DeleteUserResponse struct {
	IsSuccessfully bool `json:"is_successfully"`
}

type UpdateUserRequest struct {
	UserID   int64  `json:"user_id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Surname  string `json:"surname"`
}

type UpdateUserResponse struct {
	User User `json:"user"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

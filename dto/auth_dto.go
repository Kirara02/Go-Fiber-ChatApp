package dto

type LoginResponse struct {
	User  UserResponse `json:"user"`
	AccessToken string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}


type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}


type RegisterRequest struct {
    Name     string `json:"name" validate:"required,min=3"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=6"`
}


type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}


type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
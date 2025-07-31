package dto

import (
	"main/domain"
	"mime/multipart"
	"time"
)

type CreateUserRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type UpdateUserRequest struct {
	Name         *string               `json:"name,omitempty"`
	Email        *string               `json:"email,omitempty"`
	ProfileImage *multipart.FileHeader `form:"profile_image"`
}

type UserResponse struct {
	ID           uint      `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	ProfileImage *string   `json:"profile_image"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func ToUserResponse(user *domain.User) UserResponse {
	return UserResponse{
		ID:           user.ID,
		Name:         user.Name,
		Email:        user.Email,
		ProfileImage: user.ProfileImage,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}

func ToUserResponses(users []*domain.User) []UserResponse {
	var userResponses []UserResponse

	if len(users) == 0 {
		return userResponses
	}

	for _, u := range users {
		userResponses = append(userResponses, ToUserResponse(u))
	}

	return userResponses
}

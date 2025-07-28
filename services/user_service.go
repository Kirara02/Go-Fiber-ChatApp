package services

import (
	"errors"
	"main/domain"
	"main/dto"
	"main/repository"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)


type UserService interface {
	CreateUser(req dto.CreateUserRequest) (*dto.UserResponse, error)
	GetUserByID(id uint) (*dto.UserResponse, error)
	GetAllUsers(keyword string, currentUserID uint, includeSelf bool) ([]dto.UserResponse, error)
	UpdateUser(id uint, req dto.UpdateUserRequest) (*dto.UserResponse, error)
	DeleteUser(id uint) error
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) CreateUser(req dto.CreateUserRequest) (*dto.UserResponse, error) {
	existingUser, err := s.userRepo.GetUserByEmail(req.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("email sudah terdaftar")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	newUser := &domain.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	if err := s.userRepo.CreateUser(newUser); err != nil {
		return nil, err
	}

	response := dto.ToUserResponse(newUser)
	return &response, nil
}

func (s *userService) GetUserByID(id uint) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetUserByID(id)
	if err != nil {
		return nil, err
	}

	response := dto.ToUserResponse(user)
	return &response, nil
}

func (s *userService) GetAllUsers(keyword string, currentUserID uint, includeSelf bool) ([]dto.UserResponse, error) {
	users, err := s.userRepo.GetAllUsers(keyword, currentUserID, includeSelf)

	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return []dto.UserResponse{}, nil
	}

	responses := dto.ToUserResponses(users)
	return responses, nil
}

func (s *userService) UpdateUser(id uint, req dto.UpdateUserRequest) (*dto.UserResponse, error) {
	userToUpdate, err := s.userRepo.GetUserByID(id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		userToUpdate.Name = *req.Name
	}

	if req.Email != nil {
		userToUpdate.Email = *req.Email
	}

	if req.Password != nil {
		if *req.Password == "" {
			return nil, errors.New("password baru tidak boleh kosong")
		}
		
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		userToUpdate.Password = string(hashedPassword)
	}
	
	if err := s.userRepo.UpdateUser(userToUpdate); err != nil {
		return nil, err
	}

	response := dto.ToUserResponse(userToUpdate)
	return &response, nil
}

func (s *userService) DeleteUser(id uint) error {
	_, err := s.userRepo.GetUserByID(id)
	if err != nil {
		return err
	}
	return s.userRepo.DeleteUser(id)
}

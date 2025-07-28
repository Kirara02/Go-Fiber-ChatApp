package repository

import (
	"main/domain"

	"gorm.io/gorm"
)

type UserRepository interface {
	CreateUser(user *domain.User) error
	GetUserByEmail(email string) (*domain.User, error)
	GetUserByID(id uint) (*domain.User, error)
	GetUsersByIDs(ids []uint) ([]*domain.User, error)
	GetAllUsers(keyword string, currentUserID uint, includeSelf bool) ([]*domain.User, error)
	UpdateUser(user *domain.User) error
	DeleteUser(id uint) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(user *domain.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) GetUserByEmail(email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetUserByID(id uint) (*domain.User, error) {
	var user domain.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetUsersByIDs(ids []uint) ([]*domain.User, error) {
	var users []*domain.User
	
	if err := r.db.Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil, err
	}
	
	return users, nil
}

func (r *userRepository) GetAllUsers(keyword string, currentUserID uint, includeSelf bool) ([]*domain.User, error) {
	var users []*domain.User

	db := r.db
	
	if !includeSelf {
		db = db.Where("id != ?", currentUserID)
	}
	
	if keyword != "" {
		searchQuery := "%" + keyword + "%"
		db = db.Where("name ILIKE ?", searchQuery)
	}

	if err := db.Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

func (r *userRepository) UpdateUser(user *domain.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) DeleteUser(id uint) error {
	return r.db.Delete(&domain.User{}, id).Error
}
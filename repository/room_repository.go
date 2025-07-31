package repository

import (
	"errors"
	"main/domain"

	"gorm.io/gorm"
)

type RoomRepository interface {
	SaveRoom(room *domain.Room) error
	CreateRoom(room *domain.Room, memberIDs []uint) (*domain.Room, error)
	GetRoomByID(id uint) (*domain.Room, error)
	GetUserRoomsWithDetails(userID uint) ([]*domain.Room, error)
	GetSimpleUserRooms(userID uint) ([]*domain.Room, error)
	CheckUserInRoom(userID, roomID uint) (bool, error)
	FindPrivateRoomByMembers(memberIDs []uint) (*domain.Room, error)
	UpdateRoom(room *domain.Room) error
}

type roomRepository struct {
	db *gorm.DB
}

func NewRoomRepository(db *gorm.DB) RoomRepository {
	return &roomRepository{db: db}
}

func (r *roomRepository) SaveRoom(room *domain.Room) error {
	return r.db.Save(room).Error
}

func (r *roomRepository) CreateRoom(room *domain.Room, memberIDs []uint) (*domain.Room, error) {
	tx := r.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	if err := tx.Create(room).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	var members []*domain.User
	if err := tx.Where("id IN ?", memberIDs).Find(&members).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if len(members) == 0 {
		tx.Rollback()
		return nil, errors.New("tidak ada anggota valid yang ditemukan untuk membuat room")
	}

	if err := tx.Model(room).Association("Users").Append(members); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	err := r.db.Preload("Users").First(room, room.ID).Error
	return room, err
}

func (r *roomRepository) GetRoomByID(id uint) (*domain.Room, error) {
	var room domain.Room
	if err := r.db.Preload("Users").First(&room, id).Error; err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *roomRepository) GetUserRoomsWithDetails(userID uint) ([]*domain.Room, error) {
	var user domain.User
	if err := r.db.Preload("Rooms.Users").First(&user, userID).Error; err != nil {
		return nil, err
	}
	if len(user.Rooms) == 0 {
		return []*domain.Room{}, nil
	}

	var roomIDs []uint
	for _, room := range user.Rooms {
		roomIDs = append(roomIDs, room.ID)
	}

	var lastMessages []domain.ChatMessage

	subQuery := r.db.Model(&domain.ChatMessage{}).Select("MAX(id)").Where("room_id IN ?", roomIDs).Group("room_id")
	if err := r.db.Where("id IN (?)", subQuery).Order("created_at desc").Find(&lastMessages).Error; err != nil {
		return nil, err
	}

	lastMessageMap := make(map[uint]domain.ChatMessage)
	for _, msg := range lastMessages {
		lastMessageMap[msg.RoomID] = msg
	}

	for _, room := range user.Rooms {
		if lastMsg, ok := lastMessageMap[room.ID]; ok {
			room.LastMessage = lastMsg
		}
	}

	return user.Rooms, nil
}

func (r *roomRepository) GetSimpleUserRooms(userID uint) ([]*domain.Room, error) {
	var user domain.User

	err := r.db.
		Preload("Rooms", func(db *gorm.DB) *gorm.DB {
			return db.Order("rooms.created_at DESC")
		}).
		Preload("Rooms.Users").
		First(&user, userID).Error

	if err != nil {
		return nil, err
	}

	return user.Rooms, nil
}

func (r *roomRepository) CheckUserInRoom(userID, roomID uint) (bool, error) {
	var count int64
	err := r.db.Table("user_rooms").Where("user_id = ? AND room_id = ?", userID, roomID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *roomRepository) FindPrivateRoomByMembers(memberIDs []uint) (*domain.Room, error) {
	var room domain.Room

	// Ini adalah query yang agak kompleks, mari kita bedah:
	// 1. `SELECT room_id FROM user_rooms WHERE user_id IN (?) GROUP BY room_id HAVING COUNT(DISTINCT user_id) = ?`
	//    - Subquery ini menemukan semua `room_id` yang memiliki jumlah anggota yang sama persis dengan yang kita cari.
	// 2. `SELECT * FROM rooms WHERE id IN (...) AND is_private = true`
	//    - Query utama mencari room berdasarkan ID yang ditemukan di subquery, dan memastikan itu adalah room privat.
	// 3. `HAVING (SELECT COUNT(*) FROM user_rooms WHERE room_id = rooms.id) = ?`
	//    - Ini adalah lapisan verifikasi kedua untuk memastikan room tersebut tidak memiliki anggota lain selain yang kita cari.

	err := r.db.Joins("JOIN user_rooms ur ON ur.room_id = rooms.id").
		Where("rooms.is_private = ?", true).
		Where("ur.user_id IN ?", memberIDs).
		Group("rooms.id").
		Having("COUNT(DISTINCT ur.user_id) = ?", len(memberIDs)).
		Having("(SELECT COUNT(*) FROM user_rooms WHERE room_id = rooms.id) = ?", len(memberIDs)).
		Preload("Users").
		First(&room).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &room, err
}

func (r *roomRepository) UpdateRoom(room *domain.Room) error {
	return r.db.Save(room).Error
}
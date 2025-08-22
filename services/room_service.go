package services

import (
	"errors"
	"fmt"
	"main/domain"
	"main/dto"
	"main/repository"
	"main/utils"
	"mime/multipart"
	"sort"
	"time"
)

type RoomService interface {
	GetOrCreateDirectMessageRoom(user1ID, user2ID uint) (*dto.RoomResponse, error)
	CreateGroupRoom(req dto.CreateRoomRequest, creatorID uint) (*dto.RoomResponse, error)
	GetMyRooms(userID uint, view string, includeMembers bool, showEmpty bool) ([]dto.RoomResponse, error)
	IsUserMember(userID, roomID uint) (bool, error)
	GetRoomByID(roomID uint) (*domain.Room, error)
	UpdateRoomImage(roomID uint, currentUserID uint, file *multipart.FileHeader) (*dto.RoomResponse, error)
}

type roomService struct {
	roomRepo      repository.RoomRepository
	userRepo      repository.UserRepository
	uploadService UploadService
}

func NewRoomService(roomRepo repository.RoomRepository, userRepo repository.UserRepository, upload UploadService) RoomService {
	return &roomService{
		roomRepo:      roomRepo,
		userRepo:      userRepo,
		uploadService: upload,
	}
}

func (s *roomService) CreateGroupRoom(req dto.CreateRoomRequest, creatorID uint) (*dto.RoomResponse, error) {
	if req.Name == "" {
		return nil, errors.New("nama grup wajib diisi")
	}

	memberIDs := append(req.UserIDs, creatorID)
	memberIDs = utils.UniqueUintSlice(memberIDs)

	if len(memberIDs) < 3 {
		return nil, errors.New("sebuah grup membutuhkan minimal 3 anggota")
	}

	members, err := s.userRepo.GetUsersByIDs(memberIDs)
	if err != nil || len(members) != len(memberIDs) {
		return nil, errors.New("satu atau lebih ID pengguna tidak valid")
	}

	newRoom := &domain.Room{
		Name:    req.Name,
		Type:    domain.RoomTypeGroup, // <-- Tipe diatur secara eksplisit
		OwnerID: &creatorID,
	}

	createdRoom, err := s.roomRepo.CreateRoom(newRoom, memberIDs)
	if err != nil {
		return nil, err
	}

	response := dto.ToRoomResponse(createdRoom, creatorID, true)
	return &response, nil
}

func (s *roomService) GetOrCreateDirectMessageRoom(user1ID, user2ID uint) (*dto.RoomResponse, error) {
	if user1ID == user2ID {
		return nil, errors.New("tidak bisa membuat DM dengan diri sendiri")
	}
	memberIDs := []uint{user1ID, user2ID}

	// Cek apakah DM sudah ada
	existingRoom, err := s.roomRepo.FindDirectMessageRoomByMembers(memberIDs)
	if err != nil {
		return nil, err // Error database
	}
	if existingRoom != nil {
		// DM sudah ada, kembalikan saja
		response := dto.ToRoomResponse(existingRoom, user1ID, true)
		return &response, nil
	}

	// Buat DM baru jika belum ada
	newRoom := &domain.Room{
		Type: domain.RoomTypeDM, // <-- Tipe diatur secara eksplisit
	}

	createdRoom, err := s.roomRepo.CreateRoom(newRoom, memberIDs)
	if err != nil {
		return nil, err
	}

	response := dto.ToRoomResponse(createdRoom, user1ID, true)
	return &response, nil
}

func (s *roomService) GetMyRooms(userID uint, view string, includeMembers bool, showEmpty bool) ([]dto.RoomResponse, error) {
	var rooms []*domain.Room
	var err error

	if view == "simple" {
		rooms, err = s.roomRepo.GetSimpleUserRooms(userID)
	} else {
		rooms, err = s.roomRepo.GetUserRoomsWithDetails(userID)

		if err == nil && len(rooms) > 0 {
			// Sort berdasarkan waktu terakhir
			sort.Slice(rooms, func(i, j int) bool {
				var timeI, timeJ time.Time
				if rooms[i].LastMessage.ID != 0 {
					timeI = rooms[i].LastMessage.CreatedAt
				} else {
					timeI = rooms[i].CreatedAt
				}
				if rooms[j].LastMessage.ID != 0 {
					timeJ = rooms[j].LastMessage.CreatedAt
				} else {
					timeJ = rooms[j].CreatedAt
				}
				return timeI.After(timeJ)
			})
		}
	}

	if err != nil {
		return nil, err
	}

	// ✅ Filter jika showEmpty == false
	if !showEmpty {
		filteredRooms := []*domain.Room{}
		for _, room := range rooms {
			if room.LastMessage.ID != 0 {
				filteredRooms = append(filteredRooms, room)
			}
		}
		rooms = filteredRooms
	}

	responses := dto.ToRoomResponses(rooms, userID, includeMembers)
	return responses, nil
}

func (s *roomService) IsUserMember(userID, roomID uint) (bool, error) {
	return s.roomRepo.CheckUserInRoom(userID, roomID)
}

func (s *roomService) GetRoomByID(roomID uint) (*domain.Room, error) {
	return s.roomRepo.GetRoomByID(roomID)
}

func (s *roomService) UpdateRoomImage(roomID uint, currentUserID uint, file *multipart.FileHeader) (*dto.RoomResponse, error) {
	room, err := s.roomRepo.GetRoomByID(roomID)
	if err != nil {
		return nil, err
	}

	// Hapus gambar lama kalau ada
	if room.RoomImage != nil {
		oldPublicID := utils.ExtractPublicIDFromURL(*room.RoomImage)
		_ = s.uploadService.DeleteFile(oldPublicID)
	}

	// Generate publicID baru untuk image-nya
	safeName := utils.SanitizeFilename(room.Name)
	publicID := fmt.Sprintf("%s_%d", safeName, room.ID)

	imageUrl, err := s.uploadService.UploadFile(file, "rooms", publicID)
	if err != nil {
		return nil, errors.New("gagal mengunggah gambar room")
	}

	room.RoomImage = &imageUrl

	if err := s.roomRepo.UpdateRoom(room); err != nil {
		return nil, errors.New("gagal menyimpan URL gambar ke room")
	}

	resp := dto.ToRoomResponse(room, currentUserID, false)
	return &resp, nil
}

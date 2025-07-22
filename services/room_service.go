package services

import (
	"errors"
	"main/domain"
	"main/dto"
	"main/repository"
	"main/utils"
	"sort"
	"strings"
)

type RoomService interface {
	CreateRoom(req dto.CreateRoomRequest, creatorID uint) (*dto.RoomResponse, error)
	GetMyRooms(userID uint) ([]dto.RoomResponse, error)
	IsUserMember(userID, roomID uint) (bool, error)
}

type roomService struct {
	roomRepo repository.RoomRepository
	userRepo repository.UserRepository
}

func NewRoomService(roomRepo repository.RoomRepository, userRepo repository.UserRepository) RoomService {
	return &roomService{
		roomRepo: roomRepo,
		userRepo: userRepo,
	}
}

func (s *roomService) CreateRoom(req dto.CreateRoomRequest, creatorID uint) (*dto.RoomResponse, error) {

	memberIDs := append(req.UserIDs, creatorID)
	memberIDs = utils.UniqueUintSlice(memberIDs)

	if len(memberIDs) < 2 {
		return nil, errors.New("sebuah room membutuhkan minimal 2 anggota")
	}
	
	if len(memberIDs) == 2 {
		existingRoom, err := s.roomRepo.FindPrivateRoomByMembers(memberIDs)
		if err != nil {
			return nil, err
		}
		if existingRoom != nil {
			return nil, errors.New("direct message dengan pengguna ini sudah ada")
		}
	}

	members, err := s.userRepo.GetUsersByIDs(memberIDs)
	if err != nil || len(members) != len(memberIDs) {
		return nil, errors.New("satu atau lebih ID pengguna tidak valid")
	}
	
	newRoom := &domain.Room{
		Name: req.Name,
	}

	if len(memberIDs) > 2 {
		// GRUP
		if req.Name == "" {
			return nil, errors.New("nama grup wajib diisi untuk room dengan lebih dari 2 anggota")
		}
		newRoom.OwnerID = &creatorID 
	} else {
		// DIRECT MESSAGE (DM)
		newRoom.IsPrivate = true
		
		var names []string
		for _, member := range members {
			names = append(names, member.Name)
		}
		sort.Strings(names)
		
		if newRoom.Name == "" {
			newRoom.Name = strings.Join(names, " & ")
		}
	}
	
	createdRoom, err := s.roomRepo.CreateRoom(newRoom, memberIDs)
	if err != nil {
		return nil, err
	}
	
	response := dto.ToRoomResponse(createdRoom)
	return &response, nil
}


func (s *roomService) GetMyRooms(userID uint) ([]dto.RoomResponse, error) {
	rooms, err := s.roomRepo.GetUserRooms(userID)
	if err != nil {
		return nil, err
	}
	
	responses := []dto.RoomResponse{}

	if len(rooms) > 0 {
		responses = dto.ToRoomResponses(rooms)
	}

	return responses, nil
}

func (s *roomService) IsUserMember(userID, roomID uint) (bool, error) {
	return s.roomRepo.CheckUserInRoom(userID, roomID)
}
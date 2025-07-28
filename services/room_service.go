package services

import (
	"errors"
	"main/domain"
	"main/dto"
	"main/repository"
	"main/utils"
	"sort"
	"strings"
	"time"
)

type RoomService interface {
	CreateRoom(req dto.CreateRoomRequest, creatorID uint) (*dto.RoomResponse, error)
	GetMyRooms(userID uint, view string, includeMembers bool) ([]dto.RoomResponse, error)
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
	
	response := dto.ToRoomResponse(createdRoom, creatorID, true)
	return &response, nil
}


func (s *roomService) GetMyRooms(userID uint, view string, includeMembers bool) ([]dto.RoomResponse, error) {
	var rooms []*domain.Room
	var err error

	if view == "simple" {
		rooms, err = s.roomRepo.GetSimpleUserRooms(userID)
	} else {
		rooms, err = s.roomRepo.GetUserRoomsWithDetails(userID)
		
		if err == nil && len(rooms) > 0 {
			sort.Slice(rooms, func(i, j int) bool {
				var timeI, timeJ time.Time
				if rooms[i].LastMessage.ID != 0 { timeI = rooms[i].LastMessage.CreatedAt } else { timeI = rooms[i].CreatedAt }
				if rooms[j].LastMessage.ID != 0 { timeJ = rooms[j].LastMessage.CreatedAt } else { timeJ = rooms[j].CreatedAt }
				return timeI.After(timeJ)
			})
		}
	}

	if err != nil {
		return nil, err
	}
	
	responses := dto.ToRoomResponses(rooms, userID, includeMembers)

	return responses, nil
}

func (s *roomService) IsUserMember(userID, roomID uint) (bool, error) {
	return s.roomRepo.CheckUserInRoom(userID, roomID)
}
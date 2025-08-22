package handlers

import (
	"main/dto"
	"main/services"
	"main/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type RoomHandler struct {
	roomService services.RoomService
}

func NewRoomHandler(roomService services.RoomService) *RoomHandler {
	return &RoomHandler{roomService: roomService}
}

// --- HANDLER BARU UNTUK MEMBUAT GRUP ---
func (h *RoomHandler) CreateGroup(c *fiber.Ctx) error {
	creatorID, err := utils.ExtractUserIDFromContext(c)
	if err != nil {
		return err
	}

	var req dto.CreateRoomRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Request body tidak valid")
	}

	// Panggil service baru yang spesifik untuk grup
	roomResponse, err := h.roomService.CreateGroupRoom(req, creatorID)
	if err != nil {
		// Service akan memberikan pesan error yang relevan (misal: "nama grup wajib diisi")
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(utils.BaseResponse{
		Success: true,
		Message: "Grup berhasil dibuat",
		Data:    roomResponse,
	})
}

func (h *RoomHandler) GetOrCreateDM(c *fiber.Ctx) error {
	myUserID, err := utils.ExtractUserIDFromContext(c)
	if err != nil {
		return err
	}
	var req struct {
		TargetUserID uint `json:"target_user_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Request body tidak valid: 'target_user_id' dibutuhkan")
	}

	if req.TargetUserID == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "'target_user_id' tidak boleh kosong")
	}

	// Panggil service baru yang spesifik untuk DM
	roomResponse, err := h.roomService.GetOrCreateDirectMessageRoom(myUserID, req.TargetUserID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(utils.BaseResponse{
		Success: true,
		Message: "Room DM berhasil didapatkan",
		Data:    roomResponse,
	})
}

// --- HANDLER YANG SUDAH ADA (TIDAK BERUBAH) ---

func (h *RoomHandler) GetMyRooms(c *fiber.Ctx) error {
	userID, err := utils.ExtractUserIDFromContext(c)
	if err != nil {
		return err
	}

	view := c.Query("view", "detailed")
	includeMembers := c.Query("include_members", "true") == "true"
	showEmpty := c.Query("show_empty", "true") == "true"

	rooms, err := h.roomService.GetMyRooms(userID, view, includeMembers, showEmpty)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Gagal mengambil daftar room")
	}

	return c.Status(fiber.StatusOK).JSON(utils.BaseResponse{
		Success: true,
		Message: "Daftar room berhasil diambil",
		Data:    rooms,
	})
}

func (h *RoomHandler) GetRoomByID(c *fiber.Ctx) error {
	roomIDParam, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "ID room tidak valid")
	}
	roomID := uint(roomIDParam)

	currentUserID, err := utils.ExtractUserIDFromContext(c)
	if err != nil {
		return err
	}

	room, err := h.roomService.GetRoomByID(roomID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	isMember, err := h.roomService.IsUserMember(currentUserID, roomID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Gagal memverifikasi keanggotaan room")
	}
	if !isMember {
		return fiber.NewError(fiber.StatusForbidden, "Kamu bukan anggota room ini")
	}

	roomResp := dto.ToRoomResponse(room, currentUserID, true)

	return c.Status(fiber.StatusOK).JSON(utils.BaseResponse{
		Success: true,
		Message: "Detail room berhasil diambil",
		Data:    roomResp,
	})
}

func (h *RoomHandler) UpdateRoomImage(c *fiber.Ctx) error {
	roomIDParam, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "ID room tidak valid")
	}
	roomID := uint(roomIDParam)

	currentUserID, err := utils.ExtractUserIDFromContext(c)
	if err != nil {
		return err
	}

	file, err := c.FormFile("room_image")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "File gambar room tidak ditemukan")
	}

	roomResponse, err := h.roomService.UpdateRoomImage(roomID, currentUserID, file)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(utils.BaseResponse{
		Success: true,
		Message: "Gambar room berhasil diperbarui",
		Data:    roomResponse,
	})
}

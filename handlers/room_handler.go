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

func (h *RoomHandler) CreateRoom(c *fiber.Ctx) error {
	creatorIDLocals := c.Locals("user_id")
	if creatorIDLocals == nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Gagal mendapatkan ID pembuat dari token")
	}
	creatorID, ok := creatorIDLocals.(float64)
	if !ok {
		return fiber.NewError(fiber.StatusInternalServerError, "Tipe ID pembuat tidak valid di context")
	}

	var req dto.CreateRoomRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Request body tidak valid")
	}

	roomResponse, err := h.roomService.CreateRoom(req, uint(creatorID))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(utils.BaseResponse{
		Success: true,
		Message: "Room berhasil dibuat",
		Data:    roomResponse,
	})
}

func (h *RoomHandler) GetMyRooms(c *fiber.Ctx) error {
	userIDLocals := c.Locals("user_id")
	if userIDLocals == nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Gagal mendapatkan ID pengguna dari token")
	}
	userID, ok := userIDLocals.(float64)
	if !ok {
		return fiber.NewError(fiber.StatusInternalServerError, "Tipe ID pengguna tidak valid di context")
	}

	view := c.Query("view", "detailed")
	includeMembers := c.Query("include_members", "true") == "true"

	showEmpty := c.Query("show_empty", "true") == "true"

	rooms, err := h.roomService.GetMyRooms(uint(userID), view, includeMembers, showEmpty)
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

	userIDLocals := c.Locals("user_id")
	if userIDLocals == nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Gagal mendapatkan ID pengguna dari token")
	}
	userIDFloat, ok := userIDLocals.(float64)
	if !ok {
		return fiber.NewError(fiber.StatusInternalServerError, "Tipe ID pengguna tidak valid")
	}
	currentUserID := uint(userIDFloat)

	// Ambil room dari service
	room, err := h.roomService.GetRoomByID(roomID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}

	// Cek apakah user termasuk anggota room
	isMember, err := h.roomService.IsUserMember(currentUserID, roomID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Gagal memverifikasi keanggotaan room")
	}
	if !isMember {
		return fiber.NewError(fiber.StatusForbidden, "Kamu bukan anggota room ini")
	}

	// Konversi ke response
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

	userIDLocals := c.Locals("user_id")
	if userIDLocals == nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Gagal mendapatkan ID pengguna dari token")
	}
	userIDFloat, ok := userIDLocals.(float64)
	if !ok {
		return fiber.NewError(fiber.StatusInternalServerError, "Tipe ID pengguna tidak valid")
	}
	currentUserID := uint(userIDFloat)

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

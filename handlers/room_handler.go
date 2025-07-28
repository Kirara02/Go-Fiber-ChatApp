package handlers

import (
	"main/dto"
	"main/services"
	"main/utils"

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

	rooms, err := h.roomService.GetMyRooms(uint(userID), view, includeMembers)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Gagal mengambil daftar room")
	}

	return c.Status(fiber.StatusOK).JSON(utils.BaseResponse{
		Success: true,
		Message: "Daftar room berhasil diambil",
		Data:    rooms,
	})
}
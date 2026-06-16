package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/gilabs/indosupplier/api/internal/chat/domain/dto"
	"github.com/gilabs/indosupplier/api/internal/chat/domain/usecase"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/jwt"
	"github.com/gilabs/indosupplier/api/internal/core/response"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for dev
	},
}

type ChatHandler struct {
	chatUC     usecase.ChatUsecase
	jwtManager *jwt.JWTManager
}

func NewChatHandler(chatUC usecase.ChatUsecase, jwtManager *jwt.JWTManager) *ChatHandler {
	return &ChatHandler{
		chatUC:     chatUC,
		jwtManager: jwtManager,
	}
}

func (h *ChatHandler) ListRooms(c *gin.Context) {
	userID := c.GetString("user_id")
	rooms, err := h.chatUC.ListRooms(c.Request.Context(), userID)
	if err != nil {
		if err == usecase.ErrBuyerProfileNotFound {
			response.ErrorResponse(c, http.StatusNotFound, "BUYER_PROFILE_NOT_FOUND", err.Error(), nil, nil)
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to retrieve rooms", nil, nil)
		return
	}
	response.SuccessResponse(c, rooms, nil)
}

func (h *ChatHandler) GetOrCreateRoom(c *gin.Context) {
	userID := c.GetString("user_id")
	var req dto.CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body", nil, nil)
		return
	}

	room, err := h.chatUC.GetOrCreateRoom(c.Request.Context(), userID, &req)
	if err != nil {
		if err == usecase.ErrBuyerProfileNotFound {
			response.ErrorResponse(c, http.StatusNotFound, "BUYER_PROFILE_NOT_FOUND", err.Error(), nil, nil)
			return
		}
		if err == usecase.ErrSupplierProfileNotFound {
			response.ErrorResponse(c, http.StatusNotFound, "SUPPLIER_PROFILE_NOT_FOUND", err.Error(), nil, nil)
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to get or create room", nil, nil)
		return
	}
	response.SuccessResponseCreated(c, room, nil)
}

func (h *ChatHandler) GetRoomMessages(c *gin.Context) {
	userID := c.GetString("user_id")
	roomID := c.Param("id")

	messages, err := h.chatUC.GetRoomMessages(c.Request.Context(), userID, roomID)
	if err != nil {
		if err == usecase.ErrRoomNotFound {
			response.ErrorResponse(c, http.StatusNotFound, "ROOM_NOT_FOUND", err.Error(), nil, nil)
			return
		}
		if err == usecase.ErrRoomAccessDenied {
			response.ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", err.Error(), nil, nil)
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to get messages", nil, nil)
		return
	}
	response.SuccessResponse(c, messages, nil)
}

func (h *ChatHandler) SendMessage(c *gin.Context) {
	userID := c.GetString("user_id")
	roomID := c.Param("id")

	var req dto.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body", nil, nil)
		return
	}

	msg, buyerUserID, supplierUserID, err := h.chatUC.SendMessage(c.Request.Context(), userID, roomID, req.Body)
	if err != nil {
		if err == usecase.ErrRoomNotFound {
			response.ErrorResponse(c, http.StatusNotFound, "ROOM_NOT_FOUND", err.Error(), nil, nil)
			return
		}
		if err == usecase.ErrRoomAccessDenied {
			response.ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", err.Error(), nil, nil)
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to send message", nil, nil)
		return
	}

	// Broadcast via WebSocket hub to active participants
	if buyerUserID != "" {
		DefaultChatHub.SendToUser(buyerUserID, "chat.message_received", msg)
	}
	if supplierUserID != "" {
		DefaultChatHub.SendToUser(supplierUserID, "chat.message_received", msg)
	}

	response.SuccessResponseCreated(c, msg, nil)
}

func (h *ChatHandler) MarkAsRead(c *gin.Context) {
	userID := c.GetString("user_id")
	roomID := c.Param("id")

	err := h.chatUC.MarkAsRead(c.Request.Context(), userID, roomID)
	if err != nil {
		if err == usecase.ErrRoomNotFound {
			response.ErrorResponse(c, http.StatusNotFound, "ROOM_NOT_FOUND", err.Error(), nil, nil)
			return
		}
		if err == usecase.ErrRoomAccessDenied {
			response.ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", err.Error(), nil, nil)
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to mark room as read", nil, nil)
		return
	}

	response.SuccessResponse(c, gin.H{"status": "ok"}, nil)
}

func (h *ChatHandler) ConnectWebSocket(c *gin.Context) {
	tokenString := c.Query("token")
	if tokenString == "" {
		cookie, err := c.Cookie("indosupplier_access_token")
		if err == nil {
			tokenString = cookie
		}
	}

	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token missing"})
		return
	}

	claims, err := h.jwtManager.ValidateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token invalid"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("failed to upgrade connection: %v", err)
		return
	}

	client := DefaultChatHub.Register(claims.UserID, conn)

	// Keep connection alive, listen for close events
	go func() {
		defer func() {
			DefaultChatHub.Unregister(client)
			_ = conn.Close()
		}()
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}()
}

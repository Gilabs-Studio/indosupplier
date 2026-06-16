package usecase

import (
	"context"
	"errors"

	"gorm.io/gorm"

	buyerModels "github.com/gilabs/indosupplier/api/internal/buyer/data/models"
	chatModels "github.com/gilabs/indosupplier/api/internal/chat/data/models"
	"github.com/gilabs/indosupplier/api/internal/chat/domain/dto"
	"github.com/gilabs/indosupplier/api/internal/core/apptime"
	supplierModels "github.com/gilabs/indosupplier/api/internal/supplier/data/models"
)

var (
	ErrBuyerProfileNotFound    = errors.New("buyer profile not found")
	ErrSupplierProfileNotFound = errors.New("supplier profile not found")
	ErrRoomNotFound            = errors.New("chat room not found")
	ErrRoomAccessDenied        = errors.New("access to chat room denied")
)

type ChatUsecase interface {
	ListRooms(ctx context.Context, userID string) ([]dto.RoomResponse, error)
	GetOrCreateRoom(ctx context.Context, userID string, req *dto.CreateRoomRequest) (*dto.RoomResponse, error)
	GetRoomMessages(ctx context.Context, userID string, roomID string) ([]dto.MessageResponse, error)
	SendMessage(ctx context.Context, userID string, roomID string, body string) (*dto.MessageResponse, string, string, error) // returns message, buyerUserID, supplierUserID, error
	MarkAsRead(ctx context.Context, userID string, roomID string) error
}

type chatUsecase struct {
	db *gorm.DB
}

func NewChatUsecase(db *gorm.DB) ChatUsecase {
	return &chatUsecase{db: db}
}

func (u *chatUsecase) getBuyerProfileID(ctx context.Context, userID string) (string, error) {
	var buyer buyerModels.BuyerProfile
	if err := u.db.WithContext(ctx).Where("user_id = ?", userID).First(&buyer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrBuyerProfileNotFound
		}
		return "", err
	}
	return buyer.ID, nil
}

func (u *chatUsecase) ListRooms(ctx context.Context, userID string) ([]dto.RoomResponse, error) {
	buyerID, err := u.getBuyerProfileID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var rooms []chatModels.ChatRoom
	// Fetch all rooms with preloaded SupplierProfile and LastMessage in a single batch query
	err = u.db.WithContext(ctx).
		Preload("SupplierProfile").
		Preload("LastMessage").
		Where("buyer_profile_id = ?", buyerID).
		Order("updated_at DESC").
		Find(&rooms).Error
	if err != nil {
		return nil, err
	}

	if len(rooms) == 0 {
		return []dto.RoomResponse{}, nil
	}

	// Gather all room IDs to batch fetch unread message counts
	var roomIDs []string
	for _, r := range rooms {
		roomIDs = append(roomIDs, r.ID)
	}

	// Batch query unread counts for all rooms in a single query
	type UnreadCount struct {
		ChatRoomID string
		Count      int
	}
	var counts []UnreadCount
	err = u.db.WithContext(ctx).
		Model(&chatModels.ChatMessage{}).
		Select("chat_room_id, count(*) as count").
		Where("chat_room_id IN ? AND sender_type = ? AND is_read = ?", roomIDs, "supplier", false).
		Group("chat_room_id").
		Scan(&counts).Error
	if err != nil {
		return nil, err
	}

	// Map unread counts for quick O(1) lookup
	unreadMap := make(map[string]int)
	for _, c := range counts {
		unreadMap[c.ChatRoomID] = c.Count
	}

	var responses []dto.RoomResponse
	for _, r := range rooms {
		var lastMsg *dto.MessageResponse
		if r.LastMessage != nil {
			lastMsg = &dto.MessageResponse{
				ID:         r.LastMessage.ID,
				ChatRoomID: r.LastMessage.ChatRoomID,
				SenderID:   r.LastMessage.SenderID,
				SenderType: r.LastMessage.SenderType,
				Body:       r.LastMessage.Body,
				IsRead:     r.LastMessage.IsRead,
				CreatedAt:  r.LastMessage.CreatedAt,
			}
		}

		compName := "Supplier"
		isVerified := false
		if r.SupplierProfile != nil {
			compName = r.SupplierProfile.CompanyName
			isVerified = r.SupplierProfile.IsPremiumVerified
		}

		responses = append(responses, dto.RoomResponse{
			ID:                r.ID,
			BuyerProfileID:    r.BuyerProfileID,
			SupplierProfileID: r.SupplierProfileID,
			CompanyName:       compName,
			IsVerified:        isVerified,
			LastMessage:       lastMsg,
			UnreadCount:       unreadMap[r.ID],
			CreatedAt:         r.CreatedAt,
			UpdatedAt:         r.UpdatedAt,
		})
	}

	return responses, nil
}

func (u *chatUsecase) GetOrCreateRoom(ctx context.Context, userID string, req *dto.CreateRoomRequest) (*dto.RoomResponse, error) {
	buyerID, err := u.getBuyerProfileID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Verify supplier exists
	var supplier supplierModels.SupplierProfile
	if err := u.db.WithContext(ctx).Where("id = ?", req.SupplierProfileID).First(&supplier).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSupplierProfileNotFound
		}
		return nil, err
	}

	var room chatModels.ChatRoom
	err = u.db.WithContext(ctx).
		Preload("SupplierProfile").
		Preload("LastMessage").
		Where("buyer_profile_id = ? AND supplier_profile_id = ?", buyerID, req.SupplierProfileID).
		First(&room).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new chat room
			room = chatModels.ChatRoom{
				BuyerProfileID:    buyerID,
				SupplierProfileID: req.SupplierProfileID,
			}
			if err := u.db.WithContext(ctx).Create(&room).Error; err != nil {
				return nil, err
			}

			// Reload to get associations correctly preloaded
			err = u.db.WithContext(ctx).
				Preload("SupplierProfile").
				Where("id = ?", room.ID).
				First(&room).Error
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return &dto.RoomResponse{
		ID:                room.ID,
		BuyerProfileID:    room.BuyerProfileID,
		SupplierProfileID: room.SupplierProfileID,
		CompanyName:       supplier.CompanyName,
		IsVerified:        supplier.IsPremiumVerified,
		CreatedAt:         room.CreatedAt,
		UpdatedAt:         room.UpdatedAt,
	}, nil
}

func (u *chatUsecase) GetRoomMessages(ctx context.Context, userID string, roomID string) ([]dto.MessageResponse, error) {
	buyerID, err := u.getBuyerProfileID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var room chatModels.ChatRoom
	if err := u.db.WithContext(ctx).Where("id = ?", roomID).First(&room).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoomNotFound
		}
		return nil, err
	}

	if room.BuyerProfileID != buyerID {
		return nil, ErrRoomAccessDenied
	}

	var messages []chatModels.ChatMessage
	err = u.db.WithContext(ctx).
		Where("chat_room_id = ?", roomID).
		Order("created_at ASC").
		Find(&messages).Error
	if err != nil {
		return nil, err
	}

	var responses []dto.MessageResponse
	for _, m := range messages {
		responses = append(responses, dto.MessageResponse{
			ID:         m.ID,
			ChatRoomID: m.ChatRoomID,
			SenderID:   m.SenderID,
			SenderType: m.SenderType,
			Body:       m.Body,
			IsRead:     m.IsRead,
			CreatedAt:  m.CreatedAt,
		})
	}

	return responses, nil
}

func (u *chatUsecase) SendMessage(ctx context.Context, userID string, roomID string, body string) (*dto.MessageResponse, string, string, error) {
	buyerID, err := u.getBuyerProfileID(ctx, userID)
	if err != nil {
		return nil, "", "", err
	}

	var room chatModels.ChatRoom
	if err := u.db.WithContext(ctx).
		Preload("SupplierProfile").
		Where("id = ?", roomID).
		First(&room).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", "", ErrRoomNotFound
		}
		return nil, "", "", err
	}

	if room.BuyerProfileID != buyerID {
		return nil, "", "", ErrRoomAccessDenied
	}

	var supplierUserID string
	if room.SupplierProfile != nil {
		supplierUserID = room.SupplierProfile.UserID
	}

	message := &chatModels.ChatMessage{
		ChatRoomID: roomID,
		SenderID:   userID,
		SenderType: "buyer",
		Body:       body,
		IsRead:     false,
	}

	// Transaction to create message and update last_message_id in room
	err = u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(message).Error; err != nil {
			return err
		}

		now := apptime.Now()
		if err := tx.Model(&chatModels.ChatRoom{}).
			Where("id = ?", roomID).
			Updates(map[string]interface{}{
				"last_message_id": message.ID,
				"updated_at":      now,
			}).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, "", "", err
	}

	resp := &dto.MessageResponse{
		ID:         message.ID,
		ChatRoomID: message.ChatRoomID,
		SenderID:   message.SenderID,
		SenderType: message.SenderType,
		Body:       message.Body,
		IsRead:     message.IsRead,
		CreatedAt:  message.CreatedAt,
	}

	return resp, userID, supplierUserID, nil
}

func (u *chatUsecase) MarkAsRead(ctx context.Context, userID string, roomID string) error {
	buyerID, err := u.getBuyerProfileID(ctx, userID)
	if err != nil {
		return err
	}

	var room chatModels.ChatRoom
	if err := u.db.WithContext(ctx).Where("id = ?", roomID).First(&room).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRoomNotFound
		}
		return err
	}

	if room.BuyerProfileID != buyerID {
		return ErrRoomAccessDenied
	}

	// Mark all messages from the supplier as read
	return u.db.WithContext(ctx).
		Model(&chatModels.ChatMessage{}).
		Where("chat_room_id = ? AND sender_type = ? AND is_read = ?", roomID, "supplier", false).
		Update("is_read", true).Error
}

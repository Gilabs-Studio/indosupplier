package dto

type BuyerNotificationResponse struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Desc        string  `json:"desc"`
	Date        string  `json:"date"`
	Unread      bool    `json:"unread"`
	Type        string  `json:"type"`
	RelatedType string  `json:"related_type,omitempty"`
	RelatedID   *string `json:"related_id,omitempty"`
}

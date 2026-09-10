package dto

type BuyerNotificationResponse struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Desc   string `json:"desc"`
	Date   string `json:"date"`
	Unread bool   `json:"unread"`
	Type   string `json:"type"`
}

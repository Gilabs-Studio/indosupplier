package dto

import "time"

type BatchSelectionItem struct {
	ID           string    `json:"id"`
	BatchNumber  string    `json:"batch_number"`
	Quantity     float64   `json:"quantity"`
	ExpiredAt    time.Time `json:"expired_at"`
	ReceivedAt   time.Time `json:"received_at"`
	SelectedQty  int       `json:"selected_qty"`
}

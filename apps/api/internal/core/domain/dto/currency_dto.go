package dto

import "time"

type CreateCurrencyRequest struct {
	Code          string `json:"code" binding:"required,min=2,max=10"`
	Name          string `json:"name" binding:"required,min=2,max=100"`
	Symbol        string `json:"symbol" binding:"max=10"`
	DecimalPlaces *int   `json:"decimal_places" binding:"omitempty,min=0,max=6"`
	IsActive      *bool  `json:"is_active"`
}

type UpdateCurrencyRequest struct {
	Code          string `json:"code" binding:"omitempty,min=2,max=10"`
	Name          string `json:"name" binding:"omitempty,min=2,max=100"`
	Symbol        string `json:"symbol" binding:"max=10"`
	DecimalPlaces *int   `json:"decimal_places" binding:"omitempty,min=0,max=6"`
	IsActive      *bool  `json:"is_active"`
}

type CurrencyResponse struct {
	ID            string    `json:"id"`
	Code          string    `json:"code"`
	Name          string    `json:"name"`
	Symbol        string    `json:"symbol"`
	DecimalPlaces int       `json:"decimal_places"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

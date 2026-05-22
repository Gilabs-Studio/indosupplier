package dto

import "time"

// AvailableAssetResponse represents an available asset from Finance Assets module
type AvailableAssetResponse struct {
	ID         string                      `json:"id"`
	Code       string                      `json:"code"`
	Name       string                      `json:"name"`
	Category   *AvailableAssetCategoryLite `json:"category,omitempty"`
	Location   *AvailableAssetLocationLite `json:"location,omitempty"`
	AssetImage string                      `json:"asset_image,omitempty"`
	Status     string                      `json:"status"`
	BookValue  float64                     `json:"book_value"`
}

type AvailableAssetCategoryLite struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AvailableAssetLocationLite struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CreateEmployeeAssetRequest struct {
	AssetID         string  `json:"asset_id" binding:"required,uuid"`
	AssetName       string  `json:"asset_name" binding:"required,max=200"`
	AssetCode       string  `json:"asset_code" binding:"required,max=100"`
	AssetCategory   string  `json:"asset_category" binding:"required,max=100"`
	BorrowDate      string  `json:"borrow_date" binding:"required"`
	BorrowCondition string  `json:"borrow_condition" binding:"required,oneof=NEW GOOD FAIR POOR DAMAGED"`
	AssetImage      string  `json:"asset_image" binding:"omitempty,max=255"`
	Notes           *string `json:"notes" binding:"omitempty"`
}

type UpdateEmployeeAssetRequest struct {
	AssetID         string  `json:"asset_id" binding:"omitempty,uuid"`
	AssetName       string  `json:"asset_name" binding:"omitempty,max=200"`
	AssetCode       string  `json:"asset_code" binding:"omitempty,max=100"`
	AssetCategory   string  `json:"asset_category" binding:"omitempty,max=100"`
	BorrowDate      string  `json:"borrow_date" binding:"omitempty"`
	BorrowCondition string  `json:"borrow_condition" binding:"omitempty,oneof=NEW GOOD FAIR POOR DAMAGED"`
	AssetImage      *string `json:"asset_image" binding:"omitempty,max=255"`
	Notes           *string `json:"notes" binding:"omitempty"`
}

type ReturnEmployeeAssetRequest struct {
	ReturnDate      string  `json:"return_date" binding:"required"`
	ReturnCondition string  `json:"return_condition" binding:"required,oneof=NEW GOOD FAIR POOR DAMAGED"`
	Notes           *string `json:"notes" binding:"omitempty"`
}

type EmployeeAssetResponse struct {
	ID              string                  `json:"id"`
	EmployeeID      string                  `json:"employee_id"`
	AssetID         *string                 `json:"asset_id,omitempty"`
	Asset           *AvailableAssetResponse `json:"asset,omitempty"`
	AssetName       string                  `json:"asset_name"`
	AssetCode       string                  `json:"asset_code"`
	AssetCategory   string                  `json:"asset_category"`
	BorrowDate      string                  `json:"borrow_date"`
	ReturnDate      *string                 `json:"return_date"`
	BorrowCondition string                  `json:"borrow_condition"`
	ReturnCondition *string                 `json:"return_condition"`
	AssetImage      string                  `json:"asset_image"`
	Notes           *string                 `json:"notes"`
	Status          string                  `json:"status"`
	DaysBorrowed    int                     `json:"days_borrowed"`
	CreatedAt       *time.Time              `json:"created_at,omitempty"`
	UpdatedAt       *time.Time              `json:"updated_at,omitempty"`
}

type EmployeeAssetBriefResponse struct {
	ID            string  `json:"id"`
	AssetName     string  `json:"asset_name"`
	AssetCode     string  `json:"asset_code"`
	AssetCategory string  `json:"asset_category"`
	BorrowDate    string  `json:"borrow_date"`
	ReturnDate    *string `json:"return_date"`
	AssetImage    string  `json:"asset_image"`
	Status        string  `json:"status"`
	DaysBorrowed  int     `json:"days_borrowed"`
}

package dto

type JoinWaitingListRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Name        string `json:"name" binding:"required,min=2,max=100"`
	CompanyName string `json:"company_name" binding:"required,min=2,max=100"`
	CompanyType string `json:"company_type" binding:"required,oneof=supplier buyer other"`
	Phone       string `json:"phone" binding:"omitempty,max=20"`
	Notes       string `json:"notes" binding:"omitempty,max=1000"`
}

type UpdateWaitingListStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=pending approved contacted rejected"`
}

type WaitingListResponse struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	CompanyName string `json:"company_name"`
	CompanyType string `json:"company_type"`
	Phone       string `json:"phone"`
	Notes       string `json:"notes"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

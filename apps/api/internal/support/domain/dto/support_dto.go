package dto

type SupportTicketListItem struct {
	ID      string `json:"id"`
	Subject string `json:"subject"`
	Date    string `json:"date"`
	Status  string `json:"status"`
}

type SupportMessage struct {
	Sender string `json:"sender"`
	Name   string `json:"name"`
	Text   string `json:"text"`
	Time   string `json:"time"`
}

type SupportTicketDetail struct {
	ID       string           `json:"id"`
	Subject  string           `json:"subject"`
	Date     string           `json:"date"`
	Status   string           `json:"status"`
	Messages []SupportMessage `json:"messages"`
}

type CreateSupportTicketRequest struct {
	Subject string `json:"subject" binding:"required"`
	Message string `json:"message" binding:"required"`
}

type CreateSupportReplyRequest struct {
	Text string `json:"text" binding:"required"`
}

package dto

type FinanceSettingResponse struct {
	ID          string `json:"id"`
	SettingKey  string `json:"setting_key"`
	Value       string `json:"value"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

type UpsertFinanceSettingRequest struct {
	SettingKey  string `json:"setting_key"`
	Value       string `json:"value"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

type BatchUpsertFinanceSettingsRequest struct {
	Settings []UpsertFinanceSettingRequest `json:"settings"`
}

type UpdateAgingBucketConfigRequest struct {
	ReportType string                  `json:"report_type" binding:"required,oneof=ar ap"`
	Buckets    []AgingBucketDefinition `json:"buckets" binding:"required,min=1"`
}

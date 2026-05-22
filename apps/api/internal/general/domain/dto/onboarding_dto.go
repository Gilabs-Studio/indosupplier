package dto

// OnboardingStepsResponse reflects per-module completion status derived from real data.
type OnboardingStepsResponse struct {
	Company     bool `json:"company"`
	Outlet      bool `json:"outlet"`
	FloorLayout bool `json:"floor_layout"`
	Products    bool `json:"products"`
	Warehouse   bool `json:"warehouse"`
	Users       bool `json:"users"`
	FiscalYear  bool `json:"fiscal_year"`
}

// OnboardingStateResponse is returned by GET /general/onboarding.
type OnboardingStateResponse struct {
	BusinessType string                   `json:"business_type"`
	Completed    bool                     `json:"completed"`
	Steps        *OnboardingStepsResponse `json:"steps,omitempty"`
}

// SetBusinessTypeRequest is used by PUT /general/onboarding/business-type.
type SetBusinessTypeRequest struct {
	BusinessType string `json:"business_type" binding:"required,oneof=fnb retail service other"`
}

// CompleteOnboardingRequest is used by PUT /general/onboarding/complete.
type CompleteOnboardingRequest struct {
	Completed bool `json:"completed"`
}

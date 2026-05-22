package dto

// SystemAdminLoginRequest is the request body for system admin login
type SystemAdminLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// SystemAdminUpdateProfileRequest updates system admin username/email.
// Username maps to the existing name column in system_admins.
type SystemAdminUpdateProfileRequest struct {
	Username string `json:"username" binding:"required,min=3,max=255"`
	Email    string `json:"email" binding:"required,email,max=255"`
}

// SystemAdminChangePasswordRequest updates system admin password.
type SystemAdminChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required,min=8,max=72"`
	NewPassword     string `json:"new_password" binding:"required,min=8,max=72"`
}

// SystemAdminLoginResponse is returned after successful system admin login
type SystemAdminLoginResponse struct {
	Admin        *SystemAdminResponse `json:"admin"`
	AccessToken  string               `json:"access_token,omitempty"`
	RefreshToken string               `json:"refresh_token,omitempty"`
	ExpiresIn    int                  `json:"expires_in"`
}

// SystemAdminResponse is the safe representation of a system admin
type SystemAdminResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Status   string `json:"status"`
}

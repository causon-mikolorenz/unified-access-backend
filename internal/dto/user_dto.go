package dto

// UserRequest handles the incoming data for creating or updating a user.
type UserRequest struct {
	Username   string   `json:"user_name" binding:"required,min=4"`
	FirstName  string   `json:"first_name" binding:"required"`
	MiddleName string   `json:"middle_name"`
	LastName   string   `json:"last_name" binding:"required"`
	Email      string   `json:"email" binding:"required,email"`
	Password   string   `json:"password" binding:"required,min=8"`
	Status     string   `json:"status" binding:"required"`
	Roles      []string `json:"roles" binding:"required,gt=0"`
}

// UpdatePasswordRequest handles incoming patch data for updating password
type UpdatePasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required"`
}

//UpdateStatusRequest handles patch data for updating user status
type UpdateStatusRequest struct {
	NewStatus string `json:"new_status" binding:"required"`
}

//UpdateUserRoleRequest handles patch data for updating user roles
type UpdateUserRoleRequest struct {
	RoleIDs []int `json:"role_ids" binding:"required" validate:"required"`
}

// UserResponse provides a safe view of user data, hiding the password hash.
type UserResponse struct {
	ID         string             `json:"id"`
	Username   string             `json:"user_name"`
	FirstName  string             `json:"first_name"`
	MiddleName string             `json:"middle_name"`
	LastName   string             `json:"last_name"`
	Email      string             `json:"email"`
	Status     string             `json:"status"`
	CreatedAt  string             `json:"created_at"`
	UpdatedAt  string             `json:"updated_at"`
	Roles      []UserRoleRepsonse `json:"roles"`
}

// UserResponseList follows the same pagination pattern as Roles and Clients.
type UserResponseList struct {
	Users       []UserResponse `json:"users"`
	TotalCount  int            `json:"total_count"`
	CurrentPage int            `json:"current_page"`
	LastPage    int            `json:"last_page"`
}

// UserStatusUpdate handles administrative status changes (Active/Inactive).
type UserStatusUpdate struct {
	Status string `json:"status" binding:"required,oneof=active inactive"`
}

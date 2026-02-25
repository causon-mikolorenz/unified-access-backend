package dto

type RoleRequest struct {
	RoleName    string `json:"role_name"`
	Description string `json:"description"`
}

type RoleResponse struct {
	ID          int    `json:"id"`
	RoleName    string `json:"role_name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type UserRoleRepsonse struct {
	ID          int    `json:"id"`
	RoleName    string `json:"role_name"`
	Description string `json:"description"`
}

type RoleListResponse struct {
	Roles       []RoleResponse `json:"roles"`
	TotalCount  int            `json:"total_count"`
	CurrentPage int            `json:"current_page"`
	LastPage    int            `json:"last_page"`
}

package dto

type CreateClientRequest struct {
	Name         string   `json:"name" binding:"required" example:"LMS Portal"`
	Abbreviation string   `json:"abbreviation" binding:"required,max=10" example:"lms"`
	Description  string   `json:"description" example:"Learning Management System"`
	BaseURL      string   `json:"base_url" binding:"required,url" example:"http://localhost:3000"`
	RedirectURI  string   `json:"redirect_uri" binding:"required,url" example:"http://localhost:3000/callback"`
	LogoutURI    string   `json:"logout_uri" binding:"required,url" example:"http://localhost:3000/logout"`
	Grants       []string `json:"grants" binding:"required,dive,oneof=authorization_code refresh_token client_credentials"`
	Roles        []string `json:"roles" example:"['admin', 'student']"`
}

type UpdateClientRequest struct {
	Name          string `json:"name" binding:"required"`
	Description   string `json:"description"`
	BaseURL       string `json:"base_url" binding:"required,url"`
	RedirectURI   string `json:"redirect_uri" binding:"required,url"`
	LogoutURI     string `json:"logout_uri" binding:"required,url"`
	ImageLocation string `json:"image_location"`
}

type ClientResponse struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Abbreviation  string `json:"abbreviation"`
	Description   string `json:"description"`
	ImageLocation string `json:"image_location"`
	BaseURL       string `json:"base_url"`
	RedirectURI   string `json:"redirect_uri"`
	LogoutURI     string `json:"logout_uri"`
	CreatedAt     string `json:"created_at"`
}
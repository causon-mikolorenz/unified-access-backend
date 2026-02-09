type CreateClientRequest struct {
    Name          string   `json:"name" binding:"required"`
    Abbreviation  string   `json:"abbreviation" binding:"required,max=10"`
    Description   string   `json:"description"`
    BaseURL       string   `json:"base_url" binding:"required,url"`
    RedirectURI   string   `json:"redirect_uri" binding:"required,url"`
    LogoutURI     string   `json:"logout_uri" binding:"required,url"`
    AllowedGrants []string `json:"allowed_grants" binding:"required,dive,oneof=authorization_code refresh_token client_credentials"`
    Roles         []string `json:"roles"`
}

type ClientResponse struct {
    ID            string   `json:"id"`
    Name          string   `json:"name"`
    Abbreviation  string   `json:"abbreviation"`
    Description   string   `json:"description"`
    ImageLocation string   `json:"image_location"`
    BaseURL       string   `json:"base_url"`
    RedirectURI   string   `json:"redirect_uri"`
    LogoutURI     string   `json:"logout_uri"`
    CreatedAt     string   `json:"created_at"`
}
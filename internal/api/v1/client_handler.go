package v1

import (
	"crypto/rsa"
	"log"
	"net/http"
	"strconv"

	"github.com/causon-mikolorenz/unified-access-backend/internal/auth"
	"github.com/causon-mikolorenz/unified-access-backend/internal/dto"
	"github.com/causon-mikolorenz/unified-access-backend/internal/repository"
	"github.com/causon-mikolorenz/unified-access-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ClientHandler struct {
	Repo       *repository.ClientRepository
	PrivateKey *rsa.PrivateKey
}

// PostClient handles POST /v1/admin/clients
// @Summary Register a new Service Provider with Icon
// @Description Creates client, saves icon, hashes secret, and maps roles
// @Tags ServiceProviders
// @Accept multipart/form-data
// @Produce json
// @Param name formData string true "Client Name"
// @Param abbreviation formData string true "Abbreviation"
// @Param description formData string false "Description"
// @Param base_url formData string true "Base URL"
// @Param redirect_uri formData string true "Redirect URI"
// @Param logout_uri formData string true "Logout URI"
// @Param grants formData []string true "Grants (e.g. authorization_code)"
// @Param roles formData []string false "Initial Roles"
// @Param image formData file true "Client Icon"
// @Success 201 {object} map[string]string
// @Router /v1/admin/clients [post]
func (h *ClientHandler) PostClient(c *gin.Context) {
	file, err := c.FormFile("image")
	var imagePath string
	if err == nil {
		f, _ := file.Open()
		defer f.Close()

		header := make([]byte, 512)
		f.Read(header)

		if err := auth.ValidateImage(header, file.Filename); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		abbr := c.PostForm("abbreviation")
		imagePath = "public/icons/" + abbr + "-" + file.Filename

		if err := c.SaveUploadedFile(file, imagePath); err != nil {
			log.Printf("[PostClient] What failed: image save %v", err)
			c.JSON(http.StatusInternalServerError,
				gin.H{"error": "failed to save image"})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client icon required"})
		return
	}

	clientID := uuid.New()
	rawSecret, _ := auth.GenerateRandomString(32)
	hashedSecret, _ := auth.HashSecret(rawSecret)

	clientModel := &models.Client{
		ID:            clientID[:],
		ClientName:    c.PostForm("name"),
		Abbreviation:  c.PostForm("abbreviation"),
		ClientSecret:  hashedSecret,
		BaseUrl:       c.PostForm("base_url"),
		RedirectUri:   c.PostForm("redirect_uri"),
		LogoutUri:     c.PostForm("logout_uri"),
		Description:   c.PostForm("description"),
		ImageLocation: imagePath,
	}

	grants := c.PostFormArray("grants")
	if err = h.Repo.CreateClient(clientModel, grants); err != nil {
		log.Printf("[PostClient] What failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"client_id":     clientID.String(),
		"client_secret": rawSecret,
		"image_url":     imagePath,
		"message":       "Copy secret now. It is stored as a hash.",
	})
}

// GetClientList handles GET /v1/admin/clients
// @Summary List Service Providers
// @Description Fetch active clients with pagination
// @Tags ServiceProviders
// @Param limit query int false "Pagination Limit" default(10)
// @Param offset query int false "Pagination Offset" default(0)
// @Success 200 {array} dto.ClientResponse
// @Router /v1/admin/clients [get]
func (h *ClientHandler) GetClientList(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	clients, err := h.Repo.ListClients(limit, offset)
	if err != nil {
		log.Printf("[GetClientList] What failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "fetch error"})
		return
	}

	var res []dto.ClientResponse
	for _, cl := range clients {
		id, _ := uuid.FromBytes(cl.ID)
		res = append(res, dto.ClientResponse{
			ID:            id.String(),
			Name:          cl.ClientName,
			Abbreviation:  cl.Abbreviation,
			Description:   cl.Description,
			ImageLocation: cl.ImageLocation,
		})
	}
	c.JSON(http.StatusOK, res)
}

// GetClient handles GET /v1/admin/clients/:id
// @Summary Get Client Details
// @Description Fetch full details including grants and roles
// @Tags ServiceProviders
// @Param id path string true "Client UUID"
// @Success 200 {object} dto.ClientResponse
// @Router /v1/admin/clients/{id} [get]
func (h *ClientHandler) GetClient(c *gin.Context) {
	idParam := c.Param("id")
	clientUUID, err := uuid.Parse(idParam)
	if err != nil {
		log.Printf("[GetClient] What failed: invalid uuid %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid uuid format"})
		return
	}

	cl, err := h.Repo.GetByID(clientUUID[:])
	if err != nil {
		log.Printf("[GetClient] What failed: client not found %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "client not found"})
		return
	}

	grants, _ := h.Repo.GetGrantTypes(cl.ID)
	roles, _ := h.Repo.GetClientRoles(cl.Abbreviation)

	id, _ := uuid.FromBytes(cl.ID)
	c.JSON(http.StatusOK, gin.H{
		"client": dto.ClientResponse{
			ID:            id.String(),
			Name:          cl.ClientName,
			Abbreviation:  cl.Abbreviation,
			Description:   cl.Description,
			ImageLocation: cl.ImageLocation,
			BaseURL:       cl.BaseUrl,
			RedirectURI:   cl.RedirectUri,
			LogoutURI:     cl.LogoutUri,
		},
		"allowed_grants": grants,
		"roles":          roles,
	})
}

// PutClient handles PUT /v1/admin/clients/:id
// @Summary Update Client Info
// @Description Update safe fields (Name, Description, URLs, Image)
// @Tags ServiceProviders
// @Accept json
// @Produce json
// @Param id path string true "Client UUID"
// @Param body body dto.UpdateClientRequest true "Updated Client Data"
// @Success 200 {object} map[string]string
// @Router /v1/admin/clients/{id} [put]
func (h *ClientHandler) PutClient(c *gin.Context) {
	idParam := c.Param("id")
	clientUUID, _ := uuid.Parse(idParam)

	var req dto.UpdateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[PutClient] What failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	client := &models.Client{
		ID:            clientUUID[:],
		ClientName:    req.Name,
		Description:   req.Description,
		ImageLocation: req.ImageLocation,
		BaseUrl:       req.BaseURL,
		RedirectUri:   req.RedirectURI,
		LogoutUri:     req.LogoutURI,
	}

	if err := h.Repo.UpdateClient(client); err != nil {
		log.Printf("[PutClient] What failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update fail"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "client updated"})
}

// DeleteClient handles DELETE /v1/admin/clients/:id
// @Summary Soft Delete Client
// @Tags ServiceProviders
// @Param id path string true "Client UUID"
// @Router /v1/admin/clients/{id} [delete]
func (h *ClientHandler) DeleteClient(c *gin.Context) {
	idParam := c.Param("id")
	clientUUID, err := uuid.Parse(idParam)
	if err != nil {
		log.Printf("[DeleteClient] What failed: invalid uuid %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid uuid"})
		return
	}

	if err := h.Repo.SoftDelete(clientUUID[:]); err != nil {
		log.Printf("[DeleteClient] What failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete fail"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "client deactivated successfully"})
}
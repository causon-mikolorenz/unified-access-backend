package v1

import (
	"fmt"
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

type UserHandler struct {
	Repo *repository.UserRepository
}

// PostUser creates a new user in the system
// @Summary Create User
// @Description Register a new user with roles and encrypted password
// @Tags Users
// @Accept json
// @Produce json
// @Param user body dto.UserRequest true "User Data"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/users [post]
func (h *UserHandler) PostUser(c *gin.Context) {
	var req dto.UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[PostUser] Bind JSON Error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := uuid.New()
	passwordHash, _ := auth.HashSecret(req.Password)

	user := models.User{
		ID:           userID[:],
		Username:     req.Username,
		FirstName:    req.FirstName,
		MiddleName:   req.MiddleName,
		LastName:     req.LastName,
		Email:        req.Email,
		PasswordHash: passwordHash,
		Status:       models.StatusActive,
		RoleString:   req.Roles,
	}

	err := h.Repo.CreateUser(&user)
	if err != nil {
		log.Printf("[PostUser] Database Create Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	message := fmt.Sprintf("Created user with the id %s", userID)
	c.JSON(http.StatusCreated, gin.H{"message": message})
}

// GetUser retrieves a single user by their UUID string
// @Summary Get User
// @Description Fetch user details using their unique ID
// @Tags Users
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} map[string]string
// @Failure 440 {object} map[string]string
// @Router /api/v1/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")
	userID, err := uuid.Parse(id)
	if err != nil {
		log.Printf("[GetUser] UUID Parse Error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID Format"})
		return
	}

	user, err := h.Repo.GetUserById(userID[:])
	if err != nil {
		log.Printf("[GetUser] Database Query Error: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, dto.UserResponse{
		ID:         id,
		Username:   user.Username,
		FirstName:  user.FirstName,
		MiddleName: user.MiddleName,
		LastName:   user.LastName,
		Email:      user.Email,
		Status:     string(user.Status),
		CreatedAt:  user.CreatedAt.Format(TIME_LAYOUT),
		UpdatedAt:  user.UpdatedAt.Format(TIME_LAYOUT),
	})
}

// GetUserList retrieves a paginated list of users
// @Summary List Users
// @Description Get a paginated list of all users
// @Tags Users
// @Produce json
// @Param page query int false "Page number" default(1)
// @Success 200 {object} dto.UserResponseList
// @Failure 500 {object} map[string]string
// @Router /api/v1/users [get]
func (h *UserHandler) GetUserList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * PAGE_LIMIT
	users, err := h.Repo.GetUserList(PAGE_LIMIT, offset)
	if err != nil {
		log.Printf("[GetUserList] Fetch List Error: %v", err)
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "Failed to fetch user list"})
		return
	}

	total, err := h.Repo.CountUsers()
	if err != nil {
		log.Printf("[GetUserList] Database Count Error: %v", err)
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "Failed to fetch user count"})
		return
	}

	lastPage := (total + PAGE_LIMIT - 1) / PAGE_LIMIT
	if lastPage == 0 {
		lastPage = 1
	}

	var userResponses []dto.UserResponse
    for _, user := range users {
        userId, err := uuid.FromBytes(user.ID[:])
        if err != nil {
            log.Printf("[GetUserLIst] Parsing failed: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "parse fail"})
        }
        userResponses = append(userResponses, dto.UserResponse{
            ID:         userId.String(),
            Username:   user.Username,
            FirstName:  user.FirstName,
            MiddleName: user.MiddleName,
            LastName:   user.LastName,
            Email:      user.Email,
            Status:     string(user.Status),
            CreatedAt:  user.CreatedAt.Format(TIME_LAYOUT),
            UpdatedAt:  user.UpdatedAt.Format(TIME_LAYOUT),
        })
    }

	c.JSON(http.StatusOK, dto.UserResponseList{
		Users:       userResponses,
		TotalCount:  total,
		CurrentPage: page,
		LastPage:    lastPage,
	})
}

// PatchUserPassword updates a user's password.
// @Summary Update user password
// @Description Updates the password for a specific user identified by ID.
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body dto.UpdateUserRequest true "Password Update Data"
// @Success 200 {object} map[string]string "OK"
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 501 {object} map[string]string "Internal Server Error"
// @Router /api/v1/users/{id}/password [patch]
func (h *UserHandler) PatchUserPassword(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdatePasswordRequest
	userId, err := uuid.Parse(id)
	if err != nil {
		log.Printf("[PatchUserPassword] UUID Parse Error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID Format"})
		return
	}

	passwordHash, err := auth.HashSecret(req.NewPassword)
	if err != nil {
		log.Printf("[PatchUpdatePassword] Hashing failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Hashing failed"})
		return
	}

	user := models.User{
		ID:           userId[:],
		PasswordHash: passwordHash,
	}

	err = h.Repo.UpdateUserPassword(&user)
	if err != nil {
		log.Printf("[PatchUpdatePassword] Update failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password Updated Successfuly!"})
}

// PatchUserStatus updates the operational status of a user.
// @Summary Update user status
// @Description Modifies the status (e.g., active, disabled) of a user by ID.
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body dto.UpdateUserRequest true "Status Update Data"
// @Success 200 {object} map[string]interface{} "OK"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 501 {object} map[string]interface{} "Internal Server Error"
// @Router /api/v1/users/{id}/status [patch]
func (h *UserHandler) PatchUserStatus(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateStatusRequest
	userId, err := uuid.Parse(id)
	if err != nil {
		log.Printf("[PatchUserStatus] UUID Parse Error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID Format"})
		return
	}

	status, err := models.MapStatus(req.NewStatus)
	if err != nil {
		log.Printf("[PatchUserStatus] Invalid status: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status request"})
		return
	}

	user := models.User{
		ID:     userId[:],
		Status: status,
	}

	err = h.Repo.UpdateStatus(&user)
	if err != nil {
		log.Printf("[PatchUserStatus] Update failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status Updated Successfuly!"})
}

func (h *UserHandler) PatchUserRoles(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateUserRoleRequest
	userId, err := uuid.Parse(id)
	if err != nil {
		log.Printf("[PatchUserRoles] UUID Parse Error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID Format"})
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[PatchUserRoles] Bind JSON Error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalaid input"})
		return
	}

	err = h.Repo.UpdateUserRoles(userId[:], req.RoleIDs)
	if err != nil {
		log.Printf("[PatchUserRoles] Update failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "roles updated successfully!"})
}

// DeleteUser performs a soft delete on a user record
// @Summary Delete User
// @Description Mark a user as deleted by ID
// @Tags Users
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	userID, err := uuid.Parse(id)
	if err != nil {
		log.Printf("[DeleteUser] UUID Parse Error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.Repo.SoftDelete(userID[:]); err != nil {
		log.Printf("[DeleteUser] Database SoftDelete Error: %v", err)
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "Deletion Failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

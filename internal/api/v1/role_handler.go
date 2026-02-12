package v1

import (
	"log"
	"net/http"
	"strconv"

	"github.com/causon-mikolorenz/unified-access-backend/internal/dto"
	"github.com/causon-mikolorenz/unified-access-backend/internal/repository"
	"github.com/causon-mikolorenz/unified-access-backend/models"
	"github.com/gin-gonic/gin"
)

type RoleHandler struct {
	Repo *repository.RoleRepository
}

const (
	ROLE_PAGE_LIMIT = 10
	TIME_LAYOUT     = "2006-01-02 15:04:05"
)

// PostRole handles POST /v1/admin/roles
// @Summary Create a new role
// @Description Adds a new global or SP-prefixed role to the system
// @Tags Roles
// @Accept json
// @Produce json
// @Param body body dto.RoleRequest true "Role details"
// @Success 201 {object} map[string]string "Role created successfully"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 500 {object} map[string]string "Database error"
// @Router /v1/admin/roles [post]
func (h *RoleHandler) PostRole(c *gin.Context) {
	var req dto.RoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role := models.Role{
		RoleName:    req.RoleName,
		Description: req.Description,
	}

	if err := h.Repo.CreateRole(role); err != nil {
		log.Printf("[PostRole] DB Error: %v", err)
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "Failed to create role"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Role created successfully"})
}

// GetRoleList handles GET /v1/admin/roles
// @Summary List all roles
// @Description Retrieves a paginated list of non-deleted roles
// @Tags Roles
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Success 200 {object} dto.RoleListRequest
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /v1/admin/roles [get]
func (h *RoleHandler) GetRoleList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * ROLE_PAGE_LIMIT

	roles, err := h.Repo.ListRoles(ROLE_PAGE_LIMIT, offset)
	if err != nil {
		log.Printf("[GetRoleList] Fetch failed: %v", err)
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "Failed to fetch roles"})
		return
	}

	total, err := h.Repo.CountRoles()
	if err != nil {
		log.Printf("[GetRoleList] Count failed: %v", err)
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "Failed to determine total pages"})
		return
	}

	lastPage := (total + ROLE_PAGE_LIMIT - 1) / ROLE_PAGE_LIMIT
	if lastPage == 0 {
		lastPage = 1
	}

	var roleResponses []dto.RoleResponse
	for _, r := range roles {
		roleResponses = append(roleResponses, dto.RoleResponse{
			ID:          r.ID,
			RoleName:    r.RoleName,
			Description: r.Description,
			CreatedAt:   r.CreatedAt.Format(TIME_LAYOUT),
			UpdatedAt:   r.UpdatedAt.Format(TIME_LAYOUT),
		})
	}

	c.JSON(http.StatusOK, dto.RoleListRequest{
		Roles:       roleResponses,
		CurrentPage: page,
		LastPage:    lastPage,
	})
}

// GetRole handles GET /v1/admin/roles/:id
// @Summary Get role by ID
// @Description Fetches full details of a specific role
// @Tags Roles
// @Accept json
// @Produce json
// @Param id path int true "Role ID"
// @Success 200 {object} dto.RoleResponse
// @Failure 404 {object} map[string]string "Role not found"
// @Router /v1/admin/roles/{id} [get]
func (h *RoleHandler) GetRole(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	role, err := h.Repo.GetByID(id)
	if err != nil {
		log.Printf("[GetRole] Not found ID %d: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
		return
	}

	c.JSON(http.StatusOK, dto.RoleResponse{
		ID:          role.ID,
		RoleName:    role.RoleName,
		Description: role.Description,
		CreatedAt:   role.CreatedAt.Format(TIME_LAYOUT),
		UpdatedAt:   role.UpdatedAt.Format(TIME_LAYOUT),
	})
}

// PutRole handles PUT /v1/admin/roles/:id
// @Summary Update an existing role
// @Description Modifies role name or description
// @Tags Roles
// @Accept json
// @Produce json
// @Param id path int true "Role ID"
// @Param body body dto.RoleRequest true "Updated role data"
// @Success 200 {object} map[string]string "Role updated"
// @Failure 400 {object} map[string]string "Invalid input"
// @Router /v1/admin/roles/{id} [put]
func (h *RoleHandler) PutRole(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var req dto.RoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role := models.Role{
		ID:          id,
		RoleName:    req.RoleName,
		Description: req.Description,
	}

	if err := h.Repo.UpdateRole(role); err != nil {
		log.Printf("[PutRole] Update failed for ID %d: %v", id, err)
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "Update failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role updated"})
}

// DeleteRole handles DELETE /v1/admin/roles/:id
// @Summary Soft delete a role
// @Description Marks a role as deleted in the audit trail
// @Tags Roles
// @Param id path int true "Role ID"
// @Success 200 {object} map[string]string "Role deleted"
// @Router /v1/admin/roles/{id} [delete]
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.Repo.SoftDelete(id); err != nil {
		log.Printf("[DeleteRole] Deletion failed for ID %d: %v", id, err)
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "Deletion failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role deleted successfully"})
}

package building_admin_controllers

import (
	"strconv"
	"time"

	databases "github.com/The-Healthist/iboard_http_service/database"
	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	building_admin_services "github.com/The-Healthist/iboard_http_service/services/building_admin"
	"github.com/The-Healthist/iboard_http_service/services/container"
	"github.com/The-Healthist/iboard_http_service/utils"
	"github.com/The-Healthist/iboard_http_service/utils/field"
	"github.com/The-Healthist/iboard_http_service/utils/response"
	"github.com/gin-gonic/gin"
)

type BuildingAdminAdvertisementController struct {
	Ctx       *gin.Context
	Container *container.ServiceContainer
}

func NewBuildingAdminAdvertisementController(
	ctx *gin.Context,
	container *container.ServiceContainer,
) *BuildingAdminAdvertisementController {
	return &BuildingAdminAdvertisementController{
		Ctx:       ctx,
		Container: container,
	}
}

// HandleFuncBuildingAdminAdvertisement returns a gin.HandlerFunc for the specified method
func HandleFuncBuildingAdminAdvertisement(container *container.ServiceContainer, method string) gin.HandlerFunc {
	switch method {
	case "getAdvertisements":
		return func(ctx *gin.Context) {
			controller := NewBuildingAdminAdvertisementController(ctx, container)
			controller.GetAdvertisements()
		}
	case "getAdvertisement":
		return func(ctx *gin.Context) {
			controller := NewBuildingAdminAdvertisementController(ctx, container)
			controller.GetAdvertisement()
		}
	case "createAdvertisement":
		return func(ctx *gin.Context) {
			controller := NewBuildingAdminAdvertisementController(ctx, container)
			controller.CreateAdvertisement()
		}
	case "updateAdvertisement":
		return func(ctx *gin.Context) {
			controller := NewBuildingAdminAdvertisementController(ctx, container)
			controller.UpdateAdvertisement()
		}
	case "deleteAdvertisement":
		return func(ctx *gin.Context) {
			controller := NewBuildingAdminAdvertisementController(ctx, container)
			controller.DeleteAdvertisement()
		}
	default:
		return func(ctx *gin.Context) {
			ctx.JSON(400, gin.H{"error": "invalid method"})
		}
	}
}

func (c *BuildingAdminAdvertisementController) GetAdvertisements() {
	email := c.Ctx.GetString("email")
	if email == "" {
		c.Ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	var searchQuery struct {
		Search string `form:"search"`
		Type   string `form:"type"`
	}
	if err := c.Ctx.ShouldBindQuery(&searchQuery); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	pagination := struct {
		PageSize int  `form:"pageSize"`
		PageNum  int  `form:"pageNum"`
		Desc     bool `form:"desc"`
	}{
		PageSize: 10,
		PageNum:  1,
		Desc:     false,
	}

	if err := c.Ctx.ShouldBindQuery(&pagination); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	queryMap := utils.StructToMap(searchQuery)
	paginationMap := map[string]interface{}{
		"pageSize": pagination.PageSize,
		"pageNum":  pagination.PageNum,
		"desc":     pagination.Desc,
	}

	advertisements, paginationResult, err := c.Container.GetService("buildingAdminAdvertisement").(building_admin_services.InterfaceBuildingAdminAdvertisementService).Get(email, queryMap, paginationMap)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"data":       advertisements,
		"pagination": paginationResult,
	})
}

func (c *BuildingAdminAdvertisementController) GetAdvertisement() {
	email := c.Ctx.GetString("email")
	if email == "" {
		c.Ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "invalid advertisement ID"})
		return
	}

	advertisement, err := c.Container.GetService("buildingAdminAdvertisement").(building_admin_services.InterfaceBuildingAdminAdvertisementService).GetByID(uint(id), email)
	if err != nil {
		c.Ctx.JSON(404, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"data": advertisement})
}

type CreateAdvertisementRequest struct {
	Title       string                     `json:"title" binding:"required"`
	Description string                     `json:"description"`
	Type        field.AdvertisementType    `json:"type"`
	Status      field.Status               `json:"status"`
	Duration    int                        `json:"duration"`
	StartTime   *time.Time                 `json:"startTime"`
	EndTime     *time.Time                 `json:"endTime"`
	Display     field.AdvertisementDisplay `json:"display"`
	IsPublic    bool                       `json:"isPublic"`
	Path        string                     `json:"path"`
}

func (c *BuildingAdminAdvertisementController) CreateAdvertisement() {
	email := c.Ctx.GetString("email")
	if email == "" {
		c.Ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	var req CreateAdvertisementRequest
	if err := c.Ctx.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c.Ctx, err)
		return
	}

	// Set default values for startTime, endTime, and status
	now := time.Now()
	startTime := now
	endTime := now.AddDate(1, 0, 0) // 1 year from now
	status := field.Status("active")

	if req.StartTime != nil {
		startTime = *req.StartTime
	}
	if req.EndTime != nil {
		endTime = *req.EndTime
	}
	if req.Status != "" {
		status = req.Status
	}

	// If path is provided, find the corresponding file
	var fileID *uint
	if req.Path != "" {
		var file base_models.File
		if err := databases.DB_CONN.Where("path = ?", req.Path).First(&file).Error; err != nil {
			c.Ctx.JSON(400, gin.H{
				"error":   err.Error(),
				"message": "file not found",
			})
			return
		}
		fileID = &file.ID
	}

	advertisement := &base_models.Advertisement{
		Title:       req.Title,
		Description: req.Description,
		Type:        req.Type,
		Status:      status,
		Duration:    req.Duration,
		StartTime:   startTime,
		EndTime:     endTime,
		Display:     req.Display,
		FileID:      fileID,
		IsPublic:    false, // Force set to false
	}

	if err := c.Container.GetService("buildingAdminAdvertisement").(building_admin_services.InterfaceBuildingAdminAdvertisementService).Create(advertisement, email); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "create advertisement failed",
		})
		return
	}

	// Reload advertisement to get associated file information
	if err := databases.DB_CONN.Preload("File").First(advertisement, advertisement.ID).Error; err != nil {
		c.Ctx.JSON(200, gin.H{
			"message": "create advertisement success, but failed to load file info",
			"data":    advertisement,
		})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "create advertisement success",
		"data":    advertisement,
	})
}

func (c *BuildingAdminAdvertisementController) UpdateAdvertisement() {
	email := c.Ctx.GetString("email")
	if email == "" {
		c.Ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	var form struct {
		ID          uint                    `json:"id" binding:"required"`
		Title       string                  `json:"title"`
		Description string                  `json:"description"`
		Type        field.AdvertisementType `json:"type"`
		Status      field.Status            `json:"status"`
		StartTime   *time.Time              `json:"startTime"`
		EndTime     *time.Time              `json:"endTime"`
		IsPublic    *bool                   `json:"isPublic"`
		Path        string                  `json:"path"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if form.Title != "" {
		updates["title"] = form.Title
	}
	if form.Description != "" {
		updates["description"] = form.Description
	}
	if form.Type != "" {
		updates["type"] = form.Type
	}
	if form.Status != "" {
		updates["status"] = form.Status
	}
	if form.StartTime != nil {
		updates["start_time"] = form.StartTime
	}
	if form.EndTime != nil {
		updates["end_time"] = form.EndTime
	}
	updates["is_public"] = false // Force set to false

	// If new path is provided
	if form.Path != "" {
		// Start transaction
		tx := databases.DB_CONN.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		// Find new file
		var newFile base_models.File
		if err := tx.Where("path = ?", form.Path).First(&newFile).Error; err != nil {
			tx.Rollback()
			c.Ctx.JSON(400, gin.H{
				"error":   "File not found",
				"message": err.Error(),
			})
			return
		}

		updates["file_id"] = newFile.ID
		if err := tx.Commit().Error; err != nil {
			c.Ctx.JSON(400, gin.H{
				"error":   "Failed to update file",
				"message": err.Error(),
			})
			return
		}
	}

	if err := c.Container.GetService("buildingAdminAdvertisement").(building_admin_services.InterfaceBuildingAdminAdvertisementService).Update(form.ID, email, updates); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "update advertisement failed",
		})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "update advertisement success",
	})
}

func (c *BuildingAdminAdvertisementController) DeleteAdvertisement() {
	email := c.Ctx.GetString("email")
	if email == "" {
		c.Ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "invalid advertisement ID"})
		return
	}

	if err := c.Container.GetService("buildingAdminAdvertisement").(building_admin_services.InterfaceBuildingAdminAdvertisementService).Delete(uint(id), email); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "delete advertisement success"})
}

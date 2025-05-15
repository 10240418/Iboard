package http_base_controller

import (
	"strconv"

	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	base_services "github.com/The-Healthist/iboard_http_service/services/base"
	"github.com/The-Healthist/iboard_http_service/services/container"
	"github.com/The-Healthist/iboard_http_service/utils"
	"github.com/The-Healthist/iboard_http_service/utils/field"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type InterfaceBuildingAdminController interface {
	Create()
	Get()
	Update()
	Delete()
	GetOne()
}

type BuildingAdminController struct {
	Ctx       *gin.Context
	Container *container.ServiceContainer
}

func NewBuildingAdminController(ctx *gin.Context, container *container.ServiceContainer) *BuildingAdminController {
	return &BuildingAdminController{
		Ctx:       ctx,
		Container: container,
	}
}

// HandleFuncBuildingAdmin returns a gin.HandlerFunc for the specified method
func HandleFuncBuildingAdmin(container *container.ServiceContainer, method string) gin.HandlerFunc {
	switch method {
	case "create":
		return func(ctx *gin.Context) {
			controller := NewBuildingAdminController(ctx, container)
			controller.Create()
		}
	case "get":
		return func(ctx *gin.Context) {
			controller := NewBuildingAdminController(ctx, container)
			controller.Get()
		}
	case "update":
		return func(ctx *gin.Context) {
			controller := NewBuildingAdminController(ctx, container)
			controller.Update()
		}
	case "delete":
		return func(ctx *gin.Context) {
			controller := NewBuildingAdminController(ctx, container)
			controller.Delete()
		}
	case "getOne":
		return func(ctx *gin.Context) {
			controller := NewBuildingAdminController(ctx, container)
			controller.GetOne()
		}
	default:
		return func(ctx *gin.Context) {
			ctx.JSON(400, gin.H{"error": "invalid method"})
		}
	}
}

func (c *BuildingAdminController) Create() {
	var form struct {
		Email    string       `json:"email"    binding:"required"`
		Password string       `json:"password" binding:"required"`
		Status   field.Status `json:"status"   binding:"required"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	// Validate status
	if !field.IsValidStatus(string(form.Status)) {
		c.Ctx.JSON(400, gin.H{
			"error":   "invalid status",
			"message": "status must be one of: active, inactive, pending",
		})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(form.Password), bcrypt.DefaultCost)
	if err != nil {
		c.Ctx.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "password encryption failed",
		})
		return
	}

	buildingAdmin := &base_models.BuildingAdmin{
		Email:    form.Email,
		Password: string(hashedPassword),
		Status:   form.Status,
	}

	if err := c.Container.GetService("buildingAdmin").(base_services.InterfaceBuildingAdminService).Create(buildingAdmin); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "create building admin failed",
		})
		return
	}

	buildingAdmin.Password = "" // Don't return password
	c.Ctx.JSON(200, gin.H{
		"message": "create building admin success",
		"data":    buildingAdmin,
	})
}

func (c *BuildingAdminController) Get() {
	var searchQuery struct {
		BuildingID string        `form:"buildingId"`
		Status     *field.Status `form:"status"`
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

	buildingAdmins, paginationResult, err := c.Container.GetService("buildingAdmin").(base_services.InterfaceBuildingAdminService).Get(queryMap, paginationMap)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"data":       buildingAdmins,
		"pagination": paginationResult,
	})
}

func (c *BuildingAdminController) Update() {
	var form struct {
		ID       uint          `json:"id" binding:"required"`
		Password string        `json:"password"`
		Status   *field.Status `json:"status"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}

	if form.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(form.Password), bcrypt.DefaultCost)
		if err != nil {
			c.Ctx.JSON(500, gin.H{"error": "password encryption failed"})
			return
		}
		updates["password"] = string(hashedPassword)
	}

	if form.Status != nil {
		updates["status"] = *form.Status
	}

	if err := c.Container.GetService("buildingAdmin").(base_services.InterfaceBuildingAdminService).Update(form.ID, updates); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "update building admin success"})
}

func (c *BuildingAdminController) Delete() {
	var form struct {
		IDs []uint `json:"ids" binding:"required"`
	}
	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := c.Container.GetService("buildingAdmin").(base_services.InterfaceBuildingAdminService).Delete(form.IDs); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "delete building admin success"})
}

func (c *BuildingAdminController) GetOne() {
	if c.Container.GetService("jwt") == nil {
		c.Ctx.JSON(500, gin.H{
			"error":   "jwt service is nil",
			"message": "internal server error",
		})
		return
	}

	idStr := c.Ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   "Invalid building admin ID",
			"message": "Please check the ID format",
		})
		return
	}

	buildingAdmin, err := c.Container.GetService("buildingAdmin").(base_services.InterfaceBuildingAdminService).GetByID(uint(id))
	if err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "Failed to get building admin",
		})
		return
	}

	buildingAdmin.Password = "" // Don't return password
	c.Ctx.JSON(200, gin.H{
		"message": "Get building admin success",
		"data":    buildingAdmin,
	})
}

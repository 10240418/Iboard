package http_base_controller

import (
	"fmt"
	"strconv"
	"strings"

	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	base_services "github.com/The-Healthist/iboard_http_service/services/base"
	"github.com/The-Healthist/iboard_http_service/services/container"
	"github.com/The-Healthist/iboard_http_service/utils"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type InterfaceSuperAdminController interface {
	Login()
	GetSuperAdmins()
	CreateSuperAdmin()
	DeleteSuperAdmin()
	ResetPassword()
	ChangePassword()
	GetOne()
}

type SuperAdminController struct {
	Ctx       *gin.Context
	Container *container.ServiceContainer
}

func NewSuperAdminController(ctx *gin.Context, container *container.ServiceContainer) *SuperAdminController {
	return &SuperAdminController{
		Ctx:       ctx,
		Container: container,
	}
}

// HandleFuncSuperAdmin returns a gin.HandlerFunc for the specified method
func HandleFuncSuperAdmin(container *container.ServiceContainer, method string) gin.HandlerFunc {
	switch method {
	case "login":
		return func(ctx *gin.Context) {
			controller := NewSuperAdminController(ctx, container)
			controller.Login()
		}
	case "getSuperAdmins":
		return func(ctx *gin.Context) {
			controller := NewSuperAdminController(ctx, container)
			controller.GetSuperAdmins()
		}
	case "createSuperAdmin":
		return func(ctx *gin.Context) {
			controller := NewSuperAdminController(ctx, container)
			controller.CreateSuperAdmin()
		}
	case "deleteSuperAdmin":
		return func(ctx *gin.Context) {
			controller := NewSuperAdminController(ctx, container)
			controller.DeleteSuperAdmin()
		}
	case "resetPassword":
		return func(ctx *gin.Context) {
			controller := NewSuperAdminController(ctx, container)
			controller.ResetPassword()
		}
	case "changePassword":
		return func(ctx *gin.Context) {
			controller := NewSuperAdminController(ctx, container)
			controller.ChangePassword()
		}
	case "getOne":
		return func(ctx *gin.Context) {
			controller := NewSuperAdminController(ctx, container)
			controller.GetOne()
		}
	default:
		return func(ctx *gin.Context) {
			ctx.JSON(400, gin.H{"error": "invalid method"})
		}
	}
}

func (c *SuperAdminController) Login() {
	var form struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	form.Email = strings.TrimSpace(form.Email)
	form.Password = strings.TrimSpace(form.Password)

	if err := c.Container.GetService("superAdmin").(base_services.InterfaceSuperAdminService).CheckPassword(form.Email, form.Password); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "login failed",
		})
		return
	}

	// Get admin info
	admin, err := c.Container.GetService("superAdmin").(base_services.InterfaceSuperAdminService).GetSuperAdminByEmail(form.Email)
	if err != nil {
		c.Ctx.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "failed to get admin info",
		})
		return
	}

	// Generate token
	token, err := c.Container.GetService("jwt").(base_services.IJWTService).GenerateSuperAdminToken(admin)
	if err != nil {
		c.Ctx.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "failed to generate token",
		})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "login success",
		"token":   token,
	})
}

func (c *SuperAdminController) GetSuperAdmins() {
	var searchQuery struct {
		Search string `form:"search"`
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

	admins, paginationResult, err := c.Container.GetService("superAdmin").(base_services.InterfaceSuperAdminService).GetSuperAdmins(queryMap, paginationMap)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"data":       admins,
		"pagination": paginationResult,
	})
}

func (c *SuperAdminController) CreateSuperAdmin() {
	var form struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	superAdmin := &base_models.SuperAdmin{
		Email:    strings.TrimSpace(form.Email),
		Password: strings.TrimSpace(form.Password),
	}

	if err := c.Container.GetService("superAdmin").(base_services.InterfaceSuperAdminService).CreateSuperAdmin(superAdmin); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "create super admin failed",
		})
		return
	}

	superAdmin.Password = ""
	c.Ctx.JSON(200, gin.H{
		"message": "create super admin success",
		"data":    superAdmin,
	})
}

func (c *SuperAdminController) DeleteSuperAdmin() {
	var form struct {
		IDs []uint `json:"ids" binding:"required"`
	}
	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	claims, exists := c.Ctx.Get("claims")
	if !exists {
		c.Ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	mapClaims, ok := claims.(jwt.MapClaims)
	if !ok {
		c.Ctx.JSON(500, gin.H{"error": "invalid token claims format"})
		return
	}

	currentIDFloat, ok := mapClaims["id"].(float64)
	if !ok {
		c.Ctx.JSON(500, gin.H{"error": "invalid id in token"})
		return
	}
	currentID := uint(currentIDFloat)

	for _, id := range form.IDs {
		if id == currentID {
			c.Ctx.JSON(400, gin.H{
				"error": "cannot delete yourself",
				"id":    id,
			})
			return
		}
	}

	if err := c.Container.GetService("superAdmin").(base_services.InterfaceSuperAdminService).DeleteSuperAdmins(form.IDs); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "delete super admin success"})
}

func (c *SuperAdminController) ResetPassword() {
	var form struct {
		ID uint `json:"id" binding:"required"`
	}
	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	admin, err := c.Container.GetService("superAdmin").(base_services.InterfaceSuperAdminService).GetSuperAdminById(form.ID)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	randPassword := utils.RandStr(8, "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(randPassword), bcrypt.DefaultCost)
	if err != nil {
		c.Ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if err := c.Container.GetService("superAdmin").(base_services.InterfaceSuperAdminService).UpdateSuperAdmin(admin, map[string]interface{}{
		"password": string(hashPassword),
	}); err != nil {
		c.Ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	emailContent := fmt.Sprintf("Your password has been reset to: %s", randPassword)
	if err := c.Container.GetService("email").(base_services.IEmailService).SendEmail(
		[]string{admin.Email},
		"Password Reset Notification",
		emailContent,
	); err != nil {
		c.Ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "reset password success"})
}

func (c *SuperAdminController) ChangePassword() {
	var form struct {
		ID          uint   `json:"id" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required"`
	}
	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	admin, err := c.Container.GetService("superAdmin").(base_services.InterfaceSuperAdminService).GetSuperAdminById(form.ID)
	if err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   "admin not found",
			"message": err.Error(),
		})
		return
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(form.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.Ctx.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "password encryption failed",
		})
		return
	}

	if err := c.Container.GetService("superAdmin").(base_services.InterfaceSuperAdminService).UpdateSuperAdmin(admin, map[string]interface{}{
		"password": string(hashPassword),
	}); err != nil {
		c.Ctx.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "update password failed",
		})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "change password success"})
}

func (c *SuperAdminController) GetOne() {
	idStr := c.Ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Invalid super admin ID"})
		return
	}

	superAdmin, err := c.Container.GetService("superAdmin").(base_services.InterfaceSuperAdminService).GetSuperAdminById(uint(id))
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	superAdmin.Password = "" // Don't return password
	c.Ctx.JSON(200, gin.H{"data": superAdmin})
}

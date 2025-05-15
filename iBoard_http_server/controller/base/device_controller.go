package http_base_controller

import (
	"fmt"
	"strconv"

	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	base_services "github.com/The-Healthist/iboard_http_service/services/base"
	"github.com/The-Healthist/iboard_http_service/services/container"
	relationship_service "github.com/The-Healthist/iboard_http_service/services/relationship"
	"github.com/The-Healthist/iboard_http_service/utils"
	"github.com/gin-gonic/gin"
)

type InterfaceDeviceController interface {
	Create()
	CreateMany()
	Get()
	Update()
	Delete()
	GetOne()
	Login()
	GetDeviceAdvertisements()
	GetDeviceNotices()
	HealthTest()
}

type DeviceController struct {
	Ctx       *gin.Context
	Container *container.ServiceContainer
}

func NewDeviceController(ctx *gin.Context, container *container.ServiceContainer) *DeviceController {
	return &DeviceController{
		Ctx:       ctx,
		Container: container,
	}
}

// HandleFuncDevice returns a gin.HandlerFunc for the specified method
func HandleFuncDevice(container *container.ServiceContainer, method string) gin.HandlerFunc {
	switch method {
	case "create":
		return func(ctx *gin.Context) {
			controller := NewDeviceController(ctx, container)
			controller.Create()
		}
	case "createMany":
		return func(ctx *gin.Context) {
			controller := NewDeviceController(ctx, container)
			controller.CreateMany()
		}
	case "get":
		return func(ctx *gin.Context) {
			controller := NewDeviceController(ctx, container)
			controller.Get()
		}
	case "update":
		return func(ctx *gin.Context) {
			controller := NewDeviceController(ctx, container)
			controller.Update()
		}
	case "delete":
		return func(ctx *gin.Context) {
			controller := NewDeviceController(ctx, container)
			controller.Delete()
		}
	case "getOne":
		return func(ctx *gin.Context) {
			controller := NewDeviceController(ctx, container)
			controller.GetOne()
		}
	case "login":
		return func(ctx *gin.Context) {
			controller := NewDeviceController(ctx, container)
			controller.Login()
		}
	case "getDeviceAdvertisements":
		return func(ctx *gin.Context) {
			controller := NewDeviceController(ctx, container)
			controller.GetDeviceAdvertisements()
		}
	case "getDeviceNotices":
		return func(ctx *gin.Context) {
			controller := NewDeviceController(ctx, container)
			controller.GetDeviceNotices()
		}
	case "healthTest":
		return func(ctx *gin.Context) {
			controller := NewDeviceController(ctx, container)
			controller.HealthTest()
		}
	default:
		return func(ctx *gin.Context) {
			ctx.JSON(400, gin.H{"error": "invalid method"})
		}
	}
}

func (c *DeviceController) Create() {
	var form struct {
		DeviceID   string                     `json:"deviceId" binding:"required"`
		BuildingID uint                       `json:"buildingId" binding:"required"`
		Settings   base_models.DeviceSettings `json:"settings"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	device := &base_models.Device{
		DeviceID:   form.DeviceID,
		BuildingID: form.BuildingID,
		Settings:   form.Settings,
	}

	if err := c.Container.GetService("device").(base_services.InterfaceDeviceService).Create(device); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "create device failed",
		})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "create device success",
		"data":    device,
	})
}

func (c *DeviceController) Get() {
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
		Desc:     true,
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

	devices, paginationResult, err := c.Container.GetService("device").(base_services.InterfaceDeviceService).GetWithStatus(queryMap, paginationMap)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"data":       devices,
		"pagination": paginationResult,
	})
}

func (c *DeviceController) Update() {
	var form struct {
		ID         uint                        `json:"id" binding:"required"`
		DeviceID   string                      `json:"deviceId"`
		BuildingID uint                        `json:"buildingId"`
		Settings   *base_models.DeviceSettings `json:"settings"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 如果需要更新 buildingId，先处理绑定关系
	if form.BuildingID != 0 {
		// 先解绑
		if err := c.Container.GetService("deviceBuilding").(relationship_service.InterfaceDeviceBuildingService).UnbindDevice(form.ID); err != nil {
			c.Ctx.JSON(400, gin.H{"error": fmt.Sprintf("failed to unbind device: %v", err)})
			return
		}

		// 再绑定到新建筑物
		if err := c.Container.GetService("deviceBuilding").(relationship_service.InterfaceDeviceBuildingService).BindDevices(form.BuildingID, []uint{form.ID}); err != nil {
			c.Ctx.JSON(400, gin.H{"error": fmt.Sprintf("failed to bind device: %v", err)})
			return
		}
	}

	// 更新其他字段
	updates := map[string]interface{}{}
	if form.DeviceID != "" {
		updates["device_id"] = form.DeviceID
	}
	if form.Settings != nil {
		updates["arrearage_update_duration"] = form.Settings.ArrearageUpdateDuration
		updates["notice_update_duration"] = form.Settings.NoticeUpdateDuration
		updates["advertisement_update_duration"] = form.Settings.AdvertisementUpdateDuration
		updates["advertisement_play_duration"] = form.Settings.AdvertisementPlayDuration
		updates["notice_play_duration"] = form.Settings.NoticePlayDuration
		updates["spare_duration"] = form.Settings.SpareDuration
		updates["notice_stay_duration"] = form.Settings.NoticeStayDuration
	}

	updatedDevice, err := c.Container.GetService("device").(base_services.InterfaceDeviceService).Update(form.ID, updates)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 获取设备状态
	deviceWithStatus := base_services.DeviceWithStatus{
		Device: *updatedDevice,
		Status: c.Container.GetService("device").(base_services.InterfaceDeviceService).CheckDeviceStatus(updatedDevice.ID),
	}

	c.Ctx.JSON(200, gin.H{
		"message": "update device success",
		"data":    deviceWithStatus,
	})
}

func (c *DeviceController) Delete() {
	var form struct {
		IDs []uint `json:"ids" binding:"required"`
	}
	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := c.Container.GetService("device").(base_services.InterfaceDeviceService).Delete(form.IDs); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "delete device success"})
}

func (c *DeviceController) GetOne() {
	idStr := c.Ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Invalid device ID"})
		return
	}

	device, err := c.Container.GetService("device").(base_services.InterfaceDeviceService).GetByIDWithStatus(uint(id))
	if err != nil {
		c.Ctx.JSON(404, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "Get device success",
		"data":    device,
	})
}

func (c *DeviceController) CreateMany() {
	var form struct {
		Devices []struct {
			DeviceID   string                     `json:"deviceId" binding:"required"`
			BuildingID uint                       `json:"buildingId" binding:"required"`
			Settings   base_models.DeviceSettings `json:"settings"`
		} `json:"devices" binding:"required,dive"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	devices := make([]*base_models.Device, len(form.Devices))
	for i, d := range form.Devices {
		devices[i] = &base_models.Device{
			DeviceID:   d.DeviceID,
			BuildingID: d.BuildingID,
			Settings:   d.Settings,
		}
	}

	if err := c.Container.GetService("device").(base_services.InterfaceDeviceService).CreateMany(devices); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "create devices failed",
		})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "create devices success",
		"data":    devices,
	})
}

func (c *DeviceController) Login() {
	var form struct {
		DeviceID string `json:"deviceId" binding:"required"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "Invalid form",
		})
		return
	}

	// Get device by deviceId
	device, err := c.Container.GetService("device").(base_services.InterfaceDeviceService).GetByDeviceID(form.DeviceID)
	if err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "Invalid device ID",
		})
		return
	}

	// Generate JWT token
	token, err := c.Container.GetService("jwt").(base_services.IJWTService).GenerateDeviceToken(device)
	if err != nil {
		c.Ctx.JSON(500, gin.H{
			"error":   err.Error(),
			"message": "Failed to generate token",
		})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "Login success",
		"data":    device,
		"token":   token,
	})
}

func (c *DeviceController) GetDeviceAdvertisements() {
	claims, exists := c.Ctx.Get("claims")
	if !exists {
		c.Ctx.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	claimsMap, ok := claims.(map[string]interface{})
	if !ok {
		c.Ctx.JSON(500, gin.H{"error": "Invalid claims format"})
		return
	}

	deviceId, ok := claimsMap["deviceId"].(string)
	if !ok {
		c.Ctx.JSON(500, gin.H{"error": "Invalid device ID format"})
		return
	}

	advertisements, err := c.Container.GetService("device").(base_services.InterfaceDeviceService).GetDeviceAdvertisements(deviceId)
	if err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "Failed to get advertisements",
		})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "Get advertisements success",
		"data":    advertisements,
	})
}

func (c *DeviceController) GetDeviceNotices() {
	claims, exists := c.Ctx.Get("claims")
	if !exists {
		c.Ctx.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	claimsMap, ok := claims.(map[string]interface{})
	if !ok {
		c.Ctx.JSON(500, gin.H{"error": "Invalid claims format"})
		return
	}

	deviceId, ok := claimsMap["deviceId"].(string)
	if !ok {
		c.Ctx.JSON(500, gin.H{"error": "Invalid device ID format"})
		return
	}

	notices, err := c.Container.GetService("device").(base_services.InterfaceDeviceService).GetDeviceNotices(deviceId)
	if err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "Failed to get notices",
		})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "Get notices success",
		"data":    notices,
	})
}

func (c *DeviceController) HealthTest() {
	claims, exists := c.Ctx.Get("claims")
	if !exists {
		c.Ctx.JSON(401, gin.H{"error": "No token claims found"})
		return
	}

	claimsMap, ok := claims.(map[string]interface{})
	if !ok {
		c.Ctx.JSON(401, gin.H{"error": "Invalid token claims format"})
		return
	}

	deviceId, ok := claimsMap["deviceId"].(string)
	if !ok {
		c.Ctx.JSON(401, gin.H{"error": "Invalid device ID format"})
		return
	}

	// 获取设备信息
	device, err := c.Container.GetService("device").(base_services.InterfaceDeviceService).GetByDeviceID(deviceId)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Device not found"})
		return
	}

	// 更新设备健康状态
	if err := c.Container.GetService("device").(base_services.InterfaceDeviceService).UpdateDeviceHealth(device.ID); err != nil {
		c.Ctx.JSON(500, gin.H{
			"error":   "Failed to update device health status",
			"message": err.Error(),
		})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "Health check successful",
	})
}

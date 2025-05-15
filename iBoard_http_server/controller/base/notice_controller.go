package http_base_controller

import (
	"strconv"
	"time"

	databases "github.com/The-Healthist/iboard_http_service/database"
	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	base_services "github.com/The-Healthist/iboard_http_service/services/base"
	"github.com/The-Healthist/iboard_http_service/services/container"
	"github.com/The-Healthist/iboard_http_service/utils"
	"github.com/The-Healthist/iboard_http_service/utils/field"
	"github.com/gin-gonic/gin"
)

type InterfaceNoticeController interface {
	Create()
	CreateMany()
	Get()
	Update()
	Delete()
	GetOne()
}

type NoticeController struct {
	Ctx       *gin.Context
	Container *container.ServiceContainer
}

func NewNoticeController(ctx *gin.Context, container *container.ServiceContainer) *NoticeController {
	return &NoticeController{
		Ctx:       ctx,
		Container: container,
	}
}

// HandleFuncNotice returns a gin.HandlerFunc for the specified method
func HandleFuncNotice(container *container.ServiceContainer, method string) gin.HandlerFunc {
	switch method {
	case "create":
		return func(ctx *gin.Context) {
			controller := NewNoticeController(ctx, container)
			controller.Create()
		}
	case "createMany":
		return func(ctx *gin.Context) {
			controller := NewNoticeController(ctx, container)
			controller.CreateMany()
		}
	case "get":
		return func(ctx *gin.Context) {
			controller := NewNoticeController(ctx, container)
			controller.Get()
		}
	case "update":
		return func(ctx *gin.Context) {
			controller := NewNoticeController(ctx, container)
			controller.Update()
		}
	case "delete":
		return func(ctx *gin.Context) {
			controller := NewNoticeController(ctx, container)
			controller.Delete()
		}
	case "getOne":
		return func(ctx *gin.Context) {
			controller := NewNoticeController(ctx, container)
			controller.GetOne()
		}
	default:
		return func(ctx *gin.Context) {
			ctx.JSON(400, gin.H{"error": "invalid method"})
		}
	}
}

type CreateNoticeRequest struct {
	Title          string           `json:"title" binding:"required"`
	Description    string           `json:"description"`
	Type           field.NoticeType `json:"type" binding:"required"`
	Status         field.Status     `json:"status" binding:"required"`
	StartTime      *time.Time       `json:"startTime" binding:"required"`
	EndTime        *time.Time       `json:"endTime" binding:"required"`
	IsPublic       bool             `json:"isPublic"`
	IsIsmartNotice bool             `json:"isIsmartNotice"`
	Priority       int              `json:"priority"`
	Path           string           `json:"path" binding:"required"`
	FileType       field.FileType   `json:"fileType"`
}

func (c *NoticeController) Create() {
	var form CreateNoticeRequest

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	// Set default values for startTime, endTime, and status
	now := time.Now()
	startTime := now
	endTime := now.AddDate(1, 0, 0) // 1 year from now
	status := field.Status("active")

	if form.StartTime != nil {
		startTime = *form.StartTime
	}
	if form.EndTime != nil {
		endTime = *form.EndTime
	}
	if form.Status != "" {
		status = form.Status
	}

	// If path is provided, find the corresponding file
	var fileID *uint
	if form.Path != "" {
		var file base_models.File
		if err := databases.DB_CONN.Where("path = ?", form.Path).First(&file).Error; err != nil {
			c.Ctx.JSON(400, gin.H{
				"error":   err.Error(),
				"message": "file not found",
			})
			return
		}
		fileID = &file.ID
	}

	notice := &base_models.Notice{
		Title:          form.Title,
		Description:    form.Description,
		Type:           form.Type,
		Status:         status,
		StartTime:      startTime,
		EndTime:        endTime,
		IsPublic:       form.IsPublic,
		IsIsmartNotice: form.IsIsmartNotice,
		Priority:       form.Priority,
		FileID:         fileID,
		FileType:       form.FileType,
	}

	if err := c.Container.GetService("notice").(base_services.InterfaceNoticeService).Create(notice); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "create notice failed",
		})
		return
	}

	// 重新加载 notice 以获取关联的文件信息
	if err := databases.DB_CONN.Preload("File").First(notice, notice.ID).Error; err != nil {
		c.Ctx.JSON(200, gin.H{
			"message": "create notice success, but failed to load file info",
			"data":    notice,
		})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "create notice success",
		"data":    notice,
	})
}

func (c *NoticeController) CreateMany() {
	var forms []struct {
		Title          string           `json:"title" binding:"required"`
		Description    string           `json:"description"`
		Type           field.NoticeType `json:"type" binding:"required"`
		Status         field.Status     `json:"status" binding:"required"`
		StartTime      *time.Time       `json:"startTime" binding:"required"`
		EndTime        *time.Time       `json:"endTime" binding:"required"`
		IsPublic       bool             `json:"isPublic" binding:"required"`
		IsIsmartNotice bool             `json:"isIsmartNotice"`
		Priority       int              `json:"priority" binding:"required"`
		Path           string           `json:"path" binding:"required"`
		FileType       field.FileType   `json:"fileType"`
		Duration       int              `json:"duration" binding:"required"`
	}

	if err := c.Ctx.ShouldBindJSON(&forms); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	var notices []*base_models.Notice
	for _, form := range forms {
		notice := &base_models.Notice{
			Title:          form.Title,
			Description:    form.Description,
			Type:           form.Type,
			Status:         form.Status,
			StartTime:      *form.StartTime,
			EndTime:        *form.EndTime,
			IsPublic:       form.IsPublic,
			IsIsmartNotice: form.IsIsmartNotice,
			Priority:       form.Priority,
		}
		notices = append(notices, notice)
	}

	for _, notice := range notices {
		if err := c.Container.GetService("notice").(base_services.InterfaceNoticeService).Create(notice); err != nil {
			c.Ctx.JSON(400, gin.H{
				"error":   err.Error(),
				"message": "create notice failed",
				"notice":  notice,
			})
			return
		}
	}

	c.Ctx.JSON(200, gin.H{
		"message": "create notices success",
		"data":    notices,
	})
}

func (c *NoticeController) Get() {
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

	notices, paginationResult, err := c.Container.GetService("notice").(base_services.InterfaceNoticeService).Get(queryMap, paginationMap)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"data":       notices,
		"pagination": paginationResult,
	})
}

func (c *NoticeController) Update() {
	var form struct {
		ID             uint             `json:"id" binding:"required"`
		Title          string           `json:"title"`
		Description    string           `json:"description"`
		Type           field.NoticeType `json:"type"`
		Status         field.Status     `json:"status"`
		StartTime      *time.Time       `json:"startTime"`
		EndTime        *time.Time       `json:"endTime"`
		IsPublic       *bool            `json:"isPublic"`
		IsIsmartNotice *bool            `json:"isIsmartNotice"`
		Priority       *int             `json:"priority"`
		Path           string           `json:"path"`
		FileType       field.FileType   `json:"fileType"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 获取原有的通知信息
	notice, err := c.Container.GetService("notice").(base_services.InterfaceNoticeService).GetByID(form.ID)
	if err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to get notice",
			"message": err.Error(),
		})
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
	if form.IsPublic != nil {
		updates["is_public"] = *form.IsPublic
	}
	if form.FileType != "" {
		updates["file_type"] = form.FileType
	}

	// 如果提供了新的 path
	if form.Path != "" {
		// 开启事务
		tx := databases.DB_CONN.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		// 查找新的文件
		var newFile base_models.File
		if err := tx.Where("path = ?", form.Path).First(&newFile).Error; err != nil {
			tx.Rollback()
			c.Ctx.JSON(400, gin.H{
				"error":   "File not found",
				"message": err.Error(),
			})
			return
		}

		// 保存旧文件ID
		var oldFileID *uint
		if notice.FileID != nil {
			oldFileID = notice.FileID
		}

		// 1. 更新通知的文件ID
		updates["file_id"] = newFile.ID
		if err := tx.Model(&base_models.Notice{}).Where("id = ?", form.ID).Updates(updates).Error; err != nil {
			tx.Rollback()
			c.Ctx.JSON(400, gin.H{
				"error":   "Failed to update notice",
				"message": err.Error(),
			})
			return
		}

		// 2. 如果有旧文件，检查是否需要删除
		if oldFileID != nil {
			// 检查这个文件是否还被其他地方引用
			var adCount int64
			var noticeCount int64
			if err := tx.Model(&base_models.Advertisement{}).Where("file_id = ?", oldFileID).Count(&adCount).Error; err != nil {
				tx.Rollback()
				c.Ctx.JSON(400, gin.H{
					"error":   "Failed to check advertisement references",
					"message": err.Error(),
				})
				return
			}
			if err := tx.Model(&base_models.Notice{}).Where("file_id = ?", oldFileID).Count(&noticeCount).Error; err != nil {
				tx.Rollback()
				c.Ctx.JSON(400, gin.H{
					"error":   "Failed to check notice references",
					"message": err.Error(),
				})
				return
			}

			// 3. 如果没有其他引用，则删除文件
			if adCount == 0 && noticeCount == 0 {
				if err := tx.Delete(&base_models.File{}, "id = ?", oldFileID).Error; err != nil {
					tx.Rollback()
					c.Ctx.JSON(400, gin.H{
						"error":   "Failed to delete old file",
						"message": err.Error(),
					})
					return
				}
			}
		}

		// 4. 获取更新后的通知信息
		var updatedNotice base_models.Notice
		if err := tx.Preload("File").First(&updatedNotice, form.ID).Error; err != nil {
			tx.Rollback()
			c.Ctx.JSON(400, gin.H{
				"error":   "Failed to get updated notice",
				"message": err.Error(),
			})
			return
		}

		// 提交事务
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			c.Ctx.JSON(400, gin.H{
				"error":   "Failed to commit transaction",
				"message": err.Error(),
			})
			return
		}

		c.Ctx.JSON(200, gin.H{
			"message": "update notice success",
			"data":    updatedNotice,
		})
		return
	}

	// 如果没有更新文件，则直接更新其他字段
	updatedNotice, err := c.Container.GetService("notice").(base_services.InterfaceNoticeService).Update(form.ID, updates)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "update notice success",
		"data":    updatedNotice,
	})
}

func (c *NoticeController) Delete() {
	var form struct {
		IDs []uint `json:"ids" binding:"required"`
	}
	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 开启事务
	tx := databases.DB_CONN.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 获取所有要删除的通知信息
	var notices []base_models.Notice
	if err := tx.Where("id IN ?", form.IDs).Find(&notices).Error; err != nil {
		tx.Rollback()
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to get notices",
			"message": err.Error(),
		})
		return
	}

	// 2. 解除与建筑物的关联
	if err := tx.Exec("DELETE FROM notice_buildings WHERE notice_id IN ?", form.IDs).Error; err != nil {
		tx.Rollback()
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to unbind buildings",
			"message": err.Error(),
		})
		return
	}

	// 3. 收集所有关联的文件ID
	var fileIDs []uint
	for _, notice := range notices {
		if notice.FileID != nil {
			fileIDs = append(fileIDs, *notice.FileID)
		}
	}

	// 4. 解除文件绑定
	if err := tx.Model(&base_models.Notice{}).Where("id IN ?", form.IDs).Update("file_id", nil).Error; err != nil {
		tx.Rollback()
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to unbind files",
			"message": err.Error(),
		})
		return
	}

	// 5. 检查每个文件的引用并删除未被引用的文件
	for _, fileID := range fileIDs {
		var adCount int64
		var noticeCount int64

		if err := tx.Model(&base_models.Advertisement{}).Where("file_id = ?", fileID).Count(&adCount).Error; err != nil {
			tx.Rollback()
			c.Ctx.JSON(400, gin.H{
				"error":   "Failed to check advertisement references",
				"message": err.Error(),
			})
			return
		}

		if err := tx.Model(&base_models.Notice{}).Where("file_id = ?", fileID).Count(&noticeCount).Error; err != nil {
			tx.Rollback()
			c.Ctx.JSON(400, gin.H{
				"error":   "Failed to check notice references",
				"message": err.Error(),
			})
			return
		}

		if adCount == 0 && noticeCount == 0 {
			if err := tx.Delete(&base_models.File{}, "id = ?", fileID).Error; err != nil {
				tx.Rollback()
				c.Ctx.JSON(400, gin.H{
					"error":   "Failed to delete file",
					"message": err.Error(),
				})
				return
			}
		}
	}

	// 6. 删除通知
	if err := tx.Delete(&base_models.Notice{}, form.IDs).Error; err != nil {
		tx.Rollback()
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to delete notices",
			"message": err.Error(),
		})
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to commit transaction",
			"message": err.Error(),
		})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "delete notice success"})
}

func (c *NoticeController) GetOne() {
	idStr := c.Ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Invalid notice ID"})
		return
	}

	notice, err := c.Container.GetService("notice").(base_services.InterfaceNoticeService).GetByID(uint(id))
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"data": notice})
}

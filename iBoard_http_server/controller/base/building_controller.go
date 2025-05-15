package http_base_controller

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	databases "github.com/The-Healthist/iboard_http_service/database"
	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	base_services "github.com/The-Healthist/iboard_http_service/services/base"
	"github.com/The-Healthist/iboard_http_service/services/container"
	"github.com/The-Healthist/iboard_http_service/utils"
	"github.com/The-Healthist/iboard_http_service/utils/field"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InterfaceBuildingController interface {
	Create()
	Get()
	Update()
	Delete()
	GetOne()
	SyncNotice()
	ManualSyncNotice()
}

type BuildingController struct {
	Ctx       *gin.Context
	Container *container.ServiceContainer
}

func NewBuildingController(ctx *gin.Context, container *container.ServiceContainer) *BuildingController {
	return &BuildingController{
		Ctx:       ctx,
		Container: container,
	}
}

// HandleFuncBuilding returns a gin.HandlerFunc for the specified method
func HandleFuncBuilding(container *container.ServiceContainer, method string) gin.HandlerFunc {
	switch method {
	case "create":
		return func(ctx *gin.Context) {
			controller := NewBuildingController(ctx, container)
			controller.Create()
		}
	case "get":
		return func(ctx *gin.Context) {
			controller := NewBuildingController(ctx, container)
			controller.Get()
		}
	case "update":
		return func(ctx *gin.Context) {
			controller := NewBuildingController(ctx, container)
			controller.Update()
		}
	case "delete":
		return func(ctx *gin.Context) {
			controller := NewBuildingController(ctx, container)
			controller.Delete()
		}
	case "getOne":
		return func(ctx *gin.Context) {
			controller := NewBuildingController(ctx, container)
			controller.GetOne()
		}
	case "syncNotice":
		return func(ctx *gin.Context) {
			controller := NewBuildingController(ctx, container)
			controller.SyncNotice()
		}
	case "manualSyncNotice":
		return func(ctx *gin.Context) {
			controller := NewBuildingController(ctx, container)
			controller.ManualSyncNotice()
		}
	default:
		return func(ctx *gin.Context) {
			ctx.JSON(400, gin.H{"error": "invalid method"})
		}
	}
}

// Create creates a new building
func (c *BuildingController) Create() {
	var form struct {
		Name     string `json:"name" binding:"required"`
		IsmartID string `json:"ismartId" binding:"required"`
		Remark   string `json:"remark"`
		Location string `json:"location"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "invalid form",
		})
		return
	}

	building := &base_models.Building{
		Name:     form.Name,
		IsmartID: form.IsmartID,
		Remark:   form.Remark,
		Location: form.Location,
	}

	if err := c.Container.GetService("building").(base_services.InterfaceBuildingService).Create(building); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   err.Error(),
			"message": "create building failed",
		})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "create building success",
		"data":    building,
	})
}

func (c *BuildingController) Get() {
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

	buildings, paginationResult, err := c.Container.GetService("building").(base_services.InterfaceBuildingService).Get(queryMap, paginationMap)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 确保每个建筑的字段都被正确初始化
	for i := range buildings {
		if buildings[i].Location == "" {
			buildings[i].Location = "" // 确保即使是空字符串也会被序列化
		}
	}

	c.Ctx.JSON(200, gin.H{
		"data":       buildings,
		"pagination": paginationResult,
	})
}

func (c *BuildingController) Update() {
	var form struct {
		ID       uint   `json:"id" binding:"required"`
		Name     string `json:"name"`
		IsmartID string `json:"ismartId"`
		Remark   string `json:"remark"`
		Location string `json:"location"`
	}

	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if form.Name != "" {
		updates["name"] = form.Name
	}
	if form.IsmartID != "" {
		updates["ismart_id"] = form.IsmartID
	}
	if form.Remark != "" {
		updates["remark"] = form.Remark
	}
	updates["location"] = form.Location

	if err := c.Container.GetService("building").(base_services.InterfaceBuildingService).Update(form.ID, updates); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "update building success"})
}

func (c *BuildingController) Delete() {
	var form struct {
		IDs []uint `json:"ids" binding:"required"`
	}
	if err := c.Ctx.ShouldBindJSON(&form); err != nil {
		c.Ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// start transaction
	tx := databases.DB_CONN.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. get all buildings to be deleted
	var buildings []base_models.Building
	if err := tx.Preload("Notices").Preload("Advertisements").Where("id IN ?", form.IDs).Find(&buildings).Error; err != nil {
		tx.Rollback()
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to get buildings",
			"message": err.Error(),
		})
		return
	}

	// 2. collect all file ids of notices and advertisements
	fileIDMap := make(map[uint]bool)
	var noticeIDs []uint
	var advertisementIDs []uint
	for _, building := range buildings {
		for _, notice := range building.Notices {
			if notice.FileID != nil {
				fileIDMap[*notice.FileID] = true
			}
			noticeIDs = append(noticeIDs, notice.ID)
		}
		for _, ad := range building.Advertisements {
			if ad.FileID != nil {
				fileIDMap[*ad.FileID] = true
			}
			advertisementIDs = append(advertisementIDs, ad.ID)
		}
	}

	// convert map to slice
	fileIDs := make([]uint, 0, len(fileIDMap))
	for fileID := range fileIDMap {
		fileIDs = append(fileIDs, fileID)
	}

	// 3. unbind notices
	if len(noticeIDs) > 0 {
		// unbind notice-building
		if err := tx.Exec("DELETE FROM notice_buildings WHERE building_id IN ?", form.IDs).Error; err != nil {
			tx.Rollback()
			c.Ctx.JSON(400, gin.H{
				"error":   "Failed to unbind notices from buildings",
				"message": err.Error(),
			})
			return
		}
	}

	// 4. unbind advertisements
	if len(advertisementIDs) > 0 {
		// unbind advertisement-building
		if err := tx.Exec("DELETE FROM advertisement_buildings WHERE building_id IN ?", form.IDs).Error; err != nil {
			tx.Rollback()
			c.Ctx.JSON(400, gin.H{
				"error":   "Failed to unbind advertisements from buildings",
				"message": err.Error(),
			})
			return
		}
	}

	// 5. unbind admins
	if err := tx.Exec("DELETE FROM building_admins_buildings WHERE building_id IN ?", form.IDs).Error; err != nil {
		tx.Rollback()
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to unbind admins",
			"message": err.Error(),
		})
		return
	}

	// 6. delete buildings
	if err := tx.Delete(&base_models.Building{}, form.IDs).Error; err != nil {
		tx.Rollback()
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to delete buildings",
			"message": err.Error(),
		})
		return
	}

	// commit transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to commit transaction",
			"message": err.Error(),
		})
		return
	}

	c.Ctx.JSON(200, gin.H{"message": "delete building success"})
}

func (c *BuildingController) GetOne() {
	idStr := c.Ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.Ctx.JSON(400, gin.H{"error": "Invalid building ID"})
		return
	}

	building, err := c.Container.GetService("building").(base_services.InterfaceBuildingService).GetByIDWithDevices(uint(id))
	if err != nil {
		c.Ctx.JSON(404, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(200, gin.H{
		"message": "Get building success",
		"data":    building,
	})
}

func (c *BuildingController) SyncNotice() {
	// get building id from url param
	idStr := c.Ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   "Invalid building ID",
			"message": "Please check the ID format",
		})
		return
	}

	// get building info to get ismartId
	building, err := c.Container.GetService("building").(base_services.InterfaceBuildingService).GetByID(uint(id))
	if err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to get building information",
			"message": err.Error(),
		})
		return
	}

	// Get existing iSmart notices for this building
	var existingNotices []base_models.Notice
	if err := databases.DB_CONN.Joins("JOIN notice_buildings ON notices.id = notice_buildings.notice_id").
		Where("notice_buildings.building_id = ? AND notices.is_ismart_notice = ?", id, true).
		Preload("File").Find(&existingNotices).Error; err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to get existing notices",
			"message": err.Error(),
		})
		return
	}

	// prepare request to old system api
	url := "https://uqf0jqfm77.execute-api.ap-east-1.amazonaws.com/prod/v1/building_board/building-notices"
	reqBody := struct {
		BlgID string `json:"blg_id"`
	}{
		BlgID: building.IsmartID,
	}

	reqBodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to prepare request data",
			"message": err.Error(),
		})
		return
	}

	// request old system to get notice list
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBodyJSON))
	if err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to get notices from old system",
			"message": err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	// parse response data
	var respBody []struct {
		ID        int    `json:"id"`
		MessTitle string `json:"mess_title"`
		MessType  string `json:"mess_type"`
		MessFile  string `json:"mess_file"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		c.Ctx.JSON(400, gin.H{
			"error":   "Failed to parse response data",
			"message": err.Error(),
		})
		return
	}

	// Create a map of MD5 hashes from old system notices
	oldSystemMD5Map := make(map[string]bool)
	var successCount int
	var failedNotices []string
	var hasSyncedCount int
	var deleteCount int

	// Process old system notices and build MD5 map
	for _, oldNotice := range respBody {
		// download pdf file
		fileResp, err := http.Get(oldNotice.MessFile)
		if err != nil {
			failedNotices = append(failedNotices, fmt.Sprintf("Failed to download file for notice ID %d: %v", oldNotice.ID, err))
			continue
		}

		// read file content
		fileContent, err := io.ReadAll(fileResp.Body)
		fileResp.Body.Close()
		if err != nil {
			failedNotices = append(failedNotices, fmt.Sprintf("Failed to read file content for notice ID %d: %v", oldNotice.ID, err))
			continue
		}

		md5Hash := md5.Sum(fileContent)
		md5Str := hex.EncodeToString(md5Hash[:])
		oldSystemMD5Map[md5Str] = true
	}

	// Process existing notices that are no longer in old system
	for _, existingNotice := range existingNotices {
		if existingNotice.File == nil {
			continue
		}

		if !oldSystemMD5Map[existingNotice.File.Md5] {
			// Start transaction for deletion
			tx := databases.DB_CONN.Begin()

			// Unbind notice from building
			if err := tx.Exec("DELETE FROM notice_buildings WHERE notice_id = ? AND building_id = ?", existingNotice.ID, id).Error; err != nil {
				tx.Rollback()
				failedNotices = append(failedNotices, fmt.Sprintf("Failed to unbind notice ID %d: %v", existingNotice.ID, err))
				continue
			}

			// Check if notice is bound to other buildings
			var buildingCount int64
			if err := tx.Table("notice_buildings").Where("notice_id = ?", existingNotice.ID).Count(&buildingCount).Error; err != nil {
				tx.Rollback()
				failedNotices = append(failedNotices, fmt.Sprintf("Failed to check notice bindings for ID %d: %v", existingNotice.ID, err))
				continue
			}

			if buildingCount == 0 {
				// Delete notice if no other bindings exist
				if err := tx.Delete(&existingNotice).Error; err != nil {
					tx.Rollback()
					failedNotices = append(failedNotices, fmt.Sprintf("Failed to delete notice ID %d: %v", existingNotice.ID, err))
					continue
				}

				// Check if file is used by other notices
				var fileCount int64
				if err := tx.Model(&base_models.Notice{}).Where("file_id = ?", existingNotice.FileID).Count(&fileCount).Error; err != nil {
					tx.Rollback()
					failedNotices = append(failedNotices, fmt.Sprintf("Failed to check file references for notice ID %d: %v", existingNotice.ID, err))
					continue
				}

				if fileCount == 0 {
					// Delete file if no other notices use it
					if err := tx.Delete(&existingNotice.File).Error; err != nil {
						tx.Rollback()
						failedNotices = append(failedNotices, fmt.Sprintf("Failed to delete file for notice ID %d: %v", existingNotice.ID, err))
						continue
					}
				}
			}

			if err := tx.Commit().Error; err != nil {
				failedNotices = append(failedNotices, fmt.Sprintf("Failed to commit deletion for notice ID %d: %v", existingNotice.ID, err))
				continue
			}

			deleteCount++
		}
	}

	// Process notices from old system
	for _, oldNotice := range respBody {
		// download pdf file
		fileResp, err := http.Get(oldNotice.MessFile)
		if err != nil {
			failedNotices = append(failedNotices, fmt.Sprintf("Failed to download file for notice ID %d: %v", oldNotice.ID, err))
			continue
		}

		// read file content
		fileContent, err := io.ReadAll(fileResp.Body)
		fileResp.Body.Close()
		if err != nil {
			failedNotices = append(failedNotices, fmt.Sprintf("Failed to read file content for notice ID %d: %v", oldNotice.ID, err))
			continue
		}

		fileSize := len(fileContent)
		md5Hash := md5.Sum(fileContent)
		md5Str := hex.EncodeToString(md5Hash[:])
		mimeType := http.DetectContentType(fileContent)

		// Check if a notice with this file is already bound to the building
		var existingNoticeCount int64
		if err := databases.DB_CONN.Model(&base_models.Notice{}).
			Joins("JOIN notice_buildings ON notices.id = notice_buildings.notice_id").
			Joins("JOIN files ON notices.file_id = files.id").
			Where("notice_buildings.building_id = ? AND files.md5 = ? AND notices.is_ismart_notice = ?",
				id, md5Str, true).
			Count(&existingNoticeCount).Error; err != nil {
			failedNotices = append(failedNotices, fmt.Sprintf("Failed to check existing notice for file MD5 %s: %v", md5Str, err))
			continue
		}

		if existingNoticeCount > 0 {
			// Notice with this file is already bound to the building
			hasSyncedCount++
			continue
		}

		// get uploader info from token
		claims, exists := c.Ctx.Get("claims")
		if !exists {
			failedNotices = append(failedNotices, fmt.Sprintf("Failed to get uploader info for notice ID %d: token claims not found", oldNotice.ID))
			continue
		}

		mapClaims, ok := claims.(jwt.MapClaims)
		if !ok {
			failedNotices = append(failedNotices, fmt.Sprintf("Failed to get uploader info for notice ID %d: invalid claims format", oldNotice.ID))
			continue
		}

		var uploaderID uint
		var uploaderType field.FileUploaderType
		var uploaderEmail string

		if isAdmin, ok := mapClaims["isAdmin"].(bool); ok && isAdmin {
			if id, ok := mapClaims["id"].(float64); ok {
				uploaderID = uint(id)
				uploaderType = field.UploaderTypeSuperAdmin
				if email, ok := mapClaims["email"].(string); ok {
					uploaderEmail = email
				}
			}
		}

		if uploaderID == 0 {
			failedNotices = append(failedNotices, fmt.Sprintf("Failed to get uploader info for notice ID %d: invalid uploader info", oldNotice.ID))
			continue
		}

		// Check if file with same MD5 exists
		var existingFile base_models.File
		var shouldUpload bool = true
		var shouldAddFile bool = true
		var fileForNotice *base_models.File

		err = databases.DB_CONN.Where("md5 = ?", md5Str).First(&existingFile).Error
		if err == nil {
			// File exists, check if it's associated with any notice
			var existingNotice base_models.Notice
			err := databases.DB_CONN.Where("file_id = ?", existingFile.ID).First(&existingNotice).Error

			if err == nil {
				// Check if notice is bound to current building
				var count int64
				err = databases.DB_CONN.Table("notice_buildings").
					Where("notice_id = ? AND building_id = ?", existingNotice.ID, id).
					Count(&count).Error

				if err == nil && count > 0 {
					// Notice already exists for this building
					hasSyncedCount++
					continue
				}

				// Check if notice is bound to other buildings
				err = databases.DB_CONN.Table("notice_buildings").
					Where("notice_id = ?", existingNotice.ID).
					Count(&count).Error

				if err == nil {
					if count > 0 {
						// Notice exists and is bound to other buildings, create new notice
						shouldUpload = false
						shouldAddFile = false
						fileForNotice = &existingFile
					} else {
						// Notice exists but not bound to any building, delete it and create new one
						if err := databases.DB_CONN.Delete(&existingNotice).Error; err != nil {
							failedNotices = append(failedNotices, fmt.Sprintf("Failed to delete old notice for ID %d: %v", oldNotice.ID, err))
							continue
						}
						shouldUpload = false
						shouldAddFile = false
						fileForNotice = &existingFile
					}
				}
			} else {
				// File exists but no notice associated
				shouldUpload = false
				shouldAddFile = false
				fileForNotice = &existingFile
			}
		} else if err == gorm.ErrRecordNotFound {
			// File doesn't exist, create new one
			shouldUpload = true
			shouldAddFile = true

			// generate file name and get upload params
			currentTime := time.Now()
			dir := currentTime.Format("2006-01-02") + "/"
			fileName := fmt.Sprintf("%s.pdf", uuid.New().String())
			objectKey := dir + fileName
			uploadParams, err := c.Container.GetService("upload").(base_services.IUploadService).GetUploadParamsSync(objectKey)
			if err != nil {
				failedNotices = append(failedNotices, fmt.Sprintf("Failed to get upload params for notice ID %d: %v", oldNotice.ID, err))
				continue
			}

			fileForNotice = &base_models.File{
				Path:         uploadParams["host"].(string) + "/" + objectKey,
				Size:         int64(fileSize),
				MimeType:     mimeType,
				Oss:          "aws",
				UploaderType: uploaderType,
				UploaderID:   uploaderID,
				Uploader:     uploaderEmail,
				Md5:          md5Str,
			}

			if shouldUpload {
				// Upload file to OSS
				var b bytes.Buffer
				w := multipart.NewWriter(&b)

				formFields := []struct {
					key      string
					paramKey string
				}{
					{key: "key", paramKey: "dir"},
					{key: "policy", paramKey: "policy"},
					{key: "OSSAccessKeyId", paramKey: "accessid"},
					{key: "success_action_status", paramKey: ""},
					{key: "callback", paramKey: "callback"},
					{key: "signature", paramKey: "signature"},
				}

				for _, field := range formFields {
					var fieldValue string
					switch field.key {
					case "key":
						fieldValue = objectKey
					case "success_action_status":
						fieldValue = "200"
					default:
						if field.paramKey != "" {
							if val, ok := uploadParams[field.paramKey]; ok {
								fieldValue = fmt.Sprintf("%v", val)
							} else {
								continue
							}
						}
					}
					if err := w.WriteField(field.key, fieldValue); err != nil {
						failedNotices = append(failedNotices, fmt.Sprintf("Failed to write form field %s for notice ID %d: %v", field.key, oldNotice.ID, err))
						continue
					}
				}

				fw, err := w.CreateFormFile("file", objectKey)
				if err != nil {
					failedNotices = append(failedNotices, fmt.Sprintf("Failed to create file form for notice ID %d: %v", oldNotice.ID, err))
					continue
				}
				if _, err = io.Copy(fw, bytes.NewReader(fileContent)); err != nil {
					failedNotices = append(failedNotices, fmt.Sprintf("Failed to copy file data for notice ID %d: %v", oldNotice.ID, err))
					continue
				}
				w.Close()

				uploadReq, err := http.NewRequest("POST", uploadParams["host"].(string), &b)
				if err != nil {
					failedNotices = append(failedNotices, fmt.Sprintf("Failed to create upload request for notice ID %d: %v", oldNotice.ID, err))
					continue
				}
				uploadReq.Header.Set("Content-Type", w.FormDataContentType())

				client := &http.Client{}
				uploadResp, err := client.Do(uploadReq)
				if err != nil {
					failedNotices = append(failedNotices, fmt.Sprintf("Failed to upload file for notice ID %d: %v", oldNotice.ID, err))
					continue
				}

				respBody, _ := io.ReadAll(uploadResp.Body)
				uploadResp.Body.Close()

				if uploadResp.StatusCode != http.StatusOK {
					failedNotices = append(failedNotices, fmt.Sprintf("File upload failed for notice ID %d: status code %d, response: %s", oldNotice.ID, uploadResp.StatusCode, string(respBody)))
					continue
				}
			}
		}

		if shouldAddFile {
			if err := c.Container.GetService("file").(base_services.InterfaceFileService).Create(fileForNotice); err != nil {
				failedNotices = append(failedNotices, fmt.Sprintf("Failed to create file record for notice ID %d: %v", oldNotice.ID, err))
				continue
			}
		}

		// Determine notice type
		noticeType := field.NoticeTypeNormal
		switch oldNotice.MessType {
		case string(field.NoticeOldTypeCommon):
			noticeType = field.NoticeTypeNormal
		case string(field.NoticeOldTypeIo):
			noticeType = field.NoticeTypeBuilding
		case string(field.NoticeOldTypeUrgent):
			noticeType = field.NoticeTypeUrgent
		case string(field.NoticeOldTypeGovernment):
			noticeType = field.NoticeTypeGovernment
		}

		// Create new notice
		notice := &base_models.Notice{
			Title:          oldNotice.MessTitle,
			Description:    oldNotice.MessTitle,
			Type:           noticeType,
			Status:         field.Status("active"),
			StartTime:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.FixedZone("CST", 8*3600)),
			EndTime:        time.Date(2100, 2, 1, 0, 0, 0, 0, time.FixedZone("CST", 8*3600)),
			IsPublic:       true,
			IsIsmartNotice: true, // Set this flag for notices from old system
			FileID:         &fileForNotice.ID,
			FileType:       field.FileTypePdf,
		}

		// Create notice and bind to building
		err = databases.DB_CONN.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(notice).Error; err != nil {
				return fmt.Errorf("failed to create notice record: %v", err)
			}

			if err := tx.Exec("INSERT INTO notice_buildings (notice_id, building_id) VALUES (?, ?)", notice.ID, id).Error; err != nil {
				return fmt.Errorf("failed to bind notice to building: %v", err)
			}

			return nil
		})
		if err != nil {
			failedNotices = append(failedNotices, fmt.Sprintf("Failed to process notice ID %d: %v", oldNotice.ID, err))
			continue
		}

		successCount++
	}

	// Return sync results
	c.Ctx.JSON(200, gin.H{
		"message":        "Sync completed",
		"successCount":   successCount,
		"hasSyncedCount": hasSyncedCount,
		"deleteCount":    deleteCount,
		"failedNotices":  failedNotices,
		"totalProcessed": len(respBody),
	})
}

// ManualSyncNotice handles manual notice synchronization for a building
func (c *BuildingController) ManualSyncNotice() {
	id, err := strconv.ParseUint(c.Ctx.Param("id"), 10, 64)
	if err != nil {
		c.Ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid building ID"})
		return
	}

	claims := c.Ctx.MustGet("claims").(jwt.MapClaims)
	result, err := c.Container.GetService("noticeSync").(base_services.InterfaceNoticeSyncService).ManualSyncBuildingNotices(uint(id), claims)
	if err != nil {
		c.Ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Ctx.JSON(http.StatusOK, result)
}

package base_services

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	"github.com/The-Healthist/iboard_http_service/utils/field"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// processResult represents the result of processing a single notice
type processResult struct {
	noticeID     int
	md5          string
	success      bool
	syncRequired bool
	error        error
}

const (
	minWorkers       = 10 // increase min worker count
	maxWorkers       = 30 // increase max worker count
	noticesPerWorker = 3  // decrease notices per worker for better concurrency

	// Building concurrency control
	minBuildingWorkers       = 1    // min building worker
	maxBuildingWorkerPerCore = 20   // max building worker per core
	buildingWorkerLoadFactor = 0.75 // building worker load factor (0-1)
)

func getNoticeSyncInterval() time.Duration {
	interval := os.Getenv("NOTICE_SYNC_INTERVAL")
	if interval == "" {
		return 2 * time.Minute // default to 2 minutes if not set
	}

	intervalInt, err := strconv.Atoi(interval)
	if err != nil {
		return 2 * time.Minute // default to 2 minutes if invalid value
	}

	return time.Duration(intervalInt) * time.Minute
}

// notice list redis cache duration
func getNoticeSyncCacheDuration() time.Duration {
	duration := os.Getenv("NOTICE_SYNC_CACHE_DURATION")
	if duration == "" {
		return 4 * time.Minute // default to 4 minutes if not set
	}

	durationInt, err := strconv.Atoi(duration)
	if err != nil {
		return 4 * time.Minute // default to 4 minutes if invalid value
	}

	return time.Duration(durationInt) * time.Minute
}

// building redis
func getNoticeSyncBuildingCacheDuration() time.Duration {
	duration := os.Getenv("NOTICE_SYNC_BUILDING_CACHE_DURATION")
	if duration == "" {
		return 30 * time.Minute // default to 30 minutes if not set
	}

	durationInt, err := strconv.Atoi(duration)
	if err != nil {
		return 30 * time.Minute // default to 30 minutes if invalid value
	}

	return time.Duration(durationInt) * time.Minute
}

func getNoticeSyncCountCacheDuration() time.Duration {
	duration := os.Getenv("NOTICE_SYNC_COUNT_CACHE_DURATION")
	if duration == "" {
		return 60 * time.Minute // default to 60 minutes if not set
	}

	durationInt, err := strconv.Atoi(duration)
	if err != nil {
		return 60 * time.Minute // default to 60 minutes if invalid value
	}

	return time.Duration(durationInt) * time.Minute
}

// calculateWorkerCount dynamically calculate worker count based on notice count
func calculateWorkerCount(noticeCount int) int {
	// Calculate suggested worker count based on notice count and available CPU cores
	cpuCount := runtime.NumCPU()
	suggestedWorkers := noticeCount / noticesPerWorker

	// Scale based on CPU cores
	maxWorkersForCPU := cpuCount * 4

	if suggestedWorkers < minWorkers {
		return minWorkers
	}
	if suggestedWorkers > maxWorkers || suggestedWorkers > maxWorkersForCPU {
		if maxWorkers < maxWorkersForCPU {
			return maxWorkers
		}
		return maxWorkersForCPU
	}
	return suggestedWorkers
}

// Add new function for calculating building workers
func calculateBuildingWorkers(buildingCount int, cpuCores int) int {
	// Calculate max concurrency based on CPU cores
	maxWorkers := int(float64(cpuCores) * maxBuildingWorkerPerCore * buildingWorkerLoadFactor)

	// Ensure at least minBuildingWorkers
	if maxWorkers < minBuildingWorkers {
		maxWorkers = minBuildingWorkers
	}

	// If building count is less than max workers, use building count
	if buildingCount < maxWorkers {
		return buildingCount
	}

	return maxWorkers
}

type InterfaceNoticeSyncService interface {
	SyncBuildingNotices(buildingID uint, claims jwt.MapClaims) (gin.H, error)
	StartSyncScheduler(ctx context.Context)
	ManualSyncBuildingNotices(buildingID uint, claims jwt.MapClaims) (gin.H, error)
}

type NoticeSyncService struct {
	db              *gorm.DB
	redis           *redis.Client
	buildingService InterfaceBuildingService
	uploadService   IUploadService
	fileService     InterfaceFileService
}

func NewNoticeSyncService(db *gorm.DB, redis *redis.Client, buildingService InterfaceBuildingService, uploadService IUploadService, fileService InterfaceFileService) InterfaceNoticeSyncService {
	return &NoticeSyncService{
		db:              db,
		redis:           redis,
		buildingService: buildingService,
		uploadService:   uploadService,
		fileService:     fileService,
	}
}

type OldSystemNotice struct {
	ID        int    `json:"id"`
	MessTitle string `json:"mess_title"`
	MessType  string `json:"mess_type"`
	MessFile  string `json:"mess_file"`
}

// 1. getCachedNoticeIDs
func (s *NoticeSyncService) getCachedNoticeIDs(ctx context.Context, buildingID uint) ([]int, error) {
	key := fmt.Sprintf("building_notice_ids:%d", buildingID)
	val, err := s.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var ids []int
	if err := json.Unmarshal([]byte(val), &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// 2. setCachedNoticeIDs
func (s *NoticeSyncService) setCachedNoticeIDs(ctx context.Context, buildingID uint, ids []int) error {
	key := fmt.Sprintf("building_notice_ids:%d", buildingID)
	jsonIDs, err := json.Marshal(ids)
	if err != nil {
		return err
	}
	return s.redis.Set(ctx, key, string(jsonIDs), getNoticeSyncCacheDuration()).Err()
}

// 3. getCachedBuilding
func (s *NoticeSyncService) getCachedBuilding(ctx context.Context, buildingID uint) (*base_models.Building, error) {
	key := fmt.Sprintf("building:%d", buildingID)
	val, err := s.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		// Cache miss, get from DB and cache it
		var building base_models.Building
		if err := s.db.First(&building, buildingID).Error; err != nil {
			return nil, err
		}

		// Cache building data
		jsonData, err := json.Marshal(building)
		if err != nil {
			return nil, err
		}
		if err := s.redis.Set(ctx, key, string(jsonData), getNoticeSyncBuildingCacheDuration()).Err(); err != nil {
			return nil, err
		}
		return &building, nil
	} else if err != nil {
		return nil, err
	}

	var building base_models.Building
	if err := json.Unmarshal([]byte(val), &building); err != nil {
		return nil, err
	}
	return &building, nil
}

// 4. getCachedNoticeCount
func (s *NoticeSyncService) getCachedNoticeCount(ctx context.Context, buildingID uint) (int64, error) {
	key := fmt.Sprintf("building_ismart_notice_count:%d", buildingID)
	val, err := s.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		// Cache miss, get from DB and cache it
		var totalIsmartNotices int64
		if err := s.db.Model(&base_models.Notice{}).
			Joins("JOIN notice_buildings ON notices.id = notice_buildings.notice_id").
			Where("notice_buildings.building_id = ? AND notices.is_ismart_notice = ?", buildingID, true).
			Count(&totalIsmartNotices).Error; err != nil {
			return 0, fmt.Errorf("failed to get total ismart notices: %v", err)
		}

		// Cache the count for 1 hour
		if err := s.redis.Set(ctx, key, totalIsmartNotices, time.Hour).Err(); err != nil {
			return 0, err
		}
		return totalIsmartNotices, nil
	} else if err != nil {
		return 0, err
	}

	count, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// 5. updateCachedNoticeCount
func (s *NoticeSyncService) updateCachedNoticeCount(ctx context.Context, buildingID uint, count int64) error {
	key := fmt.Sprintf("building_ismart_notice_count:%d", buildingID)
	return s.redis.Set(ctx, key, count, getNoticeSyncCountCacheDuration()).Err()
}

// 6. SyncBuildingNotices
func (s *NoticeSyncService) SyncBuildingNotices(buildingID uint, claims jwt.MapClaims) (gin.H, error) {
	ctx := context.Background()

	// Get building info from cache first
	building, err := s.getCachedBuilding(ctx, buildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get building: %v", err)
	}

	// Get cached notice IDs and log them
	cachedIDs, err := s.getCachedNoticeIDs(ctx, buildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cached notice IDs: %v", err)
	}
	fmt.Printf("[NoticeSyncService] Redis cached IDs for building %d: %v\n", buildingID, cachedIDs)

	// Get existing iSmart notices with a single query including all needed relations
	var existingNotices []base_models.Notice
	if err := s.db.Joins("JOIN notice_buildings ON notices.id = notice_buildings.notice_id").
		Where("notice_buildings.building_id = ? AND notices.is_ismart_notice = ?", buildingID, true).
		Preload("File").Find(&existingNotices).Error; err != nil {
		return nil, fmt.Errorf("failed to get existing notices: %v", err)
	}
	fmt.Printf("[NoticeSyncService] Found %d existing notices in database\n", len(existingNotices))

	// Request notices from old system
	url := "https://uqf0jqfm77.execute-api.ap-east-1.amazonaws.com/prod/v1/building_board/building-notices"
	reqBody := struct {
		BlgID string `json:"blg_id"`
	}{
		BlgID: building.IsmartID,
	}

	reqBodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBodyJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to request old system: %v", err)
	}
	defer resp.Body.Close()

	var oldNotices []OldSystemNotice
	if err := json.NewDecoder(resp.Body).Decode(&oldNotices); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	// Extract and log new notice IDs from old system
	newIDs := make([]int, len(oldNotices))
	for i, notice := range oldNotices {
		newIDs[i] = notice.ID
	}
	fmt.Printf("[NoticeSyncService] Old system IDs for building %d: %v\n", buildingID, newIDs)

	// Check if we need to force sync due to notice count mismatch
	totalIsmartNotices, err := s.getCachedNoticeCount(ctx, buildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cached notice count: %v", err)
	}

	fmt.Printf("[NoticeSyncService] Database has %d ismart notices, old system has %d notices\n",
		totalIsmartNotices, len(oldNotices))

	// Force sync if notice counts don't match
	if totalIsmartNotices != int64(len(oldNotices)) {
		fmt.Printf("[NoticeSyncService] Notice count mismatch detected, forcing full sync\n")
		// Clear caches to force sync
		if err := s.redis.Del(ctx, fmt.Sprintf("building_notice_ids:%d", buildingID)).Err(); err != nil {
			return nil, fmt.Errorf("failed to clear notice IDs cache: %v", err)
		}
		if err := s.redis.Del(ctx, fmt.Sprintf("building_ismart_notice_count:%d", buildingID)).Err(); err != nil {
			return nil, fmt.Errorf("failed to clear notice count cache: %v", err)
		}
		cachedIDs = nil
	}

	// If cached IDs exist and match new IDs, skip processing
	if len(cachedIDs) > 0 {
		match := len(cachedIDs) == len(newIDs)
		if match {
			for i := range cachedIDs {
				if cachedIDs[i] != newIDs[i] {
					match = false
					break
				}
			}
			if match && totalIsmartNotices == int64(len(oldNotices)) {
				fmt.Printf("[NoticeSyncService] No changes detected for building %d, skipping sync\n", buildingID)
				var hasSyncNotices []int
				for _, notice := range existingNotices {
					if notice.File != nil {
						hasSyncNotices = append(hasSyncNotices, int(notice.ID))
					}
				}
				return gin.H{
					"message":           "No changes detected",
					"successCount":      0,
					"hasSyncedCount":    len(cachedIDs),
					"deleteCount":       0,
					"failedNotices":     []string{},
					"totalProcessed":    len(oldNotices),
					"has_sync_notices":  hasSyncNotices,
					"need_sync_notices": []int{},
					"change_notices":    []int{},
				}, nil
			}
		}
	}

	// Create MD5 map for existing notices
	existingMD5Map := make(map[string]bool)
	existingMD5ToNoticeID := make(map[string]uint)
	for _, notice := range existingNotices {
		if notice.File != nil {
			existingMD5Map[notice.File.Md5] = true
			existingMD5ToNoticeID[notice.File.Md5] = notice.ID
		}
	}

	// Calculate optimal worker count
	workers := calculateWorkerCount(len(oldNotices))
	resultChan := make(chan processResult, len(oldNotices))
	semaphore := make(chan struct{}, workers)

	fmt.Printf("[NoticeSyncService] Starting concurrent processing with %d workers for %d notices\n",
		workers, len(oldNotices))

	// Process notices concurrently
	for _, oldNotice := range oldNotices {
		go func(notice OldSystemNotice) {
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			result := processResult{noticeID: notice.ID}

			// Download and process file
			fileResp, err := http.Get(notice.MessFile)
			if err != nil {
				result.error = fmt.Errorf("failed to download file: %v", err)
				resultChan <- result
				return
			}
			defer fileResp.Body.Close()

			fileContent, err := io.ReadAll(fileResp.Body)
			if err != nil {
				result.error = fmt.Errorf("failed to read file content: %v", err)
				resultChan <- result
				return
			}

			md5Hash := md5.Sum(fileContent)
			md5Str := hex.EncodeToString(md5Hash[:])
			result.md5 = md5Str

			// Check if file exists
			if existingMD5Map[md5Str] {
				result.success = true
				result.syncRequired = false
			} else {
				result.syncRequired = true
				if err := s.processNotice(buildingID, notice, fileContent, claims); err != nil {
					result.error = err
				} else {
					result.success = true
				}
			}

			resultChan <- result
		}(oldNotice)
	}

	// Collect results
	var successCount, hasSyncedCount int
	var failedNotices []string
	var needSyncNotices, hasSyncNotices, changeNotices []int
	processedMD5s := make(map[string]bool)

	for i := 0; i < len(oldNotices); i++ {
		result := <-resultChan
		if result.error != nil {
			failedNotices = append(failedNotices, fmt.Sprintf("Failed to process notice ID %d: %v", result.noticeID, result.error))
			continue
		}

		processedMD5s[result.md5] = true
		if result.syncRequired {
			if result.success {
				successCount++
				needSyncNotices = append(needSyncNotices, result.noticeID)
			}
		} else {
			hasSyncedCount++
			if id, ok := existingMD5ToNoticeID[result.md5]; ok {
				hasSyncNotices = append(hasSyncNotices, int(id))
			}
		}
	}

	fmt.Printf("[NoticeSyncService] Concurrent processing completed for building %d\n", buildingID)

	// Process deletions - Always check for notices to clean up
	deleteCount := 0
	deleteCount, err = s.processDeletedNotices(buildingID, existingNotices, processedMD5s, &failedNotices)
	if err != nil {
		return nil, fmt.Errorf("failed to process deletions: %v", err)
	}
	fmt.Printf("[NoticeSyncService] Processed deletions for building %d: %d notices deleted\n", buildingID, deleteCount)

	// Update cache with new IDs
	if err := s.setCachedNoticeIDs(ctx, buildingID, newIDs); err != nil {
		return nil, fmt.Errorf("failed to update cached notice IDs: %v", err)
	}

	// Update cached notice count after successful sync
	newCount := int64(len(oldNotices))
	if err := s.updateCachedNoticeCount(ctx, buildingID, newCount); err != nil {
		fmt.Printf("[NoticeSyncService] Warning: failed to update cached notice count: %v\n", err)
	}

	return gin.H{
		"message":           "Sync completed",
		"successCount":      successCount,
		"hasSyncedCount":    hasSyncedCount,
		"deleteCount":       deleteCount,
		"failedNotices":     failedNotices,
		"totalProcessed":    len(oldNotices),
		"has_sync_notices":  hasSyncNotices,
		"need_sync_notices": needSyncNotices,
		"change_notices":    changeNotices,
	}, nil
}

// processNotice handles the processing of a single notice
func (s *NoticeSyncService) processNotice(buildingID uint, oldNotice OldSystemNotice, fileContent []byte, claims jwt.MapClaims) error {
	fileSize := len(fileContent)
	md5Hash := md5.Sum(fileContent)
	md5Str := hex.EncodeToString(md5Hash[:])
	mimeType := http.DetectContentType(fileContent)

	// Get uploader info from claims
	var uploaderID uint
	var uploaderType field.FileUploaderType
	var uploaderEmail string

	if isAdmin, ok := claims["isAdmin"].(bool); ok && isAdmin {
		if id, ok := claims["id"].(float64); ok {
			uploaderID = uint(id)
			uploaderType = field.UploaderTypeSuperAdmin
			if email, ok := claims["email"].(string); ok {
				uploaderEmail = email
			}
		}
	}

	if uploaderID == 0 {
		uploaderID = 1 // Use default system admin ID if not set
		uploaderType = field.UploaderTypeSuperAdmin
		uploaderEmail = "admin@example.com"
	}

	// Check if file exists
	var existingFile base_models.File
	var shouldUpload bool = true
	var fileForNotice *base_models.File

	err := s.db.Where("md5 = ?", md5Str).First(&existingFile).Error
	if err == nil {
		// File exists, check if it's associated with any notice
		var existingNotice base_models.Notice
		err := s.db.Where("file_id = ?", existingFile.ID).First(&existingNotice).Error

		if err == nil {
			// Check if notice is bound to current building
			var count int64
			err = s.db.Table("notice_buildings").
				Where("notice_id = ? AND building_id = ?", existingNotice.ID, buildingID).
				Count(&count).Error

			if err == nil && count > 0 {
				return nil
			}

			// Check if notice is bound to other buildings
			err = s.db.Table("notice_buildings").
				Where("notice_id = ?", existingNotice.ID).
				Count(&count).Error

			if err == nil {
				if count > 0 {
					shouldUpload = false
					fileForNotice = &existingFile
				} else {
					if err := s.db.Delete(&existingNotice).Error; err != nil {
						return fmt.Errorf("failed to delete old notice: %v", err)
					}
					shouldUpload = false
					fileForNotice = &existingFile
				}
			}
		} else {
			shouldUpload = false
			fileForNotice = &existingFile
		}
	} else if err == gorm.ErrRecordNotFound {
		shouldUpload = true

		// Generate unique file name using UUID
		currentTime := time.Now()
		dir := currentTime.Format("2006-01-02") + "/"
		uuid := uuid.New().String()
		fileName := uuid + ".pdf"
		objectKey := dir + fileName

		// Get OSS configuration from environment variables
		host := os.Getenv("HOST")
		if host == "" {
			host = "http://idreamsky.oss-cn-beijing.aliyuncs.com"
		}

		// Create file record with OSS path
		fileForNotice = &base_models.File{
			Path:         host + "/" + objectKey, // Store the complete OSS URL
			Size:         int64(fileSize),
			MimeType:     mimeType,
			Oss:          "aliyun",
			UploaderType: uploaderType,
			UploaderID:   uploaderID,
			Uploader:     uploaderEmail,
			Md5:          md5Str,
		}

		if shouldUpload {
			// Get upload parameters from service
			uploadParams, err := s.uploadService.GetUploadParamsSync(objectKey)
			if err != nil {
				return fmt.Errorf("failed to get upload params: %v", err)
			}

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
					return fmt.Errorf("failed to write form field %s: %v", field.key, err)
				}
			}

			fw, err := w.CreateFormFile("file", objectKey)
			if err != nil {
				return fmt.Errorf("failed to create form file: %v", err)
			}
			if _, err = io.Copy(fw, bytes.NewReader(fileContent)); err != nil {
				return fmt.Errorf("failed to copy file content: %v", err)
			}
			w.Close()

			uploadReq, err := http.NewRequest("POST", uploadParams["host"].(string), &b)
			if err != nil {
				return fmt.Errorf("failed to create upload request: %v", err)
			}
			uploadReq.Header.Set("Content-Type", w.FormDataContentType())

			client := &http.Client{}
			uploadResp, err := client.Do(uploadReq)
			if err != nil {
				return fmt.Errorf("failed to upload file: %v", err)
			}
			defer uploadResp.Body.Close()

			if uploadResp.StatusCode != http.StatusOK {
				respBody, _ := io.ReadAll(uploadResp.Body)
				return fmt.Errorf("upload failed with status %d: %s", uploadResp.StatusCode, string(respBody))
			}
		}

		// Create file record
		if err := s.fileService.Create(fileForNotice); err != nil {
			return fmt.Errorf("failed to create file record: %v", err)
		}
	}

	// Create notice
	notice := &base_models.Notice{
		Title:          oldNotice.MessTitle,
		Description:    oldNotice.MessTitle,
		Type:           s.mapNoticeType(oldNotice.MessType),
		Status:         field.Status("active"),
		StartTime:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		EndTime:        time.Date(2100, 2, 1, 0, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		IsPublic:       true,
		IsIsmartNotice: true,
		FileID:         &fileForNotice.ID,
		FileType:       field.FileTypePdf,
	}

	// Create notice and bind to building in a transaction
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(notice).Error; err != nil {
			return fmt.Errorf("failed to create notice: %v", err)
		}

		if err := tx.Exec("INSERT INTO notice_buildings (notice_id, building_id) VALUES (?, ?)",
			notice.ID, buildingID).Error; err != nil {
			return fmt.Errorf("failed to bind notice to building: %v", err)
		}

		return nil
	})
}

// processDeletedNotices handles removing notices that are no longer in the old system
func (s *NoticeSyncService) processDeletedNotices(buildingID uint, existingNotices []base_models.Notice, processedMD5s map[string]bool, failedNotices *[]string) (int, error) {
	deleteCount := 0
	fmt.Printf("[NoticeSyncService] Starting to process deletions for building %d, checking %d existing notices\n",
		buildingID, len(existingNotices))

	// Filter out notices that don't have files or aren't iSmart notices
	var noticesToCheck []base_models.Notice
	for _, notice := range existingNotices {
		if notice.File != nil && notice.IsIsmartNotice {
			noticesToCheck = append(noticesToCheck, notice)
		}
	}

	fmt.Printf("[NoticeSyncService] Found %d iSmart notices with files to check for building %d\n",
		len(noticesToCheck), buildingID)

	for _, existingNotice := range noticesToCheck {
		// If the notice's MD5 is not in the processed list, it means this notice is not in the old system anymore
		if !processedMD5s[existingNotice.File.Md5] {
			fmt.Printf("[NoticeSyncService] Notice ID %d with MD5 %s not found in old system, will be cleaned up\n",
				existingNotice.ID, existingNotice.File.Md5)

			// Start transaction for deletion
			tx := s.db.Begin()

			// Unbind notice from building
			if err := tx.Exec("DELETE FROM notice_buildings WHERE notice_id = ? AND building_id = ?",
				existingNotice.ID, buildingID).Error; err != nil {
				tx.Rollback()
				*failedNotices = append(*failedNotices, fmt.Sprintf("Failed to unbind notice ID %d: %v", existingNotice.ID, err))
				continue
			}

			fmt.Printf("[NoticeSyncService] Unbound notice ID %d from building %d\n", existingNotice.ID, buildingID)

			// Check if notice is bound to other buildings
			var buildingCount int64
			if err := tx.Table("notice_buildings").Where("notice_id = ?", existingNotice.ID).Count(&buildingCount).Error; err != nil {
				tx.Rollback()
				*failedNotices = append(*failedNotices, fmt.Sprintf("Failed to check notice bindings for ID %d: %v", existingNotice.ID, err))
				continue
			}

			if buildingCount == 0 {
				fmt.Printf("[NoticeSyncService] Notice ID %d has no other building bindings, will be deleted\n", existingNotice.ID)

				// Save file ID for later check
				fileID := existingNotice.FileID

				// Delete notice if no other bindings exist
				if err := tx.Delete(&existingNotice).Error; err != nil {
					tx.Rollback()
					*failedNotices = append(*failedNotices, fmt.Sprintf("Failed to delete notice ID %d: %v", existingNotice.ID, err))
					continue
				}

				fmt.Printf("[NoticeSyncService] Deleted notice ID %d\n", existingNotice.ID)

				// Check if file is used by other notices
				if fileID != nil {
					var fileCount int64
					if err := tx.Model(&base_models.Notice{}).Where("file_id = ?", *fileID).Count(&fileCount).Error; err != nil {
						tx.Rollback()
						*failedNotices = append(*failedNotices, fmt.Sprintf("Failed to check file references for notice ID %d: %v", existingNotice.ID, err))
						continue
					}

					if fileCount == 0 {
						fmt.Printf("[NoticeSyncService] File ID %d has no other notice references, will be deleted\n", *fileID)

						// Delete file if no other notices use it
						if err := tx.Delete(&base_models.File{}, *fileID).Error; err != nil {
							tx.Rollback()
							*failedNotices = append(*failedNotices, fmt.Sprintf("Failed to delete file for notice ID %d: %v", existingNotice.ID, err))
							continue
						}

						fmt.Printf("[NoticeSyncService] Deleted file ID %d\n", *fileID)
					} else {
						fmt.Printf("[NoticeSyncService] File ID %d has %d other notice references, keeping file\n", *fileID, fileCount)
					}
				}
			} else {
				fmt.Printf("[NoticeSyncService] Notice ID %d has %d other building bindings, keeping notice\n",
					existingNotice.ID, buildingCount)
			}

			if err := tx.Commit().Error; err != nil {
				*failedNotices = append(*failedNotices, fmt.Sprintf("Failed to commit deletion for notice ID %d: %v", existingNotice.ID, err))
				continue
			}

			deleteCount++
		}
	}

	fmt.Printf("[NoticeSyncService] Deletion process completed for building %d: %d notices deleted\n",
		buildingID, deleteCount)
	return deleteCount, nil
}

// 9. mapNoticeType
func (s *NoticeSyncService) mapNoticeType(oldType string) field.NoticeType {
	switch oldType {
	case string(field.NoticeOldTypeCommon):
		return field.NoticeTypeNormal
	case string(field.NoticeOldTypeIo):
		return field.NoticeTypeBuilding
	case string(field.NoticeOldTypeUrgent):
		return field.NoticeTypeUrgent
	case string(field.NoticeOldTypeGovernment):
		return field.NoticeTypeGovernment
	default:
		return field.NoticeTypeNormal
	}
}

// Add new function to clear all notice sync related caches
func (s *NoticeSyncService) clearAllCaches(ctx context.Context) error {
	// Get all buildings
	var buildings []base_models.Building
	if err := s.db.Find(&buildings).Error; err != nil {
		return fmt.Errorf("failed to get buildings: %v", err)
	}

	// Clear caches for each building
	for _, building := range buildings {
		// Clear notice IDs cache
		idKey := fmt.Sprintf("building_notice_ids:%d", building.ID)
		if err := s.redis.Del(ctx, idKey).Err(); err != nil {
			fmt.Printf("[NoticeSyncService] Warning: failed to clear notice IDs cache for building %d: %v\n", building.ID, err)
		}

		// Clear notice count cache
		countKey := fmt.Sprintf("building_ismart_notice_count:%d", building.ID)
		if err := s.redis.Del(ctx, countKey).Err(); err != nil {
			fmt.Printf("[NoticeSyncService] Warning: failed to clear notice count cache for building %d: %v\n", building.ID, err)
		}

		// Clear building cache
		buildingKey := fmt.Sprintf("building:%d", building.ID)
		if err := s.redis.Del(ctx, buildingKey).Err(); err != nil {
			fmt.Printf("[NoticeSyncService] Warning: failed to clear building cache for building %d: %v\n", building.ID, err)
		}
	}

	fmt.Println("[NoticeSyncService] All caches cleared successfully")
	return nil
}

// 10. StartSyncScheduler
func (s *NoticeSyncService) StartSyncScheduler(ctx context.Context) {
	syncInterval := getNoticeSyncInterval()
	ticker := time.NewTicker(syncInterval)
	countdownTicker := time.NewTicker(1 * time.Minute)
	fmt.Println("[NoticeSyncService] Scheduler started")

	// Clear all caches before starting
	if err := s.clearAllCaches(ctx); err != nil {
		fmt.Printf("[NoticeSyncService] Warning: failed to clear caches on start: %v\n", err)
	}

	go func() {
		// Create admin claims for sync
		adminClaims := jwt.MapClaims{
			"id":      float64(1),
			"isAdmin": true,
			"email":   "admin@example.com",
		}

		// 立即执行一次同步
		fmt.Println("[NoticeSyncService] Running initial sync...")
		s.runSync(adminClaims)
		nextSync := time.Now().Add(syncInterval)

		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				countdownTicker.Stop()
				fmt.Println("[NoticeSyncService] Scheduler stopped")
				return
			case <-countdownTicker.C:
				remaining := time.Until(nextSync).Round(time.Minute)
				fmt.Printf("[NoticeSyncService] Next sync in %d minutes\n", int(remaining.Minutes()))
			case <-ticker.C:
				fmt.Println("[NoticeSyncService] Running scheduled sync...")
				s.runSync(adminClaims)
				nextSync = time.Now().Add(syncInterval)
			}
		}
	}()
}

// 11. runSync
func (s *NoticeSyncService) runSync(adminClaims jwt.MapClaims) {
	// Get all buildings
	var buildings []base_models.Building
	if err := s.db.Find(&buildings).Error; err != nil {
		fmt.Printf("[NoticeSyncService] Failed to get buildings: %v\n", err)
		return
	}

	buildingCount := len(buildings)
	fmt.Printf("[NoticeSyncService] Found %d buildings to sync\n", buildingCount)

	// Get number of CPU cores
	cpuCores := runtime.NumCPU()

	// Calculate optimal number of building workers based on CPU cores
	maxBuildingWorkers := calculateBuildingWorkers(buildingCount, cpuCores)

	fmt.Printf("[NoticeSyncService] System has %d CPU cores, using %d concurrent building workers\n",
		cpuCores, maxBuildingWorkers)

	// Create a wait group to wait for all goroutines to finish
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxBuildingWorkers)

	// Create a channel for collecting results
	resultChan := make(chan struct {
		buildingID uint
		name       string
		result     gin.H
		err        error
	}, buildingCount)

	// Start sync for each building concurrently
	for _, building := range buildings {
		wg.Add(1)
		go func(b base_models.Building) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			fmt.Printf("[NoticeSyncService] Starting sync for building %d (%s)\n", b.ID, b.Name)
			result, err := s.SyncBuildingNotices(b.ID, adminClaims)

			// Send result to channel
			resultChan <- struct {
				buildingID uint
				name       string
				result     gin.H
				err        error
			}{
				buildingID: b.ID,
				name:       b.Name,
				result:     result,
				err:        err,
			}
		}(building)
	}

	// Start a goroutine to close result channel when all syncs are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Process results as they come in
	for res := range resultChan {
		if res.err != nil {
			fmt.Printf("[NoticeSyncService] Failed to sync notices for building %d (%s): %v\n",
				res.buildingID, res.name, res.err)
			continue
		}

		// Consolidated statistics output
		fmt.Printf("[NoticeSyncService] Sync completed for building %d (%s):\n", res.buildingID, res.name)
		fmt.Printf("[NoticeSyncService]     - Total to process: %d\n", res.result["totalProcessed"])
		fmt.Printf("[NoticeSyncService]     - Successfully processed: %d\n", res.result["successCount"])
		fmt.Printf("[NoticeSyncService]     - Already synced: %d\n", res.result["hasSyncedCount"])
		fmt.Printf("[NoticeSyncService]     - Failed: %d\n", len(res.result["failedNotices"].([]string)))
		fmt.Printf("[NoticeSyncService]     - To be deleted: %d\n", res.result["deleteCount"])

		// Only show detailed IDs if there are changes
		needSync := res.result["need_sync_notices"].([]int)
		hasSync := res.result["has_sync_notices"].([]int)
		changes := res.result["change_notices"].([]int)

		if len(needSync) > 0 {
			fmt.Printf("[NoticeSyncService]     - Need sync: %v\n", needSync)
		}
		if len(hasSync) > 0 {
			fmt.Printf("[NoticeSyncService]     - Has sync: %v\n", hasSync)
		}
		if len(changes) > 0 {
			fmt.Printf("[NoticeSyncService]     - Changes: %v\n", changes)
		}

		// Only show failed notices if there are any
		failedNotices := res.result["failedNotices"].([]string)
		if len(failedNotices) > 0 {
			fmt.Printf("[NoticeSyncService]     - Failed notices:\n")
			for _, notice := range failedNotices {
				fmt.Printf("[NoticeSyncService]         * %s\n", notice)
			}
		}
		fmt.Println() // Add a blank line between buildings
	}
}

// ManualSyncBuildingNotices performs a manual sync without cache checking
func (s *NoticeSyncService) ManualSyncBuildingNotices(buildingID uint, claims jwt.MapClaims) (gin.H, error) {
	// Get building info
	building, err := s.buildingService.GetByID(buildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get building: %v", err)
	}

	// Get existing notices
	var existingNotices []base_models.Notice
	if err := s.db.Joins("JOIN notice_buildings ON notices.id = notice_buildings.notice_id").
		Where("notice_buildings.building_id = ? AND notices.is_ismart_notice = ?", buildingID, true).
		Preload("File").Find(&existingNotices).Error; err != nil {
		return nil, fmt.Errorf("failed to get existing notices: %v", err)
	}

	// Request notices from old system
	url := "https://uqf0jqfm77.execute-api.ap-east-1.amazonaws.com/prod/v1/building_board/building-notices"
	reqBody := struct {
		BlgID string `json:"blg_id"`
	}{
		BlgID: building.IsmartID,
	}

	reqBodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBodyJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to request old system: %v", err)
	}
	defer resp.Body.Close()

	var oldNotices []OldSystemNotice
	if err := json.NewDecoder(resp.Body).Decode(&oldNotices); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	// Create MD5 map for existing notices
	existingMD5Map := make(map[string]bool)
	existingMD5ToNoticeID := make(map[string]uint)
	for _, notice := range existingNotices {
		if notice.File != nil {
			existingMD5Map[notice.File.Md5] = true
			existingMD5ToNoticeID[notice.File.Md5] = notice.ID
		}
	}

	// Calculate optimal worker count
	workers := calculateWorkerCount(len(oldNotices))
	resultChan := make(chan processResult, len(oldNotices))
	semaphore := make(chan struct{}, workers)

	fmt.Printf("[NoticeSyncService] Starting concurrent processing with %d workers for %d notices\n",
		workers, len(oldNotices))

	// Process notices concurrently
	for _, oldNotice := range oldNotices {
		go func(notice OldSystemNotice) {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result := processResult{noticeID: notice.ID}

			// Download and process file
			fileResp, err := http.Get(notice.MessFile)
			if err != nil {
				result.error = fmt.Errorf("failed to download file: %v", err)
				resultChan <- result
				return
			}
			defer fileResp.Body.Close()

			fileContent, err := io.ReadAll(fileResp.Body)
			if err != nil {
				result.error = fmt.Errorf("failed to read file content: %v", err)
				resultChan <- result
				return
			}

			md5Hash := md5.Sum(fileContent)
			md5Str := hex.EncodeToString(md5Hash[:])
			result.md5 = md5Str

			// Check if file exists
			if existingMD5Map[md5Str] {
				result.success = true
				result.syncRequired = false
			} else {
				result.syncRequired = true
				if err := s.processNotice(buildingID, notice, fileContent, claims); err != nil {
					result.error = err
				} else {
					result.success = true
				}
			}

			resultChan <- result
		}(oldNotice)
	}

	// Collect results
	var successCount, hasSyncedCount int
	var failedNotices []string
	var needSyncNotices, hasSyncNotices, changeNotices []int
	processedMD5s := make(map[string]bool)

	for i := 0; i < len(oldNotices); i++ {
		result := <-resultChan
		if result.error != nil {
			failedNotices = append(failedNotices, fmt.Sprintf("Failed to process notice ID %d: %v", result.noticeID, result.error))
			continue
		}

		processedMD5s[result.md5] = true
		if result.syncRequired {
			if result.success {
				successCount++
				needSyncNotices = append(needSyncNotices, result.noticeID)
			}
		} else {
			hasSyncedCount++
			if id, ok := existingMD5ToNoticeID[result.md5]; ok {
				hasSyncNotices = append(hasSyncNotices, int(id))
			}
		}
	}

	// Process deletions
	deleteCount := 0
	if len(changeNotices) > 0 {
		var err error
		deleteCount, err = s.processDeletedNotices(buildingID, existingNotices, processedMD5s, &failedNotices)
		if err != nil {
			return nil, fmt.Errorf("failed to process deletions: %v", err)
		}
	}

	return gin.H{
		"message":           "Manual sync completed",
		"successCount":      successCount,
		"hasSyncedCount":    hasSyncedCount,
		"deleteCount":       deleteCount,
		"failedNotices":     failedNotices,
		"totalProcessed":    len(oldNotices),
		"has_sync_notices":  hasSyncNotices,
		"need_sync_notices": needSyncNotices,
		"change_notices":    changeNotices,
	}, nil
}

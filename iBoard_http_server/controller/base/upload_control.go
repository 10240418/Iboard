package http_base_controller

import (
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	base_services "github.com/The-Healthist/iboard_http_service/services/base"
	"github.com/The-Healthist/iboard_http_service/services/container"
	"github.com/The-Healthist/iboard_http_service/utils/field"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type UploadController struct {
	Ctx       *gin.Context
	Container *container.ServiceContainer
}

func NewUploadController(ctx *gin.Context, container *container.ServiceContainer) *UploadController {
	return &UploadController{
		Ctx:       ctx,
		Container: container,
	}
}

// HandleFuncUpload returns a gin.HandlerFunc for the specified method
func HandleFuncUpload(container *container.ServiceContainer, method string) gin.HandlerFunc {
	switch method {
	case "getUploadParams":
		return func(ctx *gin.Context) {
			controller := NewUploadController(ctx, container)
			controller.GetUploadParams()
		}
	case "uploadCallback":
		return func(ctx *gin.Context) {
			controller := NewUploadController(ctx, container)
			controller.UploadCallback()
		}
	case "uploadCallbackSync":
		return func(ctx *gin.Context) {
			controller := NewUploadController(ctx, container)
			controller.UploadCallbackSync()
		}
	case "getUploadParamsSync":
		return func(ctx *gin.Context) {
			controller := NewUploadController(ctx, container)
			controller.GetUploadParamsSync()
		}
	default:
		return func(ctx *gin.Context) {
			ctx.JSON(400, gin.H{"error": "invalid method"})
		}
	}
}

func (c *UploadController) GetUploadParams() {
	log.Println("Processing upload params request")

	// Get JWT claims
	claims, exists := c.Ctx.Get("claims")
	if !exists {
		log.Println("No JWT claims found in context")
		c.Ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	mapClaims, ok := claims.(jwt.MapClaims)
	if !ok {
		log.Println("Invalid JWT claims format")
		c.Ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid token claims format",
		})
		return
	}

	// Parse and validate claims
	var uploaderID uint
	var uploaderType string
	var uploaderEmail string

	isAdmin, _ := mapClaims["isAdmin"].(bool)
	isBuildingAdmin, _ := mapClaims["isBuildingAdmin"].(bool)

	if isAdmin {
		// Handle super admin case
		if id, ok := mapClaims["id"].(float64); ok {
			uploaderID = uint(id)
			uploaderType = "superAdmin"
			if email, ok := mapClaims["email"].(string); ok {
				uploaderEmail = email
			}
			log.Printf("Identified as SuperAdmin with ID: %d, Email: %s\n", uploaderID, uploaderEmail)
		} else {
			log.Println("Invalid admin ID in token")
			c.Ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid admin ID",
			})
			return
		}
	} else if isBuildingAdmin {
		// Handle building admin case
		if id, ok := mapClaims["id"].(float64); ok {
			uploaderID = uint(id)
			uploaderType = "buildingAdmin"
			if email, ok := mapClaims["email"].(string); ok {
				uploaderEmail = email
			}
			log.Printf("Identified as BuildingAdmin with ID: %d, Email: %s\n", uploaderID, uploaderEmail)
		} else {
			log.Println("Invalid building admin ID in token")
			c.Ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid building admin ID",
			})
			return
		}
	} else {
		log.Println("Invalid uploader type")
		c.Ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid uploader type",
		})
		return
	}

	// Save uploader info to cache
	if err := c.Container.GetService("upload").(base_services.IUploadService).SaveUploaderInfo(uploaderID, uploaderType, uploaderEmail); err != nil {
		log.Printf("Failed to save uploader info: %v\n", err)
		c.Ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process uploader info",
		})
		return
	}

	// Process file upload parameters
	var req struct {
		FileName string `json:"fileName" binding:"required"`
	}

	if err := c.Ctx.ShouldBindJSON(&req); err != nil {
		log.Printf("Invalid request parameters: %v\n", err)
		c.Ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing required parameters",
		})
		return
	}

	// Generate directory and filename
	currentTime := time.Now()
	dir := currentTime.Format("2006-01-02") + "/"
	ext := path.Ext(req.FileName)
	newFileName := uuid.New().String() + ext
	fullPath := dir + newFileName

	// Get upload policy
	policy, err := c.Container.GetService("upload").(base_services.IUploadService).GetUploadParams(fullPath)
	if err != nil {
		log.Printf("Failed to get upload params: %v\n", err)
		c.Ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Save filename mapping
	if err := c.Container.GetService("upload").(base_services.IUploadService).SaveFileNameMapping(newFileName, fullPath); err != nil {
		log.Printf("Failed to save filename mapping: %v\n", err)
		c.Ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save filename mapping",
		})
		return
	}

	log.Printf("Successfully generated upload params for file: %s\n", fullPath)
	c.Ctx.JSON(http.StatusOK, gin.H{
		"data": policy,
	})
}

func (c *UploadController) UploadCallback() {
	log.Println("Received upload callback request")

	// Get content-md5 from header
	contentMD5 := c.Ctx.GetHeader("content-md5")
	log.Printf("Content-MD5 from header: %s\n", contentMD5)

	// Read and log raw request body
	body, err := io.ReadAll(c.Ctx.Request.Body)
	if err != nil {
		log.Printf("Failed to read request body: %v\n", err)
		c.Ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read request body",
		})
		return
	}
	log.Printf("Raw request body: %s\n", string(body))

	// Restore request body for form parsing
	c.Ctx.Request.Body = io.NopCloser(strings.NewReader(string(body)))

	// Parse form data
	if err := c.Ctx.Request.ParseForm(); err != nil {
		log.Printf("Failed to parse form: %v\n", err)
		c.Ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to parse form data",
		})
		return
	}

	log.Printf("Form values: %+v\n", c.Ctx.Request.Form)

	// Get callback data
	objectPath := c.Ctx.Request.Form.Get("object")
	sizeStr := c.Ctx.Request.Form.Get("size")
	mimeType := c.Ctx.Request.Form.Get("mimeType")
	heightStr := c.Ctx.Request.Form.Get("height")
	widthStr := c.Ctx.Request.Form.Get("width")

	size, _ := strconv.ParseInt(sizeStr, 10, 64)
	height, _ := strconv.Atoi(heightStr)
	width, _ := strconv.Atoi(widthStr)

	log.Printf("Parsed callback data: object=%s, size=%d, mimeType=%s, height=%d, width=%d\n",
		objectPath, size, mimeType, height, width)

	// Get stored dirPath from Redis
	dirPath, err := c.Container.GetService("upload").(base_services.IUploadService).GetLatestDirPath()
	if err != nil {
		log.Printf("Failed to get stored dirPath from Redis: %v\n", err)
		c.Ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get stored dirPath",
		})
		return
	}
	log.Printf("Retrieved stored dirPath from Redis: %s\n", dirPath)

	// Compare paths
	if objectPath != dirPath {
		log.Printf("Path mismatch: callback=%s, stored=%s\n", objectPath, dirPath)
		c.Ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Path mismatch",
		})
		return
	}
	log.Println("Path validation successful")

	// Get stored uploader info from cache
	uploaderID, uploaderType, uploaderEmail, err := c.Container.GetService("upload").(base_services.IUploadService).GetLatestUploaderInfo()
	if err != nil {
		log.Printf("Failed to get uploader info from cache: %v\n", err)
		c.Ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get uploader info",
		})
		return
	}
	log.Printf("Retrieved uploader info: ID=%d, Type=%s, Email=%s\n", uploaderID, uploaderType, uploaderEmail)

	// Create file record
	var fileUploaderType field.FileUploaderType
	if uploaderType == "superAdmin" {
		fileUploaderType = field.UploaderTypeSuperAdmin
	} else {
		fileUploaderType = field.UploaderTypeBuildingAdmin
	}

	// Construct full path
	fullPath := os.Getenv("HOST") + "/" + objectPath
	log.Printf("Constructed full file path: %s\n", fullPath)

	file := &base_models.File{
		Path:         fullPath,
		Size:         size,
		MimeType:     mimeType,
		Md5:          contentMD5,
		Oss:          "aws",
		Uploader:     uploaderEmail,
		UploaderType: fileUploaderType,
		UploaderID:   uploaderID,
	}

	log.Printf("Creating file record with details: %+v\n", file)

	if err := c.Container.GetService("file").(base_services.InterfaceFileService).Create(file); err != nil {
		log.Printf("Failed to create file record: %v\n", err)
		c.Ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create file record",
		})
		return
	}
	log.Println("File record created successfully")

	// Return success response
	c.Ctx.JSON(http.StatusOK, gin.H{
		"Status": "OK",
	})
}
func (c *UploadController) UploadCallbackSync() {
	c.Ctx.JSON(http.StatusOK, gin.H{
		"Status": "OK",
	})
}
func (c *UploadController) GetUploadParamsSync() {
	log.Println("Processing upload params request")

	// Get JWT claims
	claims, exists := c.Ctx.Get("claims")
	if !exists {
		log.Println("No JWT claims found in context")
		c.Ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	mapClaims, ok := claims.(jwt.MapClaims)
	if !ok {
		log.Println("Invalid JWT claims format")
		c.Ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid token claims format",
		})
		return
	}

	// Parse and validate claims
	var uploaderID uint
	var uploaderType string
	var uploaderEmail string

	isAdmin, _ := mapClaims["isAdmin"].(bool)
	isBuildingAdmin, _ := mapClaims["isBuildingAdmin"].(bool)

	if isAdmin {
		// Handle super admin case
		if id, ok := mapClaims["id"].(float64); ok {
			uploaderID = uint(id)
			uploaderType = "superAdmin"
			if email, ok := mapClaims["email"].(string); ok {
				uploaderEmail = email
			}
			log.Printf("Identified as SuperAdmin with ID: %d, Email: %s\n", uploaderID, uploaderEmail)
		} else {
			log.Println("Invalid admin ID in token")
			c.Ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid admin ID",
			})
			return
		}
	} else if isBuildingAdmin {
		// Handle building admin case
		if id, ok := mapClaims["id"].(float64); ok {
			uploaderID = uint(id)
			uploaderType = "buildingAdmin"
			if email, ok := mapClaims["email"].(string); ok {
				uploaderEmail = email
			}
			log.Printf("Identified as BuildingAdmin with ID: %d, Email: %s\n", uploaderID, uploaderEmail)
		} else {
			log.Println("Invalid building admin ID in token")
			c.Ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid building admin ID",
			})
			return
		}
	} else {
		log.Println("Invalid uploader type")
		c.Ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid uploader type",
		})
		return
	}

	// Save uploader info to cache
	if err := c.Container.GetService("upload").(base_services.IUploadService).SaveUploaderInfo(uploaderID, uploaderType, uploaderEmail); err != nil {
		log.Printf("Failed to save uploader info: %v\n", err)
		c.Ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process uploader info",
		})
		return
	}

	// Process file upload parameters
	var req struct {
		FileName string `json:"fileName" binding:"required"`
	}

	if err := c.Ctx.ShouldBindJSON(&req); err != nil {
		log.Printf("Invalid request parameters: %v\n", err)
		c.Ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing required parameters",
		})
		return
	}

	// Generate directory and filename
	currentTime := time.Now()
	dir := currentTime.Format("2006-01-02") + "/"
	ext := path.Ext(req.FileName)
	newFileName := uuid.New().String() + ext
	fullPath := dir + newFileName

	// Get upload policy
	policy, err := c.Container.GetService("upload").(base_services.IUploadService).GetUploadParamsSync(fullPath)
	if err != nil {
		log.Printf("Failed to get upload params: %v\n", err)
		c.Ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Save filename mapping
	if err := c.Container.GetService("upload").(base_services.IUploadService).SaveFileNameMapping(newFileName, fullPath); err != nil {
		log.Printf("Failed to save filename mapping: %v\n", err)
		c.Ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save filename mapping",
		})
		return
	}

	log.Printf("Successfully generated upload params for file: %s\n", fullPath)
	c.Ctx.JSON(http.StatusOK, gin.H{
		"data": policy,
	})
}

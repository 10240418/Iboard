package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	databases "github.com/The-Healthist/iboard_http_service/database"
	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	"github.com/The-Healthist/iboard_http_service/router"
	base_services "github.com/The-Healthist/iboard_http_service/services/base"
	"github.com/The-Healthist/iboard_http_service/services/container"
	"github.com/The-Healthist/iboard_http_service/utils"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	maxLogFileSize = 1 * 1024 * 1024 // 1MB per log file
	maxLogFiles    = 365             // Maximum number of log files to keep
	logTimeFormat  = "2006-01-02"    // Log file date format
)

// setupLogging configures the logging for both file and console output
func setupLogging() (*os.File, error) {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %v", err)
	}

	// Clean up old log files if we exceed the maximum
	if err := cleanupOldLogs(); err != nil {
		return nil, fmt.Errorf("failed to cleanup old logs: %v", err)
	}

	// Create log file with timestamp
	currentTime := time.Now().Format(logTimeFormat)
	logFileName := fmt.Sprintf("logs/app_%s.log", currentTime)

	// Check if current log file exists and its size
	if info, err := os.Stat(logFileName); err == nil {
		if info.Size() >= maxLogFileSize {
			// If file exists and exceeds size limit, create a new file with timestamp
			logFileName = fmt.Sprintf("logs/app_%s_%d.log", currentTime, time.Now().Unix())
		}
	}

	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	// Set up multi-writer for both console and file output
	gin.DefaultWriter = io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(gin.DefaultWriter)

	return logFile, nil
}

func cleanupOldLogs() error {
	files, err := os.ReadDir("logs")
	if err != nil {
		return fmt.Errorf("failed to read logs directory: %v", err)
	}

	var logFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), "app_") {
			logFiles = append(logFiles, filepath.Join("logs", file.Name()))
		}
	}

	// If we have more files than the maximum allowed
	if len(logFiles) >= maxLogFiles {
		// Sort files by modification time (oldest first)
		sort.Slice(logFiles, func(i, j int) bool {
			iInfo, _ := os.Stat(logFiles[i])
			jInfo, _ := os.Stat(logFiles[j])
			return iInfo.ModTime().Before(jInfo.ModTime())
		})

		// Remove oldest files until we're under the limit
		for i := 0; i < len(logFiles)-maxLogFiles+1; i++ {
			if err := os.Remove(logFiles[i]); err != nil {
				return fmt.Errorf("failed to remove old log file %s: %v", logFiles[i], err)
			}
		}
	}

	return nil
}

func initSuperAdmin(db *gorm.DB) error {
	var admin base_models.SuperAdmin

	log.Println("Checking for existing super admin...")
	result := db.Where("email = ?", "admin@example.com").First(&admin)
	if result.Error == nil {
		log.Println("Super admin already exists")
		return nil
	}

	log.Println("Creating default super admin...")
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte("admin123"),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	newAdmin := base_models.SuperAdmin{
		Email:    "admin@example.com",
		Password: string(hashedPassword),
	}

	if err := db.Create(&newAdmin).Error; err != nil {
		return fmt.Errorf("failed to create admin: %v", err)
	}

	log.Println("Default super admin created successfully")
	return nil
}

func main() {
	// Set up logging
	logFile, err := setupLogging()
	if err != nil {
		fmt.Printf("Failed to set up logging: %v\n", err)
		return
	}
	defer logFile.Close()

	log.Println("Starting iBoard HTTP Server...")

	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	log.Println("Environment variables loaded successfully")

	// Init Email
	log.Println("Initializing email service...")
	emailPort, err := strconv.ParseInt(os.Getenv("SMTP_PORT"), 10, 32)
	if err != nil {
		log.Fatal("SMTP port configuration error:", err)
	}
	utils.InitEmail(os.Getenv("SMTP_ADDR"), int(emailPort), os.Getenv("SMTP_USER"), os.Getenv("SMTP_PASS"))
	log.Println("Email service initialized successfully")

	// Init DB
	log.Println("Initializing database connection...")
	db := databases.InitDB(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	if db == nil {
		log.Fatal("Failed to initialize database")
	}
	log.Println("Database connection established successfully")

	// Initialize super admin
	if err := initSuperAdmin(db); err != nil {
		log.Printf("Failed to initialize super admin: %v\n", err)
	}

	// Initialize Redis connection
	log.Println("Initializing Redis connection...")
	maxRedisRetries := 5
	var redisErr error
	for i := 0; i < maxRedisRetries; i++ {
		redisErr = databases.InitRedis()
		if redisErr == nil {
			log.Println("Redis connection established successfully")
			break
		}
		log.Printf("Failed to initialize Redis (attempt %d/%d): %v", i+1, maxRedisRetries, redisErr)
		if i < maxRedisRetries-1 {
			time.Sleep(time.Second * 3)
		}
	}
	if redisErr != nil {
		log.Fatal("Failed to initialize Redis after multiple attempts:", redisErr)
	}

	// Ensure Redis connection is closed when the program exits
	defer func() {
		if err := databases.CloseRedis(); err != nil {
			log.Printf("Error closing Redis connection: %v", err)
		}
	}()

	// Initialize service container
	log.Println("Initializing service container...")
	serviceContainer := container.NewServiceContainer(db)
	log.Println("Service container initialized successfully")

	// Configure Gin
	log.Println("Configuring Gin framework...")
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// Add custom logging middleware
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Custom log format
		return fmt.Sprintf("[GIN] %s | %s | %d | %s | %s | %s | %s\n",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			param.ClientIP,
			param.StatusCode,
			param.Latency,
			param.Method,
			param.Path,
			param.ErrorMessage,
		)
	}))
	r.Use(gin.Recovery())

	log.Println("Registering routes...")
	router.RegisterRoute(r)
	log.Println("Routes registered successfully")

	// Start notice sync scheduler
	log.Println("Starting notice sync scheduler...")
	ctx := context.Background()
	noticeSyncService := serviceContainer.GetService("noticeSync").(base_services.InterfaceNoticeSyncService)
	noticeSyncService.StartSyncScheduler(ctx)
	log.Println("Notice sync scheduler started successfully")

	// Start server
	serverAddr := "0.0.0.0:10031"
	log.Printf("Starting HTTP server on %s...\n", serverAddr)
	if err := r.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

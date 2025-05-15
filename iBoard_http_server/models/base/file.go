package base_models

import "github.com/The-Healthist/iboard_http_service/utils/field"

// File 文件模型
type File struct {
	ModelFields
	Size         int64                  `json:"size"`
	Md5          string                 `json:"md5"          gorm:"size:255;"`
	Path         string                 `json:"path"         gorm:"size:255;uniqueIndex"` // 唯一索引
	MimeType     string                 `json:"mimeType"     gorm:"size:100"`
	Oss          string                 `json:"oss"          gorm:"size:50"`
	Uploader     string                 `json:"uploader"     gorm:"size:255"`
	UploaderID   uint                   `json:"uploaderId"   gorm:"index"`
	UploaderType field.FileUploaderType `json:"uploaderType" gorm:"size:50"`
}

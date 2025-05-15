package base_services

import (
	"errors"
	"fmt"

	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	"gorm.io/gorm"
)

type InterfaceFileService interface {
	Create(file *base_models.File) error
	Get(query map[string]interface{}, paginate map[string]interface{}) ([]base_models.File, base_models.PaginationResult, error)
	Update(id uint, updates map[string]interface{}) error
	Delete(ids []uint) error
	GetByID(id uint) (*base_models.File, error)
}

type FileService struct {
	db *gorm.DB
}

func NewFileService(db *gorm.DB) InterfaceFileService {
	return &FileService{db: db}
}

func (s *FileService) Create(file *base_models.File) error {
	// Check if file with same path exists
	var count int64
	if err := s.db.Model(&base_models.File{}).Where("path = ?", file.Path).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("file with this path already exists")
	}

	return s.db.Create(file).Error
}

func (s *FileService) Get(query map[string]interface{}, paginate map[string]interface{}) ([]base_models.File, base_models.PaginationResult, error) {
	var files []base_models.File
	var total int64
	db := s.db.Model(&base_models.File{})

	if mimeType, ok := query["mimeType"].(string); ok && mimeType != "" {
		db = db.Where("mime_type = ?", mimeType)
	}

	if oss, ok := query["oss"].(string); ok && oss != "" {
		db = db.Where("oss = ?", oss)
	}

	if uploaderType, ok := query["uploaderType"].(string); ok && uploaderType != "" {
		db = db.Where("uploader_type = ?", uploaderType)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, base_models.PaginationResult{}, err
	}

	pageSize := paginate["pageSize"].(int)
	pageNum := paginate["pageNum"].(int)
	offset := (pageNum - 1) * pageSize

	if desc, ok := paginate["desc"].(bool); ok && desc {
		db = db.Order("created_at DESC")
	} else {
		db = db.Order("created_at ASC")
	}

	if err := db.Limit(pageSize).Offset(offset).Find(&files).Error; err != nil {
		return nil, base_models.PaginationResult{}, err
	}

	return files, base_models.PaginationResult{
		Total:    int(total),
		PageSize: pageSize,
		PageNum:  pageNum,
	}, nil
}

func (s *FileService) Update(id uint, updates map[string]interface{}) error {
	if path, ok := updates["path"].(string); ok {
		// Check if new path conflicts with existing files
		var count int64
		if err := s.db.Model(&base_models.File{}).Where("path = ? AND id != ?", path, id).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("file with this path already exists")
		}
	}

	result := s.db.Model(&base_models.File{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("file not found")
	}
	return nil
}

func (s *FileService) Delete(ids []uint) error {
	result := s.db.Delete(&base_models.File{}, ids)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no records found to delete")
	}
	return nil
}

func (s *FileService) GetByID(id uint) (*base_models.File, error) {
	var file base_models.File
	if err := s.db.First(&file, id).Error; err != nil {
		return nil, fmt.Errorf("file not found: %v", err)
	}
	return &file, nil
}

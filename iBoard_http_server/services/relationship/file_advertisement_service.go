package http_relationship_service

import (
	"errors"

	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	"gorm.io/gorm"
)

type InterfaceFileAdvertisementService interface {
	BindFile(advertisementID uint, fileID uint) error
	UnbindFile(advertisementID uint) error
	GetAdvertisementByFileID(fileID uint) (*base_models.Advertisement, error)
	GetFileByAdvertisementID(advertisementID uint) (*base_models.File, error)
	AdvertisementExists(advertisementID uint) (bool, error)
	FileExists(fileID uint) (bool, error)
}

type FileAdvertisementService struct {
	db *gorm.DB
}

func NewFileAdvertisementService(db *gorm.DB) InterfaceFileAdvertisementService {
	return &FileAdvertisementService{db: db}
}

func (s *FileAdvertisementService) BindFile(advertisementID uint, fileID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&base_models.Advertisement{}).
			Where("id = ?", advertisementID).
			Update("file_id", fileID).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *FileAdvertisementService) UnbindFile(advertisementID uint) error {
	return s.db.Model(&base_models.Advertisement{}).
		Where("id = ?", advertisementID).
		Update("file_id", nil).Error
}

func (s *FileAdvertisementService) GetAdvertisementByFileID(fileID uint) (*base_models.Advertisement, error) {
	var advertisement base_models.Advertisement
	if err := s.db.Where("file_id = ?", fileID).First(&advertisement).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &advertisement, nil
}

func (s *FileAdvertisementService) GetFileByAdvertisementID(advertisementID uint) (*base_models.File, error) {
	var advertisement base_models.Advertisement
	if err := s.db.Preload("File").First(&advertisement, advertisementID).Error; err != nil {
		return nil, err
	}
	return advertisement.File, nil
}

func (s *FileAdvertisementService) AdvertisementExists(advertisementID uint) (bool, error) {
	var advertisement base_models.Advertisement
	err := s.db.First(&advertisement, advertisementID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *FileAdvertisementService) FileExists(fileID uint) (bool, error) {
	var file base_models.File
	err := s.db.First(&file, fileID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

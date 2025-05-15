package base_services

import (
	"errors"
	"time"

	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	"github.com/The-Healthist/iboard_http_service/utils/field"
	"gorm.io/gorm"
)

type InterfaceAdvertisementService interface {
	Create(advertisement *base_models.Advertisement) error
	Get(query map[string]interface{}, paginate map[string]interface{}) ([]base_models.Advertisement, base_models.PaginationResult, error)
	Update(id uint, updates map[string]interface{}) (*base_models.Advertisement, error)
	Delete(ids []uint) error
	GetByID(id uint) (*base_models.Advertisement, error)
}

type AdvertisementService struct {
	db *gorm.DB
}

func NewAdvertisementService(db *gorm.DB) InterfaceAdvertisementService {
	if db == nil {
		panic("database connection is nil")
	}
	return &AdvertisementService{db: db}
}

func (s *AdvertisementService) Create(advertisement *base_models.Advertisement) error {
	if s.db == nil {
		return errors.New("database connection is nil")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(advertisement).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *AdvertisementService) Get(query map[string]interface{}, paginate map[string]interface{}) ([]base_models.Advertisement, base_models.PaginationResult, error) {
	if s.db == nil {
		return nil, base_models.PaginationResult{}, errors.New("database connection is nil")
	}

	var advertisements []base_models.Advertisement
	var total int64
	db := s.db.Model(&base_models.Advertisement{})

	if search, ok := query["search"].(string); ok && search != "" {
		db = db.Where("title LIKE ? OR description LIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}

	if advertisementType, ok := query["type"].(string); ok && advertisementType != "" {
		db = db.Where("type = ?", advertisementType)
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

	if err := db.Preload("File").
		Limit(pageSize).Offset(offset).
		Find(&advertisements).Error; err != nil {
		return nil, base_models.PaginationResult{}, err
	}

	now := time.Now()
	for i := range advertisements {
		if advertisements[i].StartTime.IsZero() {
			advertisements[i].StartTime = time.Date(2024, 12, 23, 16, 30, 34, 156000000, time.FixedZone("CST", 8*3600))
		}
		if advertisements[i].EndTime.IsZero() {
			advertisements[i].EndTime = time.Date(2025, 12, 23, 16, 30, 34, 156000000, time.FixedZone("CST", 8*3600))
		}
		if advertisements[i].Status == "" {
			advertisements[i].Status = field.Status("active")
		}

		// Check if advertisement has expired
		if advertisements[i].EndTime.Before(now) && advertisements[i].Status == field.Status("active") {
			// Update status in database
			if err := s.db.Model(&base_models.Advertisement{}).Where("id = ?", advertisements[i].ID).Update("status", field.Status("inactive")).Error; err != nil {
				return nil, base_models.PaginationResult{}, err
			}
			// Update status in memory
			advertisements[i].Status = field.Status("inactive")
		}
	}

	return advertisements, base_models.PaginationResult{
		Total:    int(total),
		PageSize: pageSize,
		PageNum:  pageNum,
	}, nil
}

func (s *AdvertisementService) Update(id uint, updates map[string]interface{}) (*base_models.Advertisement, error) {
	if s.db == nil {
		return nil, errors.New("database connection is nil")
	}

	var advertisement base_models.Advertisement

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&base_models.Advertisement{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return err
		}

		if err := tx.Preload("File").Preload("Buildings").First(&advertisement, id).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &advertisement, nil
}

func (s *AdvertisementService) Delete(ids []uint) error {
	if s.db == nil {
		return errors.New("database connection is nil")
	}

	result := s.db.Delete(&base_models.Advertisement{}, ids)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no records found to delete")
	}
	return nil
}

func (s *AdvertisementService) GetByID(id uint) (*base_models.Advertisement, error) {
	if s.db == nil {
		return nil, errors.New("database connection is nil")
	}

	var advertisement base_models.Advertisement
	if err := s.db.Preload("File").First(&advertisement, id).Error; err != nil {
		return nil, err
	}

	if advertisement.StartTime.IsZero() {
		advertisement.StartTime = time.Date(2024, 12, 23, 16, 30, 34, 156000000, time.FixedZone("CST", 8*3600))
	}
	if advertisement.EndTime.IsZero() {
		advertisement.EndTime = time.Date(2025, 12, 23, 16, 30, 34, 156000000, time.FixedZone("CST", 8*3600))
	}
	if advertisement.Status == "" {
		advertisement.Status = field.Status("active")
	}
	if advertisement.File != nil && advertisement.File.ID == 0 {
		advertisement.File = nil
	}

	return &advertisement, nil
}

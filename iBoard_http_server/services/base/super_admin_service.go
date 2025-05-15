package base_services

import (
	"errors"
	"fmt"

	models "github.com/The-Healthist/iboard_http_service/models/base"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type InterfaceSuperAdminService interface {
	CheckPassword(email string, password string) error
	CreateSuperAdmin(admin *models.SuperAdmin) error
	GetSuperAdminById(id uint) (*models.SuperAdmin, error)
	GetSuperAdminByEmail(email string) (*models.SuperAdmin, error)
	GetSuperAdmins(query map[string]interface{}, paginate map[string]interface{}) ([]models.SuperAdmin, models.PaginationResult, error)
	UpdateSuperAdmin(adminObj *models.SuperAdmin, admin map[string]interface{}) error
	DeleteSuperAdmin(id uint) error
	DeleteSuperAdmins(ids []uint) error
}

type SuperAdminService struct {
	db *gorm.DB
}

func NewSuperAdminService(db *gorm.DB) InterfaceSuperAdminService {
	return &SuperAdminService{db: db}
}

func (s *SuperAdminService) CheckPassword(email, password string) error {
	var admin models.SuperAdmin

	if err := s.db.Where("email = ?", email).First(&admin).Error; err != nil {
		return errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(admin.Password),
		[]byte(password),
	); err != nil {
		return errors.New("invalid credentials")
	}

	return nil
}

func (s *SuperAdminService) GetSuperAdmins(query map[string]interface{}, paginate map[string]interface{}) ([]models.SuperAdmin, models.PaginationResult, error) {
	var admins []models.SuperAdmin
	var total int64
	db := s.db.Model(&models.SuperAdmin{})

	if search, ok := query["search"].(string); ok && search != "" {
		db = db.Where("email LIKE ? OR name LIKE ?",
			"%"+search+"%",
			"%"+search+"%",
		)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, models.PaginationResult{}, err
	}

	pageSize := paginate["pageSize"].(int)
	pageNum := paginate["pageNum"].(int)
	offset := (pageNum - 1) * pageSize

	if desc, ok := paginate["desc"].(bool); ok && desc {
		db = db.Order("created_at DESC")
	} else {
		db = db.Order("created_at ASC")
	}

	if err := db.Limit(pageSize).Offset(offset).Find(&admins).Error; err != nil {
		return nil, models.PaginationResult{}, err
	}

	return admins, models.PaginationResult{
		Total:    int(total),
		PageSize: pageSize,
		PageNum:  pageNum,
	}, nil
}

func (s *SuperAdminService) CreateSuperAdmin(admin *models.SuperAdmin) error {
	var count int64
	if err := s.db.Model(&models.SuperAdmin{}).
		Where("email = ?", admin.Email).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(admin.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return fmt.Errorf("password encryption failed: %v", err)
	}
	admin.Password = string(hashedPassword)

	return s.db.Create(admin).Error
}

func (s *SuperAdminService) UpdateSuperAdmin(adminObj *models.SuperAdmin, admin map[string]interface{}) error {
	return s.db.Model(adminObj).Updates(admin).Error
}

func (s *SuperAdminService) DeleteSuperAdmin(id uint) error {
	result := s.db.Delete(&models.SuperAdmin{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no records found to delete")
	}
	return nil
}

func (s *SuperAdminService) DeleteSuperAdmins(ids []uint) error {
	result := s.db.Delete(&models.SuperAdmin{}, ids)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no records found to delete")
	}
	return nil
}

func (s *SuperAdminService) GetSuperAdminById(id uint) (*models.SuperAdmin, error) {
	var admin models.SuperAdmin
	if err := s.db.First(&admin, id).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (s *SuperAdminService) GetSuperAdminByEmail(email string) (*models.SuperAdmin, error) {
	var admin models.SuperAdmin
	if err := s.db.Where("email = ?", email).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

package base_models

import "github.com/The-Healthist/iboard_http_service/utils/field"

// Admin 管理员模型
type BuildingAdmin struct {
	ModelFields
	Email     string       `json:"email"       gorm:"size:255;not null;unique"`
	Password  string       `json:"-"           gorm:"size:255;not null"`
	Status    field.Status `json:"status"      gorm:"default:active"`
	Buildings []Building   `json:"buildings"   gorm:"many2many:building_admins_buildings;"`
}

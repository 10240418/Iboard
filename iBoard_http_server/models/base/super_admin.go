package base_models

// SuperAdmin 超级管理员模型
type SuperAdmin struct {
	ModelFields
	Email    string `json:"email"    gorm:"size:255;not null;uniqueIndex"`
	Password string `json:"-"           gorm:"size:255;not null"`
}

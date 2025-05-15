package http_relationship_service

import (
	"errors"

	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	"gorm.io/gorm"
)

type InterfaceFileNoticeService interface {
	BindFile(noticeID uint, fileID uint) error
	UnbindFile(noticeID uint) error
	GetNoticeByFileID(fileID uint) (*base_models.Notice, error)
	GetFileByNoticeID(noticeID uint) (*base_models.File, error)
	NoticeExists(noticeID uint) (bool, error)
	FileExists(fileID uint) (bool, error)
}

type FileNoticeService struct {
	db *gorm.DB
}

func NewFileNoticeService(db *gorm.DB) InterfaceFileNoticeService {
	return &FileNoticeService{db: db}
}

func (s *FileNoticeService) BindFile(noticeID uint, fileID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 更新通知的 FileID
		if err := tx.Model(&base_models.Notice{}).
			Where("id = ?", noticeID).
			Update("file_id", fileID).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *FileNoticeService) UnbindFile(noticeID uint) error {
	return s.db.Model(&base_models.Notice{}).
		Where("id = ?", noticeID).
		Update("file_id", nil).Error
}

func (s *FileNoticeService) GetNoticeByFileID(fileID uint) (*base_models.Notice, error) {
	var notice base_models.Notice
	if err := s.db.Where("file_id = ?", fileID).First(&notice).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &notice, nil
}

func (s *FileNoticeService) GetFileByNoticeID(noticeID uint) (*base_models.File, error) {
	var notice base_models.Notice
	if err := s.db.Preload("File").First(&notice, noticeID).Error; err != nil {
		return nil, err
	}
	return notice.File, nil
}

func (s *FileNoticeService) NoticeExists(noticeID uint) (bool, error) {
	var notice base_models.Notice
	err := s.db.First(&notice, noticeID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *FileNoticeService) FileExists(fileID uint) (bool, error) {
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

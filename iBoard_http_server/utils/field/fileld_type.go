package field

// advertisement type
type AdvertisementType string

const (
	AdTypeVideo AdvertisementType = "video"
	AdTypeImage AdvertisementType = "image"
)

// advertisement display type
type AdvertisementDisplay string

const (
	AdDisplayFull    AdvertisementDisplay = "full"
	AdDisplayTop     AdvertisementDisplay = "top"
	AdDisplayTopFull AdvertisementDisplay = "topfull"
)

// notice type
type NoticeType string

const (
	NoticeTypeUrgent     NoticeType = "urgent"
	NoticeTypeNormal     NoticeType = "normal"
	NoticeTypeBuilding   NoticeType = "building"
	NoticeTypeGovernment NoticeType = "government"
)

type NoticeOldType string

const (
	NoticeOldTypeUrgent     NoticeOldType = "urgent"
	NoticeOldTypeCommon     NoticeOldType = "common"
	NoticeOldTypeIo         NoticeOldType = "io"
	NoticeOldTypeGovernment NoticeOldType = "gov"
)

// notice status
type Status string

const (
	StatusPending  Status = "pending"
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
)

// file type
type FileType string

const (
	FileTypePdf FileType = "pdf"
)

// upload
type FileUploaderType string

const (
	UploaderTypeBuildingAdmin FileUploaderType = "buildingAdmin"
	UploaderTypeSuperAdmin    FileUploaderType = "superAdmin"
)

// validate method
func IsValidFileUploaderType(t string) bool {
	switch FileUploaderType(t) {
	case UploaderTypeBuildingAdmin, UploaderTypeSuperAdmin:
		return true
	}
	return false
}

func IsValidAdvertisementType(t string) bool {
	switch AdvertisementType(t) {
	case AdTypeVideo, AdTypeImage:
		return true
	}
	return false
}

func IsValidAdvertisementDisplay(d string) bool {
	switch AdvertisementDisplay(d) {
	case AdDisplayFull, AdDisplayTop, AdDisplayTopFull:
		return true
	}
	return false
}

func IsValidNoticeType(t string) bool {
	switch NoticeType(t) {
	case NoticeTypeUrgent, NoticeTypeNormal, NoticeTypeBuilding, NoticeTypeGovernment:
		return true
	}
	return false
}

// validate status
func IsValidStatus(s string) bool {
	switch Status(s) {
	case StatusPending, StatusActive, StatusInactive:
		return true
	}
	return false
}

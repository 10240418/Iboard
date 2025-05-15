package base_services

import (
	"fmt"
	"os"
	"time"

	base_models "github.com/The-Healthist/iboard_http_service/models/base"
	"github.com/golang-jwt/jwt/v4"
)

type AuthClaims struct {
	Email   string `json:"email"`
	IsAdmin bool   `json:"isAdmin"`
	jwt.StandardClaims
}

type IJWTService interface {
	GenerateToken(claims jwt.MapClaims) (string, error)
	ValidateToken(token string) (*jwt.Token, error)
	GenerateBuildingAdminToken(admin *base_models.BuildingAdmin) (string, error)
	GenerateDeviceToken(device *base_models.Device) (string, error)
	GenerateSuperAdminToken(admin *base_models.SuperAdmin) (string, error)
}

type JWTService struct {
	secretKey string
}

func NewJWTService() IJWTService {
	return &JWTService{
		secretKey: GetSecretKey(),
	}
}

func GetSecretKey() string {
	secret := os.Getenv("SECRET")
	if secret == "" {
		secret = "secret"
	}
	return secret
}

func (service *JWTService) GenerateToken(claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString([]byte(service.secretKey))
	if err != nil {
		return "", err
	}
	return t, nil
}

func (service *JWTService) ValidateToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, isValid := token.Method.(*jwt.SigningMethodHMAC); !isValid {
			return nil, fmt.Errorf("invalid token %s", token.Header["alg"])
		}
		return []byte(service.secretKey), nil
	})
}

func (s *JWTService) GenerateBuildingAdminToken(admin *base_models.BuildingAdmin) (string, error) {
	claims := jwt.MapClaims{
		"id":              admin.ID,
		"email":           admin.Email,
		"isBuildingAdmin": true,
		"exp":             time.Now().Add(time.Hour * 24).Unix(),
	}
	return s.GenerateToken(claims)
}

func (s *JWTService) GenerateDeviceToken(device *base_models.Device) (string, error) {
	claims := jwt.MapClaims{
		"deviceId":   device.DeviceID,
		"buildingId": device.BuildingID,
		"isDevice":   true,
		"exp":        time.Now().Add(time.Hour * 24).Unix(),
	}
	return s.GenerateToken(claims)
}

func (s *JWTService) GenerateSuperAdminToken(admin *base_models.SuperAdmin) (string, error) {
	claims := jwt.MapClaims{
		"id":      admin.ID,
		"email":   admin.Email,
		"isAdmin": true,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}
	return s.GenerateToken(claims)
}

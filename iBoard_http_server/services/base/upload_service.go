package base_services

import (
	"crypto"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type IUploadService interface {
	GetUploadParams(uploadDir string) (map[string]interface{}, error)
	GetUploadParamsSync(uploadDir string) (map[string]interface{}, error)
	SaveCallbackData(data *CallbackData) error
	SaveFileNameMapping(newFileName string, dirPath string) error
	GetLatestFileName() (string, error)
	GetLatestDirPath() (string, error)
	SaveUploaderInfo(id uint, uploaderType string, email string) error
	GetLatestUploaderInfo() (uint, string, string, error)
	VerifyCallback(pubKeyURL, authorization, md5, date string, body []byte) error
}

type UploadService struct {
	db         *gorm.DB
	expireTime int64
	cache      *redis.Client
}

type ConfigStruct struct {
	Expiration string     `json:"expiration"`
	Conditions [][]string `json:"conditions"`
}

type CallbackParam struct {
	CallbackUrl      string `json:"callbackUrl"`
	CallbackBody     string `json:"callbackBody"`
	CallbackBodyType string `json:"callbackBodyType"`
}

type PolicyToken struct {
	AccessKeyId string `json:"accessid"`
	Host        string `json:"host"`
	Expire      int64  `json:"expire"`
	Signature   string `json:"signature"`
	Policy      string `json:"policy"`
	Directory   string `json:"dir"`
	Callback    string `json:"callback"`
}

type CallbackData struct {
	Object   string `form:"object"`
	Size     int64  `form:"size"`
	MimeType string `form:"mimeType"`
	Height   int    `form:"height"`
	Width    int    `form:"width"`
}

func NewUploadService(db *gorm.DB, cache *redis.Client) IUploadService {
	return &UploadService{
		db:         db,
		expireTime: 3000,
		cache:      cache,
	}
}

func (s *UploadService) GetUploadParams(uploadDir string) (map[string]interface{}, error) {
	now := time.Now().Unix()
	expireEnd := now + s.expireTime
	tokenExpire := time.Unix(expireEnd, 0).UTC().Format("2006-01-02T15:04:05Z")

	// Store uploadDir in cache
	cacheKey := fmt.Sprintf("upload:dir:%d", now)
	err := s.cache.Set(s.cache.Context(), cacheKey, uploadDir, time.Duration(s.expireTime)*time.Second).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to cache upload dir: %v", err)
	}

	// Create callback with internal URL
	callbackParam := CallbackParam{
		CallbackUrl:      os.Getenv("CALLBACK_URL"),
		CallbackBody:     "object=${object}&size=${size}&mimeType=${mimeType}&height=${imageInfo.height}&width=${imageInfo.width}",
		CallbackBodyType: "application/x-www-form-urlencoded",
	}
	log.Printf("CallbackParam: %+v\n", callbackParam)

	// Convert callback parameters to JSON
	callbackStr, err := json.Marshal(callbackParam)
	if err != nil {
		return nil, fmt.Errorf("callback JSON serialization failed: %v", err)
	}
	callbackBase64 := base64.StdEncoding.EncodeToString(callbackStr)

	// Create policy
	configStruct := ConfigStruct{
		Expiration: tokenExpire,
		Conditions: [][]string{
			{"starts-with", "$key", uploadDir},
			{"eq", "$success_action_status", "200"},
		},
	}

	// Generate policy
	policyJson, err := json.Marshal(configStruct)
	if err != nil {
		return nil, fmt.Errorf("policy JSON serialization failed: %v", err)
	}
	policyBase64 := base64.StdEncoding.EncodeToString(policyJson)

	// Generate signature
	h := hmac.New(sha1.New, []byte(os.Getenv("ACCESS_KEY_SECRET")))
	io.WriteString(h, policyBase64)
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// Create policy token
	policyToken := PolicyToken{
		AccessKeyId: os.Getenv("ACCESS_KEY_ID"),
		Host:        os.Getenv("HOST"),
		Expire:      expireEnd,
		Signature:   signature,
		Directory:   uploadDir,
		Policy:      policyBase64,
		Callback:    callbackBase64,
	}

	// Convert to map
	tokenJson, err := json.Marshal(policyToken)
	if err != nil {
		return nil, fmt.Errorf("policy token JSON serialization failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(tokenJson, &result); err != nil {
		return nil, fmt.Errorf("failed to convert policy token to map: %v", err)
	}

	return result, nil
}
func (s *UploadService) GetUploadParamsSync(uploadDir string) (map[string]interface{}, error) {
	now := time.Now().Unix()
	expireEnd := now + s.expireTime
	tokenExpire := time.Unix(expireEnd, 0).UTC().Format("2006-01-02T15:04:05Z")

	// Store uploadDir in cache
	cacheKey := fmt.Sprintf("upload:dir:%d", now)
	err := s.cache.Set(s.cache.Context(), cacheKey, uploadDir, time.Duration(s.expireTime)*time.Second).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to cache upload dir: %v", err)
	}

	// Create callback with internal URL
	callbackParam := CallbackParam{
		CallbackUrl:      os.Getenv("CALLBACK_URL_SYNC"),
		CallbackBody:     "object=${object}&size=${size}&mimeType=${mimeType}&height=${imageInfo.height}&width=${imageInfo.width}",
		CallbackBodyType: "application/x-www-form-urlencoded",
	}
	log.Printf("CallbackParam: %+v\n", callbackParam)

	// Convert callback parameters to JSON
	callbackStr, err := json.Marshal(callbackParam)
	if err != nil {
		return nil, fmt.Errorf("callback JSON serialization failed: %v", err)
	}
	callbackBase64 := base64.StdEncoding.EncodeToString(callbackStr)

	// Create policy
	configStruct := ConfigStruct{
		Expiration: tokenExpire,
		Conditions: [][]string{
			{"starts-with", "$key", uploadDir},
			{"eq", "$success_action_status", "200"},
		},
	}

	// Generate policy
	policyJson, err := json.Marshal(configStruct)
	if err != nil {
		return nil, fmt.Errorf("policy JSON serialization failed: %v", err)
	}
	policyBase64 := base64.StdEncoding.EncodeToString(policyJson)

	// Generate signature
	h := hmac.New(sha1.New, []byte(os.Getenv("ACCESS_KEY_SECRET")))
	io.WriteString(h, policyBase64)
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// Create policy token
	policyToken := PolicyToken{
		AccessKeyId: os.Getenv("ACCESS_KEY_ID"),
		Host:        os.Getenv("HOST"),
		Expire:      expireEnd,
		Signature:   signature,
		Directory:   uploadDir,
		Policy:      policyBase64,
		Callback:    callbackBase64,
	}

	// Convert to map
	tokenJson, err := json.Marshal(policyToken)
	if err != nil {
		return nil, fmt.Errorf("policy token JSON serialization failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(tokenJson, &result); err != nil {
		return nil, fmt.Errorf("failed to convert policy token to map: %v", err)
	}

	return result, nil
}

func (s *UploadService) SaveCallbackData(data *CallbackData) error {
	// 将回调数据转为 JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal callback data: %v", err)
	}
	cacheKey := fmt.Sprintf("upload:callback:%d", time.Now().UnixNano())
	// 存储到 Redis，设置 1 小时过期时间
	err = s.cache.Set(s.cache.Context(), cacheKey, string(jsonData), time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to cache callback data: %v", err)
	}
	// 同时更新最新的回调数据
	latestKey := "upload:callback:latest"
	err = s.cache.Set(s.cache.Context(), latestKey, string(jsonData), time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to update latest callback data: %v", err)
	}

	return nil
}

func (s *UploadService) SaveFileNameMapping(newFileName string, dirPath string) error {
	// 存储新文件名
	fileNameKey := fmt.Sprintf("upload:filename:%d", time.Now().UnixNano())
	err := s.cache.Set(s.cache.Context(), fileNameKey, newFileName, 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to cache filename: %v", err)
	}

	// 存储完整路径
	pathKey := fmt.Sprintf("upload:path:%d", time.Now().UnixNano())
	err = s.cache.Set(s.cache.Context(), pathKey, dirPath, 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to cache path: %v", err)
	}

	return nil
}

func (s *UploadService) GetLatestFileName() (string, error) {
	// 获取所有的文件名
	pattern := "upload:filename:*"
	keys, err := s.cache.Keys(s.cache.Context(), pattern).Result()
	if err != nil {
		return "", fmt.Errorf("failed to get filename keys: %v", err)
	}

	if len(keys) == 0 {
		return "", fmt.Errorf("no filename found in cache")
	}

	// 找到最新的键（时间戳最大的）
	latestKey := keys[0]
	latestTimestamp := int64(0)
	for _, key := range keys {
		// 从键中提取时间戳
		parts := strings.Split(key, ":")
		if len(parts) != 3 {
			continue
		}
		timestamp, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			continue
		}
		if timestamp > latestTimestamp {
			latestTimestamp = timestamp
			latestKey = key
		}
	}

	// 获取最新的文件名
	fileName, err := s.cache.Get(s.cache.Context(), latestKey).Result()
	if err != nil {
		return "", fmt.Errorf("failed to get latest filename: %v", err)
	}

	return fileName, nil
}

func (s *UploadService) GetLatestDirPath() (string, error) {
	// 获取所有的路径键
	pattern := "upload:path:*"
	keys, err := s.cache.Keys(s.cache.Context(), pattern).Result()
	if err != nil {
		return "", fmt.Errorf("failed to get path keys: %v", err)
	}

	if len(keys) == 0 {
		return "", fmt.Errorf("no path found in cache")
	}

	// 找到最新的键（时间戳最大的）
	latestKey := keys[0]
	latestTimestamp := int64(0)
	for _, key := range keys {
		// 从键中提取时间戳
		parts := strings.Split(key, ":")
		if len(parts) != 3 {
			continue
		}
		timestamp, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			continue
		}
		if timestamp > latestTimestamp {
			latestTimestamp = timestamp
			latestKey = key
		}
	}

	// 获取最新的路径
	dirPath, err := s.cache.Get(s.cache.Context(), latestKey).Result()
	if err != nil {
		return "", fmt.Errorf("failed to get latest path: %v", err)
	}

	return dirPath, nil
}

func (s *UploadService) SaveUploaderInfo(id uint, uploaderType string, email string) error {
	// Save uploader ID
	idKey := fmt.Sprintf("upload:uploader:id:%d", time.Now().UnixNano())
	err := s.cache.Set(s.cache.Context(), idKey, id, 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to cache uploader id: %v", err)
	}

	// Save uploader type
	typeKey := fmt.Sprintf("upload:uploader:type:%d", time.Now().UnixNano())
	err = s.cache.Set(s.cache.Context(), typeKey, uploaderType, 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to cache uploader type: %v", err)
	}

	// Save uploader email
	emailKey := fmt.Sprintf("upload:uploader:email:%d", time.Now().UnixNano())
	err = s.cache.Set(s.cache.Context(), emailKey, email, 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to cache uploader email: %v", err)
	}

	return nil
}

func (s *UploadService) GetLatestUploaderInfo() (uint, string, string, error) {
	// Get latest uploader ID
	idPattern := "upload:uploader:id:*"
	idKeys, err := s.cache.Keys(s.cache.Context(), idPattern).Result()
	if err != nil {
		return 0, "", "", fmt.Errorf("failed to get uploader id keys: %v", err)
	}
	if len(idKeys) == 0 {
		return 0, "", "", fmt.Errorf("no uploader id found in cache")
	}

	// Find the latest ID key
	latestIDKey := idKeys[0]
	latestTimestamp := int64(0)
	for _, key := range idKeys {
		parts := strings.Split(key, ":")
		if len(parts) != 4 {
			continue
		}
		timestamp, err := strconv.ParseInt(parts[3], 10, 64)
		if err != nil {
			continue
		}
		if timestamp > latestTimestamp {
			latestTimestamp = timestamp
			latestIDKey = key
		}
	}

	// Get ID value
	idStr, err := s.cache.Get(s.cache.Context(), latestIDKey).Result()
	if err != nil {
		return 0, "", "", fmt.Errorf("failed to get uploader id: %v", err)
	}
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0, "", "", fmt.Errorf("failed to parse uploader id: %v", err)
	}

	// Get latest uploader type
	typePattern := "upload:uploader:type:*"
	typeKeys, err := s.cache.Keys(s.cache.Context(), typePattern).Result()
	if err != nil {
		return 0, "", "", fmt.Errorf("failed to get uploader type keys: %v", err)
	}
	if len(typeKeys) == 0 {
		return 0, "", "", fmt.Errorf("no uploader type found in cache")
	}

	// Find the latest type key
	latestTypeKey := typeKeys[0]
	latestTimestamp = int64(0)
	for _, key := range typeKeys {
		parts := strings.Split(key, ":")
		if len(parts) != 4 {
			continue
		}
		timestamp, err := strconv.ParseInt(parts[3], 10, 64)
		if err != nil {
			continue
		}
		if timestamp > latestTimestamp {
			latestTimestamp = timestamp
			latestTypeKey = key
		}
	}

	// Get type value
	uploaderType, err := s.cache.Get(s.cache.Context(), latestTypeKey).Result()
	if err != nil {
		return 0, "", "", fmt.Errorf("failed to get uploader type: %v", err)
	}

	// Get latest uploader email
	emailPattern := "upload:uploader:email:*"
	emailKeys, err := s.cache.Keys(s.cache.Context(), emailPattern).Result()
	if err != nil {
		return 0, "", "", fmt.Errorf("failed to get uploader email keys: %v", err)
	}
	if len(emailKeys) == 0 {
		return 0, "", "", fmt.Errorf("no uploader email found in cache")
	}

	// Find the latest email key
	latestEmailKey := emailKeys[0]
	latestTimestamp = int64(0)
	for _, key := range emailKeys {
		parts := strings.Split(key, ":")
		if len(parts) != 4 {
			continue
		}
		timestamp, err := strconv.ParseInt(parts[3], 10, 64)
		if err != nil {
			continue
		}
		if timestamp > latestTimestamp {
			latestTimestamp = timestamp
			latestEmailKey = key
		}
	}

	// Get email value
	email, err := s.cache.Get(s.cache.Context(), latestEmailKey).Result()
	if err != nil {
		return 0, "", "", fmt.Errorf("failed to get uploader email: %v", err)
	}

	return uint(id), uploaderType, email, nil
}

// VerifyCallback verifies the callback request from OSS
func (s *UploadService) VerifyCallback(pubKeyURL, authorization, md5, date string, body []byte) error {
	// Decode public key URL
	decodedURL, err := base64.StdEncoding.DecodeString(pubKeyURL)
	if err != nil {
		return fmt.Errorf("decode public key url error: %v", err)
	}
	log.Printf("Decoded public key URL: %s\n", string(decodedURL))

	// Get public key content
	resp, err := http.Get(string(decodedURL))
	if err != nil {
		return fmt.Errorf("get public key error: %v", err)
	}
	defer resp.Body.Close()

	publicKey, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read public key error: %v", err)
	}
	log.Printf("Retrieved public key content length: %d\n", len(publicKey))

	// Parse public key
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return fmt.Errorf("failed to parse public key PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("parse public key error: %v", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("public key is not RSA type")
	}

	// Decode authorization
	decodedAuth, err := base64.StdEncoding.DecodeString(authorization)
	if err != nil {
		return fmt.Errorf("decode authorization error: %v", err)
	}

	// Prepare verification content
	strToSign := fmt.Sprintf("%s\n%s\n%s\n", md5, date, string(body))
	log.Printf("String to sign: %s\n", strToSign)

	// Calculate signature
	h := sha1.New()
	h.Write([]byte(strToSign))
	hashed := h.Sum(nil)

	// Verify signature
	err = rsa.VerifyPKCS1v15(rsaPub, crypto.SHA1, hashed, decodedAuth)
	if err != nil {
		return fmt.Errorf("verify signature error: %v", err)
	}

	return nil
}

package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	SettingKeyAnnouncementImageStorage = "announcement_image_storage"
	AnnouncementImageMaxBytes          = 5 << 20
	defaultAnnouncementImagePrefix     = "announcements/"
)

var (
	ErrAnnouncementImageStorageNotConfigured = infraerrors.BadRequest("ANNOUNCEMENT_IMAGE_STORAGE_NOT_CONFIGURED", "announcement image storage is not configured")
	ErrAnnouncementImageTooLarge             = infraerrors.BadRequest("ANNOUNCEMENT_IMAGE_TOO_LARGE", "announcement image must be 5MB or smaller")
	ErrAnnouncementImageTypeUnsupported      = infraerrors.BadRequest("ANNOUNCEMENT_IMAGE_TYPE_UNSUPPORTED", "announcement image type must be png, jpg, jpeg, webp, or gif")
)

type AnnouncementImageStorageConfig struct {
	Endpoint           string `json:"endpoint"`
	Region             string `json:"region"`
	Bucket             string `json:"bucket"`
	AccessKeyID        string `json:"access_key_id"`
	SecretAccessKey    string `json:"secret_access_key,omitempty"`
	HasSecretAccessKey bool   `json:"has_secret_access_key"`
	PublicBaseURL      string `json:"public_base_url"`
	ObjectPrefix       string `json:"object_prefix"`
}

type AnnouncementImageUploadResult struct {
	URL         string `json:"url"`
	Key         string `json:"key"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
}

type AnnouncementImageObjectStore interface {
	Upload(ctx context.Context, key string, body io.Reader, contentType string) error
}

type AnnouncementImageObjectStoreFactory func(ctx context.Context, cfg AnnouncementImageStorageConfig) (AnnouncementImageObjectStore, error)

type AnnouncementImageService struct {
	settingRepo SettingRepository
	storeFactory AnnouncementImageObjectStoreFactory
}

func NewAnnouncementImageService(settingRepo SettingRepository, storeFactory AnnouncementImageObjectStoreFactory) *AnnouncementImageService {
	return &AnnouncementImageService{
		settingRepo:   settingRepo,
		storeFactory: storeFactory,
	}
}

func (s *AnnouncementImageService) GetStorageConfig(ctx context.Context) (AnnouncementImageStorageConfig, error) {
	cfg, err := s.loadStorageConfig(ctx)
	if err != nil {
		if err == ErrSettingNotFound {
			return AnnouncementImageStorageConfig{ObjectPrefix: defaultAnnouncementImagePrefix}, nil
		}
		return AnnouncementImageStorageConfig{}, err
	}
	return maskAnnouncementImageSecret(cfg), nil
}

func (s *AnnouncementImageService) UpdateStorageConfig(ctx context.Context, cfg AnnouncementImageStorageConfig) error {
	if s == nil || s.settingRepo == nil {
		return infraerrors.InternalServer("SETTING_REPOSITORY_UNAVAILABLE", "setting repository is unavailable")
	}
	cfg = normalizeAnnouncementImageStorageConfig(cfg)
	if strings.TrimSpace(cfg.SecretAccessKey) == "" {
		current, err := s.loadStorageConfig(ctx)
		if err == nil {
			cfg.SecretAccessKey = current.SecretAccessKey
		}
	}
	payload, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal announcement image storage config: %w", err)
	}
	return s.settingRepo.Set(ctx, SettingKeyAnnouncementImageStorage, string(payload))
}

func (s *AnnouncementImageService) UploadImage(ctx context.Context, filename, contentType string, body io.Reader, size int64) (*AnnouncementImageUploadResult, error) {
	if s == nil || s.storeFactory == nil {
		return nil, ErrAnnouncementImageStorageNotConfigured
	}
	if size > AnnouncementImageMaxBytes {
		return nil, ErrAnnouncementImageTooLarge
	}
	cfg, err := s.loadStorageConfig(ctx)
	if err != nil {
		if err == ErrSettingNotFound {
			return nil, ErrAnnouncementImageStorageNotConfigured
		}
		return nil, err
	}
	cfg = normalizeAnnouncementImageStorageConfig(cfg)
	if err := validateAnnouncementImageStorageConfig(cfg); err != nil {
		return nil, err
	}
	data, err := io.ReadAll(io.LimitReader(body, AnnouncementImageMaxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read announcement image: %w", err)
	}
	if int64(len(data)) > AnnouncementImageMaxBytes {
		return nil, ErrAnnouncementImageTooLarge
	}
	if isLikelySVG(filename, contentType, data) {
		return nil, ErrAnnouncementImageTypeUnsupported
	}
	if size <= 0 {
		size = int64(len(data))
	}
	contentType = normalizeAnnouncementImageContentType(contentType, filename, data)
	ext, ok := announcementImageExtension(contentType)
	if !ok {
		return nil, ErrAnnouncementImageTypeUnsupported
	}
	key, err := buildAnnouncementImageObjectKey(cfg.ObjectPrefix, ext)
	if err != nil {
		return nil, err
	}
	store, err := s.storeFactory(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create announcement image store: %w", err)
	}
	if err := store.Upload(ctx, key, bytes.NewReader(data), contentType); err != nil {
		return nil, fmt.Errorf("upload announcement image: %w", err)
	}
	return &AnnouncementImageUploadResult{
		URL:         joinPublicObjectURL(cfg.PublicBaseURL, key),
		Key:         key,
		ContentType: contentType,
		Size:        size,
	}, nil
}

func (s *AnnouncementImageService) loadStorageConfig(ctx context.Context) (AnnouncementImageStorageConfig, error) {
	if s == nil || s.settingRepo == nil {
		return AnnouncementImageStorageConfig{}, infraerrors.InternalServer("SETTING_REPOSITORY_UNAVAILABLE", "setting repository is unavailable")
	}
	raw, err := s.settingRepo.GetValue(ctx, SettingKeyAnnouncementImageStorage)
	if err != nil {
		return AnnouncementImageStorageConfig{}, err
	}
	if strings.TrimSpace(raw) == "" {
		return AnnouncementImageStorageConfig{}, ErrSettingNotFound
	}
	var cfg AnnouncementImageStorageConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return AnnouncementImageStorageConfig{}, fmt.Errorf("parse announcement image storage config: %w", err)
	}
	return normalizeAnnouncementImageStorageConfig(cfg), nil
}

func normalizeAnnouncementImageStorageConfig(cfg AnnouncementImageStorageConfig) AnnouncementImageStorageConfig {
	cfg.Endpoint = strings.TrimSpace(cfg.Endpoint)
	cfg.Region = strings.TrimSpace(cfg.Region)
	cfg.Bucket = strings.TrimSpace(cfg.Bucket)
	cfg.AccessKeyID = strings.TrimSpace(cfg.AccessKeyID)
	cfg.SecretAccessKey = strings.TrimSpace(cfg.SecretAccessKey)
	cfg.PublicBaseURL = strings.TrimRight(strings.TrimSpace(cfg.PublicBaseURL), "/")
	cfg.ObjectPrefix = strings.Trim(strings.TrimSpace(cfg.ObjectPrefix), "/")
	if cfg.ObjectPrefix == "" {
		cfg.ObjectPrefix = strings.Trim(defaultAnnouncementImagePrefix, "/")
	}
	cfg.ObjectPrefix += "/"
	cfg.HasSecretAccessKey = cfg.SecretAccessKey != ""
	return cfg
}

func validateAnnouncementImageStorageConfig(cfg AnnouncementImageStorageConfig) error {
	if cfg.Endpoint == "" || cfg.Region == "" || cfg.Bucket == "" || cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" || cfg.PublicBaseURL == "" {
		return ErrAnnouncementImageStorageNotConfigured
	}
	return nil
}

func maskAnnouncementImageSecret(cfg AnnouncementImageStorageConfig) AnnouncementImageStorageConfig {
	cfg.HasSecretAccessKey = strings.TrimSpace(cfg.SecretAccessKey) != ""
	cfg.SecretAccessKey = ""
	return cfg
}

func normalizeAnnouncementImageContentType(contentType, filename string, data []byte) string {
	contentType = strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	if contentType != "" {
		return contentType
	}
	if ext := strings.ToLower(path.Ext(filename)); ext != "" {
		if guessed := mime.TypeByExtension(ext); guessed != "" {
			return strings.ToLower(strings.Split(guessed, ";")[0])
		}
	}
	return strings.ToLower(strings.Split(http.DetectContentType(data), ";")[0])
}

func isLikelySVG(filename, contentType string, data []byte) bool {
	if strings.EqualFold(path.Ext(filename), ".svg") {
		return true
	}
	if strings.EqualFold(strings.TrimSpace(strings.Split(contentType, ";")[0]), "image/svg+xml") {
		return true
	}
	head := strings.ToLower(string(bytes.TrimSpace(data)))
	if len(head) > 512 {
		head = head[:512]
	}
	return strings.Contains(head, "<svg") || strings.Contains(head, "<!doctype svg")
}

func announcementImageExtension(contentType string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "image/png":
		return ".png", true
	case "image/jpeg", "image/jpg":
		return ".jpg", true
	case "image/webp":
		return ".webp", true
	case "image/gif":
		return ".gif", true
	default:
		return "", false
	}
}

func buildAnnouncementImageObjectKey(prefix, ext string) (string, error) {
	var random [16]byte
	if _, err := rand.Read(random[:]); err != nil {
		return "", fmt.Errorf("generate image key: %w", err)
	}
	random[6] = (random[6] & 0x0f) | 0x40
	random[8] = (random[8] & 0x3f) | 0x80
	now := time.Now().UTC()
	cleanPrefix := strings.Trim(strings.TrimSpace(prefix), "/")
	if cleanPrefix == "" {
		cleanPrefix = strings.Trim(defaultAnnouncementImagePrefix, "/")
	}
	id := hex.EncodeToString(random[:])
	uuid := fmt.Sprintf("%s-%s-%s-%s-%s", id[0:8], id[8:12], id[12:16], id[16:20], id[20:32])
	return fmt.Sprintf("%s/%04d/%02d/%s%s", cleanPrefix, now.Year(), int(now.Month()), uuid, ext), nil
}

func joinPublicObjectURL(baseURL, key string) string {
	return strings.TrimRight(strings.TrimSpace(baseURL), "/") + "/" + strings.TrimLeft(key, "/")
}

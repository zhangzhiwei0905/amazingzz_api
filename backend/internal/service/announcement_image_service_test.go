package service

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type announcementImageSettingRepoStub struct {
	values map[string]string
}

func newAnnouncementImageSettingRepoStub() *announcementImageSettingRepoStub {
	return &announcementImageSettingRepoStub{values: map[string]string{}}
}

func (s *announcementImageSettingRepoStub) Get(context.Context, string) (*Setting, error) {
	return nil, ErrSettingNotFound
}

func (s *announcementImageSettingRepoStub) GetValue(_ context.Context, key string) (string, error) {
	if value, ok := s.values[key]; ok {
		return value, nil
	}
	return "", ErrSettingNotFound
}

func (s *announcementImageSettingRepoStub) Set(_ context.Context, key, value string) error {
	s.values[key] = value
	return nil
}

func (s *announcementImageSettingRepoStub) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	out := map[string]string{}
	for _, key := range keys {
		out[key] = s.values[key]
	}
	return out, nil
}

func (s *announcementImageSettingRepoStub) SetMultiple(_ context.Context, values map[string]string) error {
	for key, value := range values {
		s.values[key] = value
	}
	return nil
}

func (s *announcementImageSettingRepoStub) GetAll(context.Context) (map[string]string, error) {
	out := map[string]string{}
	for key, value := range s.values {
		out[key] = value
	}
	return out, nil
}

func (s *announcementImageSettingRepoStub) Delete(_ context.Context, key string) error {
	delete(s.values, key)
	return nil
}

type announcementImageStoreStub struct {
	key         string
	contentType string
	data        []byte
}

func (s *announcementImageStoreStub) Upload(ctx context.Context, key string, body io.Reader, contentType string) error {
	data, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	s.key = key
	s.contentType = contentType
	s.data = data
	return ctx.Err()
}

func TestAnnouncementImageServiceUploadStoresImageAndReturnsPublicURL(t *testing.T) {
	repo := newAnnouncementImageSettingRepoStub()
	store := &announcementImageStoreStub{}
	svc := NewAnnouncementImageService(repo, func(context.Context, AnnouncementImageStorageConfig) (AnnouncementImageObjectStore, error) {
		return store, nil
	})
	require.NoError(t, svc.UpdateStorageConfig(context.Background(), AnnouncementImageStorageConfig{
		Endpoint:        "https://oss-cn-hangzhou.aliyuncs.com",
		Region:          "cn-hangzhou",
		Bucket:          "sub2api",
		AccessKeyID:     "ak",
		SecretAccessKey: "sk",
		PublicBaseURL:   "https://cdn.example.com/assets",
		ObjectPrefix:    "announcements/",
	}))

	result, err := svc.UploadImage(context.Background(), "clipboard.png", "image/png", bytes.NewReader([]byte("png-bytes")), int64(len("png-bytes")))
	require.NoError(t, err)
	require.Equal(t, "image/png", store.contentType)
	require.Equal(t, []byte("png-bytes"), store.data)
	require.True(t, strings.HasPrefix(store.key, "announcements/"))
	require.True(t, strings.HasSuffix(store.key, ".png"))
	require.Equal(t, "https://cdn.example.com/assets/"+store.key, result.URL)
	require.Equal(t, store.key, result.Key)
}

func TestAnnouncementImageServiceRejectsUnsupportedImages(t *testing.T) {
	repo := newAnnouncementImageSettingRepoStub()
	svc := NewAnnouncementImageService(repo, func(context.Context, AnnouncementImageStorageConfig) (AnnouncementImageObjectStore, error) {
		return &announcementImageStoreStub{}, nil
	})
	require.NoError(t, svc.UpdateStorageConfig(context.Background(), AnnouncementImageStorageConfig{
		Endpoint:        "https://oss-cn-hangzhou.aliyuncs.com",
		Region:          "cn-hangzhou",
		Bucket:          "sub2api",
		AccessKeyID:     "ak",
		SecretAccessKey: "sk",
		PublicBaseURL:   "https://cdn.example.com",
		ObjectPrefix:    "announcements/",
	}))

	_, err := svc.UploadImage(context.Background(), "unsafe.svg", "image/svg+xml", strings.NewReader("<svg></svg>"), int64(len("<svg></svg>")))
	require.Error(t, err)
	require.Contains(t, err.Error(), "image type")
}

func TestAnnouncementImageServiceMasksStoredSecret(t *testing.T) {
	repo := newAnnouncementImageSettingRepoStub()
	svc := NewAnnouncementImageService(repo, nil)
	require.NoError(t, svc.UpdateStorageConfig(context.Background(), AnnouncementImageStorageConfig{
		Endpoint:        "https://oss-cn-hangzhou.aliyuncs.com",
		Region:          "cn-hangzhou",
		Bucket:          "sub2api",
		AccessKeyID:     "ak",
		SecretAccessKey: "sk",
		PublicBaseURL:   "https://cdn.example.com",
		ObjectPrefix:    "announcements/",
	}))

	cfg, err := svc.GetStorageConfig(context.Background())
	require.NoError(t, err)
	require.Equal(t, "", cfg.SecretAccessKey)
	require.True(t, cfg.HasSecretAccessKey)
}

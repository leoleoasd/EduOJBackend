package utils

import (
	"bytes"
	"context"
	"github.com/EduOJ/backend/base"
	"github.com/minio/minio-go/v7"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strconv"
	"testing"
)

func getPresignedURLContent(t *testing.T, presignedUrl string) (content string) {
	resp, err := http.Get(presignedUrl)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	length, err := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	assert.NoError(t, err)
	body := make([]byte, length)
	_, err = resp.Body.Read(body)
	return string(body)
}

func TestMustCreateBucket(t *testing.T) {
	t.Run("testMustCreateBucketExistingBucket", func(t *testing.T) {
		assert.NoError(t, base.Storage.MakeBucket(context.Background(), "existing-bucket", minio.MakeBucketOptions{
			Region: viper.GetString("storage.region"),
		}))
		found, err := base.Storage.BucketExists(context.Background(), "existing-bucket")
		assert.True(t, found)
		assert.NoError(t, err)
		assert.NoError(t, CreateBucket("existing-bucket"))
		found, err = base.Storage.BucketExists(context.Background(), "existing-bucket")
		assert.True(t, found)
		assert.NoError(t, err)
	})
	t.Run("testMustCreateBucketNonExistingBucket", func(t *testing.T) {
		found, err := base.Storage.BucketExists(context.Background(), "non-existing-bucket")
		assert.False(t, found)
		assert.NoError(t, err)
		assert.NoError(t, CreateBucket("non-existing-bucket"))
		found, err = base.Storage.BucketExists(context.Background(), "non-existing-bucket")
		assert.True(t, found)
		assert.NoError(t, err)
	})
}

func TestGetPresignedURL(t *testing.T) {
	b := []byte("test_get_presigned_url")
	reader := bytes.NewReader(b)
	info, err := base.Storage.PutObject(context.Background(), "test-bucket", "test_get_presigned_url_object", reader, int64(len(b)), minio.PutObjectOptions{})
	assert.NoError(t, err)
	assert.Equal(t, int64(len(b)), info.Size)
	presignedUrl, err := GetPresignedURL("test-bucket", "test_get_presigned_url_object", "test_get_presigned_url_file_name")
	assert.NoError(t, err)
	resp, err := http.Get(presignedUrl)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, string(b), getPresignedURLContent(t, presignedUrl))
}

package controller_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"github.com/EduOJ/backend/app/response"
	"github.com/EduOJ/backend/base"
	"github.com/EduOJ/backend/base/utils"
	"github.com/EduOJ/backend/database/models"
	"github.com/minio/minio-go/v7"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetImage(t *testing.T) {
	t.Parallel()
	// base64_encoded image for testing.
	imageBlob := "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"
	imageBuffer := bytes.Buffer{}
	_, err := io.Copy(&imageBuffer, base64.NewDecoder(base64.StdEncoding, strings.NewReader(imageBlob)))
	assert.NoError(t, err)
	imageBytes := imageBuffer.Bytes()
	{
		// Write image to database and storage.
		fileModel := models.Image{
			Filename: "eduoj_test.png",
			FilePath: "test_image_path",
		}
		utils.PanicIfDBError(base.DB.Save(&fileModel), "could not save image")

		nonExistingModel := models.Image{
			Filename: "eduoj_test.png",
			FilePath: "test_non_exiting",
		}
		utils.PanicIfDBError(base.DB.Save(&nonExistingModel), "could not save image")

		illegalTypeModel := models.Image{
			Filename: "eduoj_test.png",
			FilePath: "test_illegal_type",
		}
		utils.PanicIfDBError(base.DB.Save(&illegalTypeModel), "could not save image")

		found, err := base.Storage.BucketExists(context.Background(), "images")
		if err != nil {
			panic(errors.Wrap(err, "could not query if bucket exists"))
		}
		if !found {
			err = base.Storage.MakeBucket(context.Background(), "images", minio.MakeBucketOptions{
				Region: viper.GetString("storage.region"),
			})
			if err != nil {
				panic(errors.Wrap(err, "could not query if bucket exists"))
			}
		}
		_, err = base.Storage.PutObject(context.Background(), "images", "test_image_path", &imageBuffer, int64(imageBuffer.Len()), minio.PutObjectOptions{
			ContentType: "image/png",
		})
		if err != nil {
			panic(errors.Wrap(err, "could write image to s3 storage."))
		}
	}

	// test found
	httpResp := makeResp(makeReq(t, "GET", base.Echo.Reverse("image.getImage", "test_image_path"), nil))
	respbytes, err := ioutil.ReadAll(httpResp.Body)
	assert.NoError(t, err)
	assert.Equal(t, 200, httpResp.StatusCode)
	assert.Equal(t, imageBytes, []byte(respbytes))

	// test not found
	httpResp = makeResp(makeReq(t, "GET", base.Echo.Reverse("image.getImage", "-1"), nil))
	assert.Equal(t, 404, httpResp.StatusCode)
	jsonEQ(t, response.Response{
		Message: "IMAGE_NOT_FOUND",
		Error:   nil,
		Data:    nil,
	}, httpResp)

	// test not found
	httpResp = makeResp(makeReq(t, "GET", base.Echo.Reverse("image.getImage", "test_non_exiting"), nil))
	assert.Equal(t, 404, httpResp.StatusCode)
	jsonEQ(t, response.Response{
		Message: "IMAGE_NOT_FOUND",
		Error:   nil,
		Data:    nil,
	}, httpResp)
}

func TestCreateImage(t *testing.T) {
	t.Parallel()
	// base64_encoded image for testing.
	b64 := "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"

	t.Run("testCreateImageNoFile", func(t *testing.T) {
		t.Parallel()
		var b bytes.Buffer
		w := multipart.NewWriter(&b)
		assert.NoError(t, w.Close())
		req := httptest.NewRequest("POST", base.Echo.Reverse("image.createImage"), &b)
		req.Header.Set("Set-User-For-Test", applyNormalUser["Set-User-For-Test"][0])
		req.Header.Set("Content-Type", w.FormDataContentType())
		httpResp := makeResp(req)
		resp := response.CreateImageResponse{}
		mustJsonDecode(httpResp, &resp)
		assert.Equal(t, "INTERNAL_ERROR", resp.Message)
		assert.Equal(t, 500, httpResp.StatusCode)
	})

	t.Run("testCreateImageIllegalType", func(t *testing.T) {
		t.Parallel()
		b64 := `aaaaaaaa`
		var b bytes.Buffer
		w := multipart.NewWriter(&b)
		fw, err := w.CreateFormFile("file", "baidu.png")
		assert.NoError(t, err)
		_, err = io.Copy(fw, base64.NewDecoder(base64.StdEncoding, strings.NewReader(b64)))
		assert.NoError(t, err)
		assert.NoError(t, w.Close())
		req := httptest.NewRequest("POST", base.Echo.Reverse("image.createImage"), &b)
		req.Header.Set("Set-User-For-Test", applyNormalUser["Set-User-For-Test"][0])
		req.Header.Set("Content-Type", w.FormDataContentType())
		httpResp := makeResp(req)
		resp := response.CreateImageResponse{}
		mustJsonDecode(httpResp, &resp)
		assert.Equal(t, "ILLEGAL_TYPE", resp.Message)
		assert.Nil(t, resp.Error)
		assert.Equal(t, 403, httpResp.StatusCode)
	})

	t.Run("testCreateImageIllegalFileExtension", func(t *testing.T) {
		t.Parallel()
		var b bytes.Buffer
		w := multipart.NewWriter(&b)
		fw, err := w.CreateFormFile("file", "baidu.jpg")
		assert.NoError(t, err)
		_, err = io.Copy(fw, base64.NewDecoder(base64.StdEncoding, strings.NewReader(b64)))
		assert.NoError(t, err)
		assert.NoError(t, w.Close())
		req := httptest.NewRequest("POST", base.Echo.Reverse("image.createImage"), &b)
		req.Header.Set("Set-User-For-Test", applyNormalUser["Set-User-For-Test"][0])
		req.Header.Set("Content-Type", w.FormDataContentType())
		httpResp := makeResp(req)
		resp := response.CreateImageResponse{}
		mustJsonDecode(httpResp, &resp)
		assert.Equal(t, "ILLEGAL_TYPE", resp.Message)
		assert.Nil(t, resp.Error)
		assert.Equal(t, 403, httpResp.StatusCode)
	})

	t.Run("testCreateImageSuccess", func(t *testing.T) {
		t.Parallel()
		var b bytes.Buffer
		w := multipart.NewWriter(&b)
		fw, err := w.CreateFormFile("file", "baidu.png")
		assert.NoError(t, err)
		_, err = io.Copy(fw, base64.NewDecoder(base64.StdEncoding, strings.NewReader(b64)))
		assert.NoError(t, err)
		assert.NoError(t, w.Close())
		req := httptest.NewRequest("POST", base.Echo.Reverse("image.createImage"), &b)
		req.Header.Set("Set-User-For-Test", applyNormalUser["Set-User-For-Test"][0])
		req.Header.Set("Content-Type", w.FormDataContentType())
		httpResp := makeResp(req)
		resp := response.CreateImageResponse{}
		mustJsonDecode(httpResp, &resp)
		assert.Equal(t, 201, httpResp.StatusCode)
		assert.Equal(t, "SUCCESS", resp.Message)
		assert.Nil(t, resp.Error)
		sepPath := strings.Split(resp.Data.FilePath, "/")
		assert.Equal(t, base.Echo.Reverse("image.createImage"), resp.Data.FilePath[:len(base.Echo.Reverse("image.createImage"))])
		assert.Equal(t, 4, len(sepPath))

		imageModel := models.Image{}
		base.DB.Model(models.Image{}).Where("file_path = ?", sepPath[len(sepPath)-1]).Find(&imageModel)
		assert.Equal(t, imageModel.Filename, "baidu.png")
		o, err := base.Storage.GetObject(context.Background(), "images", sepPath[len(sepPath)-1], minio.GetObjectOptions{})
		assert.NoError(t, err)
		buf := bytes.Buffer{}
		_, err = io.Copy(&buf, o)
		assert.NoError(t, err)
		imgBuf := bytes.Buffer{}
		_, err = io.Copy(&imgBuf, base64.NewDecoder(base64.StdEncoding, strings.NewReader(b64)))
		assert.NoError(t, err)
		assert.Equal(t, imgBuf.Bytes(), buf.Bytes())
	})
}

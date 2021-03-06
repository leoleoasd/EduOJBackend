package controller

import (
	"context"
	"fmt"
	"github.com/EduOJ/backend/app/response"
	"github.com/EduOJ/backend/base"
	"github.com/EduOJ/backend/base/log"
	"github.com/EduOJ/backend/base/utils"
	"github.com/EduOJ/backend/database/models"
	"github.com/gabriel-vasile/mimetype"
	"github.com/labstack/echo/v4"
	"github.com/minio/minio-go/v7"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"io"
	"net/http"
	path2 "path"
	"strings"
)

func GetImage(c echo.Context) error {
	// TODO: check referrer
	id := c.Param("id")
	imageModel := models.Image{}
	err := base.DB.Model(&models.Image{}).Where("file_path = ?", id).First(&imageModel).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.JSON(http.StatusNotFound, response.ErrorResp("IMAGE_NOT_FOUND", nil))
	} else if err != nil {
		panic(err)
	}
	object, err := base.Storage.GetObject(context.Background(), "images", imageModel.FilePath, minio.GetObjectOptions{})
	if err != nil {
		panic(err)
	}
	mime, err := mimetype.DetectReader(object)
	if err != nil {
		if merr, ok := err.(minio.ErrorResponse); ok {
			if merr.StatusCode == 404 {
				return c.JSON(http.StatusNotFound, response.ErrorResp("IMAGE_NOT_FOUND", nil))
			} else {
				panic(merr)
			}
		}
		log.Error("could not detect MIME of image!")
		log.Error(err)
		return c.JSON(http.StatusForbidden, response.ErrorResp("ILLEGAL_TYPE", nil))
	}
	_, err = object.Seek(0, io.SeekStart)
	if err != nil {
		panic(err)
	}

	c.Response().Header().Set("Access-Control-Allow-Origin", strings.Join(utils.Origins, ", "))
	c.Response().Header().Set("Cache-Control", "public; max-age=31536000")
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, imageModel.Filename))
	return c.Stream(http.StatusOK, mime.String(), object)
}

func CreateImage(c echo.Context) error {
	user, ok := c.Get("user").(models.User)
	if !ok {
		panic("could not convert my user into type models.User")
	}
	file, err := c.FormFile("file")
	if err != nil {
		panic(err)
	}
	count := int64(1)
	fileIndex := ""
	for count != 0 {
		fileIndex = utils.RandStr(32)
		utils.PanicIfDBError(base.DB.Model(&models.Image{}).Where("file_path = ?", fileIndex).Count(&count), "could not save image")
	}
	fileModel := models.Image{
		Filename: file.Filename,
		FilePath: fileIndex,
		User:     user,
	}
	utils.PanicIfDBError(base.DB.Save(&fileModel), "could not save image")
	src, err := file.Open()
	if err != nil {
		panic(err)
	}
	defer src.Close()
	mime, err := mimetype.DetectReader(src)
	if err != nil {
		log.Error("could not detect MIME of image!")
		log.Error(err)
		return c.JSON(http.StatusForbidden, response.ErrorResp("ILLEGAL_TYPE", nil))
	}
	if mime.String()[:5] != "image" || mime.Extension() != path2.Ext(file.Filename) {
		return c.JSON(http.StatusForbidden, response.ErrorResp("ILLEGAL_TYPE", nil))
	}
	_, err = src.Seek(0, io.SeekStart)
	if err != nil {
		panic(errors.Wrap(err, "could not seek to file start"))
	}

	_, err = base.Storage.PutObject(c.Request().Context(), "images", fileIndex, src, file.Size, minio.PutObjectOptions{
		ContentType: mime.String(),
	})
	if err != nil {
		panic(errors.Wrap(err, "could write image to s3 storage."))
	}
	return c.JSON(http.StatusCreated, response.CreateImageResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			FilePath string `json:"filename"`
		}{
			base.Echo.Reverse("image.getImage", fileIndex),
		},
	})
}

package utils

import (
	"context"
	"github.com/EduOJ/backend/app/response"
	"github.com/EduOJ/backend/base"
	"github.com/EduOJ/backend/base/log"
	validator2 "github.com/EduOJ/backend/base/validator"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/minio/minio-go/v7"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"mime/multipart"
	"net/http"
)

func PanicIfDBError(db *gorm.DB, message string) {
	if db.Error != nil {
		panic(errors.Wrap(db.Error, message))
	}
}

func BindAndValidate(req interface{}, c echo.Context) (err error, ok bool) {
	// TODO: return a HttpError instead of writing to response.
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResp("BAD_REQUEST_PARAMETER", nil)), false
	}
	if err := c.Validate(req); err != nil {
		if e, ok := err.(validator.ValidationErrors); ok {
			validationErrors := make([]response.ValidationError, len(e))
			for i, v := range e {
				validationErrors[i] = response.ValidationError{
					Field:       v.Field(),
					Reason:      v.Tag(),
					Translation: v.Translate(validator2.Trans),
				}
			}
			return c.JSON(http.StatusBadRequest, response.ErrorResp("VALIDATION_ERROR", validationErrors)), false
		}
		log.Error(errors.Wrap(err, "validate failed"), c)
		return response.InternalErrorResp(c), false
	}
	return nil, true
}

func MustPutObject(object *multipart.FileHeader, ctx context.Context, bucket string, path string) {
	// TODO: use reader
	src, err := object.Open()
	if err != nil {
		panic(err)
	}
	defer src.Close()
	_, err = base.Storage.PutObject(ctx, bucket, path, src, object.Size, minio.PutObjectOptions{})
	if err != nil {
		panic(errors.Wrap(err, "could write file to s3 storage."))
	}
}

func MustGetObject(c echo.Context, bucket string, path string) *minio.Object {
	object, err := base.Storage.GetObject(c.Request().Context(), bucket, path, minio.GetObjectOptions{})
	if err != nil {
		panic(err)
	}
	return object
}

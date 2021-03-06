package controller

import (
	"github.com/EduOJ/backend/app/request"
	"github.com/EduOJ/backend/app/response"
	"github.com/EduOJ/backend/app/response/resource"
	"github.com/EduOJ/backend/base"
	"github.com/EduOJ/backend/base/utils"
	"github.com/EduOJ/backend/database/models"
	"github.com/labstack/echo/v4"
	"github.com/minio/minio-go/v7"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"net/http"
)

func JudgerGetScript(c echo.Context) error {
	script := models.Script{}
	name := c.Param("name")
	err := base.DB.First(&script, "name = ?", name).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
		} else {
			panic(errors.Wrap(err, "could not query script"))
		}
	}
	url, err := utils.GetPresignedURL("scripts", script.Name, script.Filename)
	if err != nil {
		panic(errors.Wrap(err, "could not get presigned url of script"))
	}
	return c.Redirect(http.StatusFound, url)
}

func CreateScript(c echo.Context) error {
	req := request.CreateScriptRequest{}
	err, ok := utils.BindAndValidate(&req, c)
	if !ok {
		return err
	}

	file, err := c.FormFile("file")
	if err != nil && err != http.ErrMissingFile && err.Error() != "request Content-Type isn't multipart/form-data" {
		panic(errors.Wrap(err, "could not read file"))
	}
	if file == nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResp("INVALID_FILE", nil))
	}

	script := models.Script{
		Name:     req.Name,
		Filename: file.Filename,
	}
	utils.PanicIfDBError(base.DB.Create(&script), "could not create script for creating script")

	utils.MustPutObject(file, c.Request().Context(), "scripts", req.Name)

	return c.JSON(http.StatusCreated, response.CreateScriptResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			*resource.Script `json:"script"`
		}{
			resource.GetScript(&script),
		},
	})
}

func GetScript(c echo.Context) error {
	script := models.Script{}
	if err := base.DB.First(&script, "name = ?", c.Param("name")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
		}
		panic(errors.Wrap(err, "could not find script for getting script"))
	}
	return c.JSON(http.StatusOK, response.GetScriptResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			*resource.Script `json:"script"`
		}{
			resource.GetScript(&script),
		},
	})
}

func GetScriptFile(c echo.Context) error {
	script := models.Script{}
	if err := base.DB.First(&script, "name = ?", c.Param("name")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
		}
		panic(errors.Wrap(err, "could not find script for getting script file"))
	}
	url, err := utils.GetPresignedURL("scripts", script.Name, script.Filename)
	if err != nil {
		panic(errors.Wrap(err, "could not get presigned url of script for getting script file"))
	}
	return c.Redirect(http.StatusFound, url)
}

func GetScripts(c echo.Context) error {
	req := request.GetScriptsRequest{}
	err, ok := utils.BindAndValidate(&req, c)
	if !ok {
		return err
	}
	query := base.DB.Model(&models.Script{}).Order("created_at DESC") // Force order by created_at desc.
	var scripts []*models.Script
	total, prevUrl, nextUrl, err := utils.Paginator(query, req.Limit, req.Offset, c.Request().URL, &scripts)
	if err != nil {
		if herr, ok := err.(utils.HttpError); ok {
			return herr.Response(c)
		}
		panic(err)
	}
	return c.JSON(http.StatusOK, response.GetScriptsResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			Scripts []*resource.Script `json:"scripts"`
			Total   int                `json:"total"`
			Count   int                `json:"count"`
			Offset  int                `json:"offset"`
			Prev    *string            `json:"prev"`
			Next    *string            `json:"next"`
		}{
			Scripts: resource.GetScriptSlice(scripts),
			Total:   total,
			Count:   len(scripts),
			Offset:  req.Offset,
			Prev:    prevUrl,
			Next:    nextUrl,
		},
	})
}

func UpdateScript(c echo.Context) error {
	req := request.UpdateScriptRequest{}
	err, ok := utils.BindAndValidate(&req, c)
	if !ok {
		return err
	}
	script := models.Script{}
	if err := base.DB.First(&script, "name = ?", c.Param("name")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
		}
		panic(errors.Wrap(err, "could not find script for updating script"))
	}

	file, err := c.FormFile("file")
	if err != nil && err != http.ErrMissingFile && err.Error() != "request Content-Type isn't multipart/form-data" {
		panic(errors.Wrap(err, "could not read file"))
	}
	if script.Name != req.Name {
		if file == nil {
			_, err = base.Storage.CopyObject(c.Request().Context(), minio.CopyDestOptions{
				Bucket: "scripts",
				Object: req.Name,
			}, minio.CopySrcOptions{
				Bucket: "scripts",
				Object: script.Name,
			})
			if err != nil {
				panic(errors.Wrap(err, "could not copy object for updating script"))
			}
		}
		if err := base.Storage.RemoveObject(c.Request().Context(), "scripts", script.Name, minio.RemoveObjectOptions{}); err != nil {
			panic(errors.Wrap(err, "could not remove object for updating script"))
		}
		script.Name = req.Name
	}
	if file != nil {
		script.Filename = file.Filename
		utils.MustPutObject(file, c.Request().Context(), "scripts", req.Name)
	}
	utils.PanicIfDBError(base.DB.Save(&script), "could not save script for updating script")

	return c.JSON(http.StatusOK, response.CreateScriptResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			*resource.Script `json:"script"`
		}{
			resource.GetScript(&script),
		},
	})
}

func DeleteScript(c echo.Context) error {
	script := models.Script{}
	if err := base.DB.First(&script, "name = ?", c.Param("name")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusOK, response.Response{
				Message: "SUCCESS",
				Error:   nil,
				Data:    nil,
			})
		}
		panic(errors.Wrap(err, "could not find script for deleting script"))
	}
	var languageCount, problemCount int64
	utils.PanicIfDBError(base.DB.Find(&models.Language{}, "build_script_name = ? or run_script_name = ?", script.Name, script.Name).
		Count(&languageCount), "could not get count of language for deleting script")
	utils.PanicIfDBError(base.DB.Find(&models.Problem{}, "compare_script_name = ?", script.Name).
		Count(&problemCount), "could not get count of language for deleting script")
	if languageCount > 0 || problemCount > 0 {
		return c.JSON(http.StatusBadRequest, response.ErrorResp("SCRIPT_IN_USE", nil))
	}
	utils.PanicIfDBError(base.DB.Delete(&script), "could not delete script for deleting script")
	return c.JSON(http.StatusOK, response.Response{
		Message: "SUCCESS",
		Error:   nil,
		Data:    nil,
	})
}

package controller

import (
	"github.com/EduOJ/backend/app/request"
	"github.com/EduOJ/backend/app/response"
	"github.com/EduOJ/backend/app/response/resource"
	"github.com/EduOJ/backend/base"
	"github.com/EduOJ/backend/base/utils"
	"github.com/EduOJ/backend/database/models"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"net/http"
)

func CreateClass(c echo.Context) error {
	user := c.Get("user").(models.User)
	req := request.CreateClassRequest{}
	err, ok := utils.BindAndValidate(&req, c)
	if !ok {
		return err
	}
	class := models.Class{
		Name:        req.Name,
		CourseName:  req.CourseName,
		Description: req.Description,
		InviteCode:  utils.GenerateInviteCode(),
		Managers: []*models.User{
			&user,
		},
		Students: []*models.User{},
	}
	utils.PanicIfDBError(base.DB.Create(&class), "could not create class")
	user.GrantRole("class_creator", class)
	return c.JSON(http.StatusCreated, response.CreateClassResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			*resource.ClassDetail `json:"class"`
		}{
			resource.GetClassDetail(&class),
		},
	})
}

func GetClass(c echo.Context) error {
	class := models.Class{}
	if err := base.DB.Preload("Managers").Preload("Students").First(&class, c.Param("id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
		}
		panic(errors.Wrap(err, "could not get class while getting class"))
	}
	user := c.Get("user").(models.User)
	if user.Can("read_class_secrets", class) || user.Can("read_class_secrets") {
		return c.JSON(http.StatusOK, response.GetClassResponseForAdmin{
			Message: "SUCCESS",
			Error:   nil,
			Data: struct {
				*resource.ClassDetail `json:"class"`
			}{
				resource.GetClassDetail(&class),
			},
		})
	} else {
		count := base.DB.Model(&class).Where("id = ?", user.ID).Association("Students").Count()
		if count > 0 {
			return c.JSON(http.StatusOK, response.GetClassResponse{
				Message: "SUCCESS",
				Error:   nil,
				Data: struct {
					*resource.Class `json:"class"`
				}{
					resource.GetClass(&class),
				},
			})
		} else {
			return c.JSON(http.StatusForbidden, response.ErrorResp("PERMISSION_DENIED", nil))
		}
	}
}

func GetClassesIManage(c echo.Context) error {
	user := c.Get("user").(models.User)
	var classes []models.Class
	if err := base.DB.Model(&user).Association("ClassesManaging").Find(&classes); err != nil {
		panic(errors.Wrap(err, "could not find class managing"))
	}
	return c.JSON(http.StatusOK, response.GetClassesIManageResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			Classes []resource.Class `json:"classes"`
		}{
			Classes: resource.GetClassSlice(classes),
		},
	})
}

func GetClassesITake(c echo.Context) error {
	user := c.Get("user").(models.User)
	var classes []models.Class
	if err := base.DB.Model(&user).Preload("ProblemSets").Association("ClassesTaking").Find(&classes); err != nil {
		panic(errors.Wrap(err, "could not find class taking"))
	}
	return c.JSON(http.StatusOK, response.GetClassesITakeResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			Classes []resource.Class `json:"classes"`
		}{
			Classes: resource.GetClassSlice(classes),
		},
	})
}

func UpdateClass(c echo.Context) error {
	req := request.UpdateClassRequest{}
	err, ok := utils.BindAndValidate(&req, c)
	if !ok {
		return err
	}
	class := models.Class{}
	if err := base.DB.Preload("Managers").Preload("Students").First(&class, c.Param("id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
		} else {
			panic(errors.Wrap(err, "could not find class for updating"))
		}
	}
	class.Name = req.Name
	class.CourseName = req.CourseName
	class.Description = req.Description
	utils.PanicIfDBError(base.DB.Save(&class), "could not update class")
	return c.JSON(http.StatusOK, response.UpdateClassResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			*resource.ClassDetail `json:"class"`
		}{
			resource.GetClassDetail(&class),
		},
	})
}

func RefreshInviteCode(c echo.Context) error {
	class := models.Class{}
	if err := base.DB.Preload("Managers").Preload("Students").First(&class, c.Param("id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
		} else {
			panic(errors.Wrap(err, "could not find class for refreshing invite code"))
		}
	}
	class.InviteCode = utils.GenerateInviteCode()
	utils.PanicIfDBError(base.DB.Save(&class), "could not update class for refreshing invite code")
	return c.JSON(http.StatusOK, response.RefreshInviteCodeResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			*resource.ClassDetail `json:"class"`
		}{
			resource.GetClassDetail(&class),
		},
	})
}

func AddStudents(c echo.Context) error {
	req := request.AddStudentsRequest{}
	err, ok := utils.BindAndValidate(&req, c)
	if !ok {
		return err
	}
	class := models.Class{}
	if err := base.DB.Preload("Managers").Preload("Students").First(&class, c.Param("id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
		} else {
			panic(errors.Wrap(err, "could not find class for adding students"))
		}
	}
	if err := class.AddStudents(req.UserIds); err != nil {
		panic(errors.Wrap(err, "could not add students"))
	}
	return c.JSON(http.StatusOK, response.AddStudentsResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			*resource.ClassDetail `json:"class"`
		}{
			resource.GetClassDetail(&class),
		},
	})
}

func DeleteStudents(c echo.Context) error {
	req := request.DeleteStudentsRequest{}
	err, ok := utils.BindAndValidate(&req, c)
	if !ok {
		return err
	}
	class := models.Class{}
	if err := base.DB.Preload("Managers").Preload("Students").First(&class, c.Param("id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
		} else {
			panic(errors.Wrap(err, "could not find class for deleting students"))
		}
	}
	if err := class.DeleteStudents(req.UserIds); err != nil {
		panic(errors.Wrap(err, "could not delete students"))
	}
	return c.JSON(http.StatusOK, response.DeleteStudentsResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			*resource.ClassDetail `json:"class"`
		}{
			resource.GetClassDetail(&class),
		},
	})
}

func JoinClass(c echo.Context) error {
	req := request.JoinClassRequest{}
	err, ok := utils.BindAndValidate(&req, c)
	if !ok {
		return err
	}
	class := models.Class{}
	if err := base.DB.Preload("Managers").Preload("Students").First(&class, "invite_code = ?", req.InviteCode).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
		} else {
			panic(errors.Wrap(err, "could not find class for deleting students"))
		}
	}
	user := c.Get("user").(models.User)
	count := base.DB.Model(&class).Where("id = ?", user.ID).Association("Students").Count()
	if count > 0 {
		return c.JSON(http.StatusBadRequest, response.ErrorResp("ALREADY_IN_CLASS", nil))
	}
	if err := base.DB.Model(&class).Association("Students").Append(&user); err != nil {
		panic(errors.Wrap(err, "could not add student for joining class"))
	}
	return c.JSON(http.StatusOK, response.JoinClassResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			*resource.Class `json:"class"`
		}{
			resource.GetClass(&class),
		},
	})
}

func DeleteClass(c echo.Context) error {
	class := models.Class{}
	if err := base.DB.First(&class, c.Param("id")).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		panic(errors.Wrap(err, "could not find class for deleting"))
	}
	utils.PanicIfDBError(base.DB.Delete(&class), "could not delete class for deleting")
	return c.JSON(http.StatusOK, response.Response{
		Message: "SUCCESS",
		Error:   nil,
		Data:    nil,
	})
}

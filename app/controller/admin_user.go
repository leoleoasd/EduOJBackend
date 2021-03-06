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
	"strconv"
)

func AdminCreateUser(c echo.Context) error {
	req := request.AdminCreateUserRequest{}
	err, ok := utils.BindAndValidate(&req, c)
	if !ok {
		return err
	}
	hashed := utils.HashPassword(req.Password)
	count := int64(0)
	utils.PanicIfDBError(base.DB.Model(&models.User{}).Where("email = ?", req.Email).Count(&count), "could not query user count")
	if count != 0 {
		return c.JSON(http.StatusConflict, response.ErrorResp("CONFLICT_EMAIL", nil))
	}
	utils.PanicIfDBError(base.DB.Model(&models.User{}).Where("username = ?", req.Username).Count(&count), "could not query user count")
	if count != 0 {
		return c.JSON(http.StatusConflict, response.ErrorResp("CONFLICT_USERNAME", nil))
	}
	user := models.User{
		Username: req.Username,
		Nickname: req.Nickname,
		Email:    req.Email,
		Password: hashed,
	}
	utils.PanicIfDBError(base.DB.Create(&user), "could not create user")
	return c.JSON(http.StatusCreated, response.AdminCreateUserResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			*resource.UserForAdmin `json:"user"`
		}{
			resource.GetUserForAdmin(&user),
		},
	})
}

func AdminUpdateUser(c echo.Context) error {
	req := request.AdminUpdateUserRequest{}
	err, ok := utils.BindAndValidate(&req, c)
	if !ok {
		return err
	}
	user, err := utils.FindUser(c.Param("id"))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
	} else if err != nil {
		panic(err)
	}
	count := int64(0)
	utils.PanicIfDBError(base.DB.Model(&models.User{}).Where("email = ?", req.Email).Count(&count), "could not query user count")
	if count > 1 || (count == 1 && user.Email != req.Email) {
		return c.JSON(http.StatusConflict, response.ErrorResp("CONFLICT_EMAIL", nil))
	}
	utils.PanicIfDBError(base.DB.Model(&models.User{}).Where("username = ?", req.Username).Count(&count), "could not query user count")
	if count > 1 || (count == 1 && user.Username != req.Username) {
		return c.JSON(http.StatusConflict, response.ErrorResp("CONFLICT_USERNAME", nil))
	}
	user.Username = req.Username
	user.Nickname = req.Nickname
	user.Email = req.Email
	if req.Password != "" {
		hashed := utils.HashPassword(req.Password)
		user.Password = hashed
	}
	utils.PanicIfDBError(base.DB.Save(&user), "could not update user")
	return c.JSON(http.StatusOK, response.AdminUpdateUserResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			*resource.UserForAdmin `json:"user"`
		}{
			resource.GetUserForAdmin(user),
		},
	})
}

func AdminDeleteUser(c echo.Context) error {
	user, err := utils.FindUser(c.Param("id"))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
	} else if err != nil {
		panic(err)
	}
	utils.PanicIfDBError(base.DB.Delete(&user), "could not delete user")
	return c.JSON(http.StatusOK, response.Response{
		Message: "SUCCESS",
		Error:   nil,
		Data:    nil,
	})
}

func AdminGetUser(c echo.Context) error {

	user, err := utils.FindUser(c.Param("id"))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
	} else if err != nil {
		panic(err)
	}
	if !user.RoleLoaded {
		user.LoadRoles()
	}
	return c.JSON(http.StatusOK, response.AdminGetUserResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			*resource.UserForAdmin `json:"user"`
		}{
			resource.GetUserForAdmin(user),
		},
	})
}

func AdminGetUsers(c echo.Context) error {
	req := request.AdminGetUsersRequest{}
	if err, ok := utils.BindAndValidate(&req, c); !ok {
		return err
	}

	query, err := utils.Sorter(base.DB.Model(&models.User{}), req.OrderBy, "id", "username", "nickname", "email")
	if err != nil {
		if herr, ok := err.(utils.HttpError); ok {
			return herr.Response(c)
		}
		panic(err)
	}

	if req.Search != "" {
		id, _ := strconv.ParseUint(req.Search, 10, 64)
		query = query.Where("id = ? or username like ? or email like ? or nickname like ?", id, "%"+req.Search+"%", "%"+req.Search+"%", "%"+req.Search+"%")
	}
	var users []*models.User
	total, prevUrl, nextUrl, err := utils.Paginator(query, req.Limit, req.Offset, c.Request().URL, &users)
	if err != nil {
		if herr, ok := err.(utils.HttpError); ok {
			return herr.Response(c)
		}
		panic(err)
	}
	return c.JSON(http.StatusOK, response.AdminGetUsersResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			Users  []resource.User `json:"users"`
			Total  int             `json:"total"`
			Count  int             `json:"count"`
			Offset int             `json:"offset"`
			Prev   *string         `json:"prev"`
			Next   *string         `json:"next"`
		}{
			resource.GetUserSlice(users),
			total,
			len(users),
			req.Offset,
			prevUrl,
			nextUrl,
		},
	})
}

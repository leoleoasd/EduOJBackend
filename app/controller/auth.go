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

func Login(c echo.Context) error {
	req := request.LoginRequest{}
	if err, ok := utils.BindAndValidate(&req, c); !ok {
		return err
	}
	user := models.User{}
	err := base.DB.Where("email = ? or username = ?", req.UsernameOrEmail, req.UsernameOrEmail).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, response.ErrorResp("WRONG_USERNAME", nil))
		} else {
			panic(errors.Wrap(err, "could not query username or email"))
		}
	}
	if !utils.VerifyPassword(req.Password, user.Password) {
		return c.JSON(http.StatusForbidden, response.ErrorResp("WRONG_PASSWORD", nil))
	}
	token := models.Token{
		Token:      utils.RandStr(32),
		User:       user,
		RememberMe: req.RememberMe,
	}
	utils.PanicIfDBError(base.DB.Create(&token), "could not create token for users")
	if !user.RoleLoaded {
		user.LoadRoles()
	}
	return c.JSON(http.StatusOK, response.RegisterResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			User  resource.UserForAdmin `json:"user"`
			Token string                `json:"token"`
		}{
			User:  *resource.GetUserForAdmin(&user),
			Token: token.Token,
		},
	})
}

func Register(c echo.Context) error {
	req := request.RegisterRequest{}
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
	token := models.Token{
		Token: utils.RandStr(32),
		User:  user,
	}
	utils.PanicIfDBError(base.DB.Create(&token), "could not create token for user")
	return c.JSON(http.StatusCreated, response.RegisterResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			User  resource.UserForAdmin `json:"user"`
			Token string                `json:"token"`
		}{
			User:  *resource.GetUserForAdmin(&user),
			Token: token.Token,
		},
	})
}

func EmailRegistered(c echo.Context) error {
	req := request.EmailRegistered{}
	err, ok := utils.BindAndValidate(&req, c)
	if !ok {
		return err
	}
	var count int64
	utils.PanicIfDBError(base.DB.Model(&models.User{}).Where("email = ?", req.Email).Count(&count), "could not query user count")
	if count != 0 {
		return c.JSON(http.StatusConflict, response.ErrorResp("EMAIL_REGISTERED", nil))
	}
	return c.JSON(http.StatusOK, response.Response{
		Message: "SUCCESS",
		Error:   nil,
		Data:    nil,
	})
}

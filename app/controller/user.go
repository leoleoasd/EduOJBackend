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
	"gorm.io/gorm/clause"
	"net/http"
	"strconv"
)

func GetUser(c echo.Context) error {

	user, err := utils.FindUser(c.Param("id"))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
	} else if err != nil {
		panic(err)
	}
	return c.JSON(http.StatusOK, response.GetUserResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			*resource.User `json:"user"`
		}{
			resource.GetUser(user),
		},
	})
}

func GetMe(c echo.Context) error {
	user := c.Get("user").(models.User)
	if !user.RoleLoaded {
		user.LoadRoles()
	}
	return c.JSON(http.StatusOK, response.GetMeResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			*resource.UserForAdmin `json:"user"`
		}{
			resource.GetUserForAdmin(&user),
		},
	})
}

func GetUsers(c echo.Context) error {
	// TODO: refactor to use utils.Sorter and utils.Paginator.
	req := new(request.GetUsersRequest)
	if err, ok := utils.BindAndValidate(req, c); !ok {
		return err
	}
	var users []*models.User
	var total int

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

	total, prevUrl, nextUrl, err := utils.Paginator(query, req.Limit, req.Offset, c.Request().URL, &users)
	if err != nil {
		if herr, ok := err.(utils.HttpError); ok {
			return herr.Response(c)
		}
		panic(err)
	}

	return c.JSON(http.StatusOK, response.GetUsersResponse{
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

func UpdateMe(c echo.Context) error {
	user, ok := c.Get("user").(models.User)
	if !ok {
		panic("could not convert my user into type models.User")
	}
	if !user.RoleLoaded {
		user.LoadRoles()
	}
	req := request.UpdateMeRequest{}
	err, ok := utils.BindAndValidate(&req, c)
	if !ok {
		return err
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
	utils.PanicIfDBError(base.DB.Omit(clause.Associations).Save(&user), "could not update user")
	return c.JSON(http.StatusOK, response.UpdateMeResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			*resource.UserForAdmin `json:"user"`
		}{
			resource.GetUserForAdmin(&user),
		},
	})
}

func ChangePassword(c echo.Context) error {
	req := request.ChangePasswordRequest{}
	err, ok := utils.BindAndValidate(&req, c)
	if !ok {
		return err
	}
	user, ok := c.Get("user").(models.User)
	if !ok {
		panic("could not get user from context")
	}
	if !utils.VerifyPassword(req.OldPassword, user.Password) {
		return c.JSON(http.StatusForbidden, response.ErrorResp("WRONG_PASSWORD", nil))
	}
	tokenString := c.Request().Header.Get("Authorization")
	if tokenString == "" {
		panic("could not get tokenString from request header")
	}
	utils.PanicIfDBError(base.DB.Where("user_id = ? and token != ?", user.ID, tokenString).Delete(models.Token{}), "could not remove token")
	user.Password = utils.HashPassword(req.NewPassword)
	base.DB.Save(&user)
	return c.JSON(http.StatusOK, response.Response{
		Message: "SUCCESS",
		Error:   nil,
		Data:    nil,
	})
}

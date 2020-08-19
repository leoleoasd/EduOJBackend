package controller

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
	"github.com/leoleoasd/EduOJBackend/app/request"
	"github.com/leoleoasd/EduOJBackend/app/response"
	"github.com/leoleoasd/EduOJBackend/base"
	"github.com/leoleoasd/EduOJBackend/base/utils"
	"github.com/leoleoasd/EduOJBackend/database/models"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
	"strings"
)

func GetUser(c echo.Context) error {

	user, err := utils.FindUser(c.Param("id"))
	if err == gorm.ErrRecordNotFound {
		return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
	} else if err != nil {
		panic(err)
	}
	return c.JSON(http.StatusOK, response.GetUserResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			*models.User `json:"user"`
		}{
			user,
		},
	})
}

func GetUserMe(c echo.Context) error {
	var user models.User
	var ok bool
	if user, ok = c.Get("user").(models.User); !ok {
		panic("could not convert my user into type models.User")
	}
	if !user.RoleLoaded {
		user.LoadRoles()
	}
	return c.JSON(http.StatusOK, response.GetUserResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			*models.User `json:"user"`
		}{
			&user,
		},
	})
}

func GetUsers(c echo.Context) error {
	req := new(request.GetUsersRequest)
	if err, ok := utils.BindAndValidate(req, c); !ok {
		return err
	}
	var users []models.User
	var total int

	query := base.DB.Model(&models.User{})
	if req.OrderBy != "" {
		order := strings.SplitN(req.OrderBy, ".", 2)
		if len(order) != 2 {
			return c.JSON(http.StatusBadRequest, response.ErrorResp("INVALID_ORDER", nil))
		}
		if !utils.Contain(order[0], []string{"username", "id", "nickname"}) {
			return c.JSON(http.StatusBadRequest, response.ErrorResp("INVALID_ORDER", nil))
		}
		if !utils.Contain(order[1], []string{"ASC", "DESC"}) {
			return c.JSON(http.StatusBadRequest, response.ErrorResp("INVALID_ORDER", nil))
		}
		query = query.Order(strings.Join(order, " "))
	}
	if req.Search != "" {
		query = query.Where("id like ? or username like ? or email like ? or nickname like ?", "%"+req.Search+"%", "%"+req.Search+"%", "%"+req.Search+"%", "%"+req.Search+"%")
	}
	if req.Limit == 0 {
		req.Limit = 20 // Default limit
	}
	err := query.Limit(req.Limit).Offset(req.Offset).Find(&users).Error
	if err != nil {
		panic(errors.Wrap(err, "could not query users"))
	}
	err = query.Count(&total).Error
	if err != nil {
		panic(errors.Wrap(err, "could not query count of users"))
	}

	var nextUrlStr *string
	var prevUrlStr *string

	if req.Offset-req.Limit >= 0 {
		prevURL := c.Request().URL
		q, err := url.ParseQuery(prevURL.RawQuery)
		if err != nil {
			panic(errors.Wrap(err, "could not parse query for url"))
		}
		q.Set("offset", fmt.Sprint(req.Offset-req.Limit))
		q.Set("limit", fmt.Sprint(req.Limit))
		prevURL.RawQuery = q.Encode()
		temp := prevURL.String()
		prevUrlStr = &temp
	} else {
		prevUrlStr = nil
	}
	if req.Offset+len(users) < total {
		nextURL := c.Request().URL
		q, err := url.ParseQuery(nextURL.RawQuery)
		if err != nil {
			panic(errors.Wrap(err, "could not parse query for url"))
		}
		q.Set("offset", fmt.Sprint(req.Offset+req.Limit))
		q.Set("limit", fmt.Sprint(req.Limit))
		nextURL.RawQuery = q.Encode()
		temp := nextURL.String()
		nextUrlStr = &temp
	} else {
		nextUrlStr = nil
	}

	return c.JSON(http.StatusOK, response.GetUsersResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			Users  []models.User `json:"users"`
			Total  int           `json:"total"`
			Count  int           `json:"count"`
			Offset int           `json:"offset"`
			Prev   *string       `json:"prev"`
			Next   *string       `json:"next"`
		}{
			users,
			total,
			len(users),
			req.Offset,
			prevUrlStr,
			nextUrlStr,
		},
	})
}

package controller

import (
	"github.com/EduOJ/backend/app/request"
	"github.com/EduOJ/backend/app/response"
	"github.com/EduOJ/backend/base"
	"github.com/EduOJ/backend/base/utils"
	"github.com/EduOJ/backend/database/models"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
	"strings"
)

func AdminGetLogs(c echo.Context) error {
	req := request.AdminGetLogsRequest{}
	if err, ok := utils.BindAndValidate(&req, c); !ok {
		return err
	}

	query := base.DB.Model(&models.Log{}).Order("id desc")

	if len(req.Levels) != 0 {
		levelsS := strings.Split(req.Levels, ",")
		levels := make([]int, 0, len(levelsS))
		for _, ll := range levelsS {
			l, err := strconv.ParseInt(ll, 10, 32)
			if err != nil {
				return c.JSON(http.StatusBadRequest, response.ErrorResp("INVALID_LEVEL", nil))
			}
			if l < 0 || l > 4 { // TODO: replace hard-coded level
				return c.JSON(http.StatusBadRequest, response.ErrorResp("INVALID_LEVEL", nil))
			}
			levels = append(levels, int(l))
		}
		query = query.Where("level in (?)", levels)
	}

	var logs []models.Log
	total, prevUrl, nextUrl, err := utils.Paginator(query, req.Limit, req.Offset, c.Request().URL, &logs)
	if err != nil {
		if herr, ok := err.(utils.HttpError); ok {
			return herr.Response(c)
		}
		panic(err)
	}
	return c.JSON(http.StatusOK, response.AdminGetLogsResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			Logs   []models.Log `json:"logs"`
			Total  int          `json:"total"`
			Count  int          `json:"count"`
			Offset int          `json:"offset"`
			Prev   *string      `json:"prev"`
			Next   *string      `json:"next"`
		}{
			logs,
			total,
			len(logs),
			req.Offset,
			prevUrl,
			nextUrl,
		},
	})
}

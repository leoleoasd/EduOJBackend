package middleware

import (
	"github.com/EduOJ/backend/app/response"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"net/http"
)

func Judger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Request().Header.Get("Authorization") == viper.GetString("judger.token") {
			if c.Request().Header.Get("Judger-Name") == "" {
				return c.JSON(http.StatusBadRequest, response.ErrorResp("JUDGER_NAME_EXPECTED", nil))
			} else {
				return next(c)
			}
		}
		return c.JSON(http.StatusForbidden, response.ErrorResp("PERMISSION_DENIED", nil))
	}
}

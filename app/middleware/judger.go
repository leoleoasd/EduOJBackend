package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/leoleoasd/EduOJBackend/app/response"
	"github.com/spf13/viper"
	"net/http"
)

func Judger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Request().Header.Get("Authorization") == viper.GetString("judger.token") {
			return next(c)
		}
		return c.JSON(http.StatusForbidden, response.ErrorResp("PERMISSION_DENIED", nil))
	}
}

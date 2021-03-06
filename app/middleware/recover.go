package middleware

import (
	"fmt"
	"github.com/EduOJ/backend/app/response"
	"github.com/EduOJ/backend/base/log"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"net/http"
	"runtime/debug"
)

func Recover(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		defer func() {
			if xx := recover(); xx != nil {
				if err, ok := xx.(error); ok {
					log.Error(errors.Wrap(err, "controller panics"))
				} else {
					log.Error("controller panics: ", xx)
				}
				if viper.GetBool("debug") {
					stack := debug.Stack()
					err = c.JSON(http.StatusInternalServerError, response.ErrorResp("INTERNAL_ERROR", fmt.Sprintf("%+v\n%s\n", xx, stack)))
				} else {
					err = response.InternalErrorResp(c)
				}
			}
		}()
		err = next(c)
		return
	}
}

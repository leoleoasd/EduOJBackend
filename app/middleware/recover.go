package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/leoleoasd/EduOJBackend/app/response"
	"github.com/leoleoasd/EduOJBackend/base/config"
	"github.com/leoleoasd/EduOJBackend/base/log"
	"github.com/pkg/errors"
	"net/http"
	"runtime/debug"
)

func Recover(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		defer func() {
			if xx := recover(); xx != nil {
				if err, ok := xx.(error); ok {
					log.Error(errors.Wrap(err, "controller panics"))
				} else {
					log.Error("controller panics: ", xx)
				}
				if config.MustGet("debug", false).Value().(bool) {
					stack := debug.Stack()
					c.JSON(http.StatusInternalServerError, response.ErrorResp(-1, "internal error", string(stack)))
				} else {
					response.InternalErrorResp(c)
				}
			}
		}()
		return next(c)
	}
}

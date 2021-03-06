package utils

import (
	"fmt"
	"github.com/EduOJ/backend/app/response"
	"github.com/labstack/echo/v4"
)

type HttpError struct {
	Code    int
	Message string
	Err     error
}

func (e HttpError) Error() string {
	return fmt.Sprintf("[%d]%s", e.Code, e.Message)
}

func (e HttpError) Response(c echo.Context) error {
	return c.JSON(e.Code, response.ErrorResp(e.Message, e.Err))
}

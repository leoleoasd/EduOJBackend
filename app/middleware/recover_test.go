package middleware_test

import (
	"bytes"
	"github.com/EduOJ/backend/app/middleware"
	"github.com/EduOJ/backend/app/response"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRecover(t *testing.T) {
	t.Parallel()

	e := echo.New()
	e.Use(middleware.Recover)
	e.POST("/panics_with_error", func(context echo.Context) error {
		panic(errors.New("123"))
	})
	e.POST("/panics_with_other", func(context echo.Context) error {
		panic("123")
	})

	resp := response.Response{}
	req := makeReq(t, "POST", "/panics_with_error", &bytes.Buffer{})
	httpResp := makeResp(req, e)
	mustJsonDecode(httpResp, &resp)
	assert.Equal(t, response.MakeInternalErrorResp(), resp)
	req = makeReq(t, "POST", "/panics_with_other", &bytes.Buffer{})
	httpResp = makeResp(req, e)
	mustJsonDecode(httpResp, &resp)
	assert.Equal(t, response.MakeInternalErrorResp(), resp)
}

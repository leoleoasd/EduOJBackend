package middleware_test

import (
	"github.com/EduOJ/backend/app/middleware"
	"github.com/EduOJ/backend/app/response"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestJudger(t *testing.T) {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}, middleware.Judger)
	viper.Set("judger.token", "token_for_test_random_str_askudhoewiudhozSDjkfhqosuidfhasloihoase")

	t.Run("Success", func(t *testing.T) {
		resp := makeResp(makeReq(t, "GET", "/", "", headerOption{
			"Authorization": []string{"token_for_test_random_str_askudhoewiudhozSDjkfhqosuidfhasloihoase"},
			"Judger-Name":   []string{"name"},
		}), e)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		assert.Equal(t, "OK", string(body))
		assert.NoError(t, err)
	})

	t.Run("NoName", func(t *testing.T) {
		resp := makeResp(makeReq(t, "GET", "/", "", headerOption{
			"Authorization": []string{"token_for_test_random_str_askudhoewiudhozSDjkfhqosuidfhasloihoase"},
		}), e)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		jsonEQ(t, response.ErrorResp("JUDGER_NAME_EXPECTED", nil), resp)
	})

	t.Run("WrongToken", func(t *testing.T) {
		resp := makeResp(makeReq(t, "GET", "/", "", headerOption{
			"Authorization": []string{"wrong_token"},
		}), e)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
		jsonEQ(t, response.ErrorResp("PERMISSION_DENIED", nil), resp)
	})

	t.Run("MissionToken", func(t *testing.T) {
		resp := makeResp(makeReq(t, "GET", "/", ""), e)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
		jsonEQ(t, response.ErrorResp("PERMISSION_DENIED", nil), resp)
	})
}

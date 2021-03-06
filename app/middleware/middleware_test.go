package middleware_test

import (
	"bytes"
	"encoding/json"
	"github.com/EduOJ/backend/app"
	"github.com/EduOJ/backend/base"
	"github.com/EduOJ/backend/base/exit"
	"github.com/EduOJ/backend/base/log"
	"github.com/EduOJ/backend/base/validator"
	"github.com/EduOJ/backend/database"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func jsonEQ(t *testing.T, expected, actual interface{}) {
	assert.JSONEq(t, mustJsonEncode(t, expected), mustJsonEncode(t, actual))
}

func mustJsonEncode(t *testing.T, data interface{}) string {
	var err error
	if dataResp, ok := data.(*http.Response); ok {
		data, err = ioutil.ReadAll(dataResp.Body)
		assert.NoError(t, err)
	}
	if dataString, ok := data.(string); ok {
		data = []byte(dataString)
	}
	if dataBytes, ok := data.([]byte); ok {
		err := json.Unmarshal(dataBytes, &data)
		assert.NoError(t, err)
	}
	j, err := json.Marshal(data)
	if err != nil {
		t.Fatal(data, err)
	}
	return string(j)
}

func mustJsonDecode(data interface{}, out interface{}) {
	var err error
	if dataResp, ok := data.(*http.Response); ok {
		data, err = ioutil.ReadAll(dataResp.Body)
		if err != nil {
			panic(err)
		}
	}
	if dataString, ok := data.(string); ok {
		data = []byte(dataString)
	}
	if dataBytes, ok := data.([]byte); ok {
		err = json.Unmarshal(dataBytes, out)
		if err != nil {
			panic(err)
		}
	}
}

type reqOption interface {
	make(r *http.Request)
}

type headerOption map[string][]string
type queryOption map[string][]string

func (h headerOption) make(r *http.Request) {
	for k, v := range h {
		for _, s := range v {
			r.Header.Add(k, s)
		}
	}
}

func (q queryOption) make(r *http.Request) {
	for k, v := range q {
		for _, s := range v {
			r.URL.Query().Add(k, s)
		}
	}
}

func makeReq(t *testing.T, method string, path string, data interface{}, options ...reqOption) *http.Request {
	j, err := json.Marshal(data)
	assert.NoError(t, err)
	req := httptest.NewRequest(method, path, bytes.NewReader(j))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	for _, option := range options {
		option.make(req)
	}
	return req
}

func makeResp(req *http.Request, e *echo.Echo) *http.Response {
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec.Result()
}

func TestMain(m *testing.M) {
	defer database.SetupDatabaseForTest()()
	defer exit.SetupExitForTest()()
	viper.SetConfigType("yaml")
	configFile := bytes.NewBufferString(`debug: false
server:
  port: 8080
  origin:
    - http://127.0.0.1:8000
auth:
  session_timeout: 1200
  remember_me_timeout: 604800
  session_count: 10`)
	err := viper.ReadConfig(configFile)
	if err != nil {
		panic(err)
	}
	log.Disable()

	base.Echo = echo.New()
	base.Echo.Validator = validator.NewEchoValidator()
	app.Register(base.Echo)
	os.Exit(m.Run())
}

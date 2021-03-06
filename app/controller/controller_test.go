package controller_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/EduOJ/backend/app"
	"github.com/EduOJ/backend/app/response"
	"github.com/EduOJ/backend/base"
	"github.com/EduOJ/backend/base/exit"
	"github.com/EduOJ/backend/base/utils"
	"github.com/EduOJ/backend/base/validator"
	"github.com/EduOJ/backend/database"
	"github.com/EduOJ/backend/database/models"
	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
	"github.com/labstack/echo/v4"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"hash/fnv"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

var applyAdminUser headerOption
var applyNormalUser headerOption
var judgerAuthorize reqOption

func initGeneralTestingUsers() {
	adminRole := models.Role{
		Name:   "testUsersGlobalAdmin",
		Target: nil,
	}
	base.DB.Create(&adminRole)
	_ = adminRole.AddPermission("all")
	adminUser := models.User{
		Username: "test_admin_user",
		Nickname: "test_admin_nickname",
		Email:    "test_admin@mail.com",
		Password: "test_admin_password",
	}
	normalUser := models.User{
		Username: "test_normal_user",
		Nickname: "test_normal_nickname",
		Email:    "test_normal@mail.com",
		Password: "test_normal_password",
	}
	base.DB.Create(&adminUser)
	base.DB.Create(&normalUser)
	adminUser.GrantRole(adminRole.Name)
	applyAdminUser = headerOption{
		"Set-User-For-Test": {fmt.Sprintf("%d", adminUser.ID)},
	}
	applyNormalUser = headerOption{
		"Set-User-For-Test": {fmt.Sprintf("%d", normalUser.ID)},
	}
	languages := []models.Language{
		{
			Name:             "test_language",
			ExtensionAllowed: []string{"test_language"},
			RunScript: &models.Script{
				Name:     "test_language_run",
				Filename: "test_language_run",
			},
			BuildScript: &models.Script{
				Name:     "test_language_build",
				Filename: "test_language_build",
			},
		},
		{
			Name:             "golang",
			ExtensionAllowed: []string{"go"},
		},
	}
	base.DB.Save(&languages)
}

func applyUser(user models.User) headerOption {
	return headerOption{
		"Set-User-For-Test": {fmt.Sprintf("%d", user.ID)},
	}
}

func setUserForTest(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userIdString := c.Request().Header.Get("Set-User-For-Test")
		if userIdString == "" {
			return next(c)
		}
		userId, err := strconv.Atoi(userIdString)
		if err != nil {
			panic(errors.Wrap(err, "could not convert user id string to user id"))
		}
		user := models.User{}
		base.DB.First(&user, userId)
		c.Set("user", user)
		return next(c)
	}
}

func hashStringToTime(s string) time.Time {
	h := fnv.New32()
	if _, err := h.Write([]byte(s)); err != nil {
		panic(err)
	}
	return time.Unix(int64(h.Sum32()), 0).UTC()
}

type failTest struct {
	name       string
	method     string
	path       string
	req        interface{}
	reqOptions []reqOption
	statusCode int
	resp       response.Response
}

func runFailTests(t *testing.T, tests []failTest, groupName string) {
	t.Run("test"+groupName+"Fail", func(t *testing.T) {
		t.Parallel()
		for _, test := range tests {
			test := test
			t.Run("test"+groupName+test.name, func(t *testing.T) {
				t.Parallel()
				var req *http.Request
				req = makeReq(t, test.method, test.path, test.req, test.reqOptions...)
				httpResp := makeResp(req)
				resp := response.Response{}
				mustJsonDecode(httpResp, &resp)
				assert.Equal(t, test.statusCode, httpResp.StatusCode)
				assert.Equal(t, test.resp, resp)
			})
		}
	})
}

func createUserForTest(t *testing.T, name string, id int) (user models.User) {
	user = models.User{
		Username: fmt.Sprintf("test_%s_user_%d", name, id),
		Nickname: fmt.Sprintf("test_%s_user_%d_nick", name, id),
		Email:    fmt.Sprintf("test_%s_user_%d@e.e", name, id),
		Password: utils.HashPassword(fmt.Sprintf("test_%s_user_%d_pwd", name, id)),
	}
	assert.NoError(t, base.DB.Create(&user).Error)
	return
}

func jsonEQ(t *testing.T, expected, actual interface{}) {
	assert.JSONEq(t, mustJsonEncode(t, expected), mustJsonEncode(t, actual))
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

type reqContent interface {
	add(r *multipart.Writer) error
}

type fieldContent struct {
	key   string
	value string
}

type fileContent struct {
	key      string
	fileName string
	reader   io.ReadSeeker
	size     int64
}

func (c *fieldContent) add(w *multipart.Writer) (err error) {
	tw, err := w.CreateFormField(c.key)
	if err != nil {
		return
	}
	_, err = tw.Write([]byte(c.value))
	return
}

func (c *fileContent) add(w *multipart.Writer) (err error) {
	fw, err := w.CreateFormFile(c.key, c.fileName)
	if err != nil {
		return
	}
	b, err := ioutil.ReadAll(c.reader)
	_, err = io.Copy(fw, bytes.NewReader(b))
	if err != nil {
		return err
	}
	_, err = c.reader.Seek(0, io.SeekStart)
	return
}

func addFieldContentSlice(c []reqContent, fields map[string]string) []reqContent {
	fieldContents := make([]reqContent, 0, len(fields))
	for key, value := range fields {
		fieldContents = append(fieldContents, &fieldContent{
			key:   key,
			value: value,
		})
	}
	return append(c, fieldContents...)
}

func newFileContent(key, fileName, base64Data string) *fileContent {
	b, err := ioutil.ReadAll(base64.NewDecoder(base64.StdEncoding, strings.NewReader(base64Data)))
	if err != nil {
		panic(err)
	}
	return &fileContent{
		key:      key,
		fileName: fileName,
		reader:   bytes.NewReader(b),
		size:     int64(len(b)),
	}
}

func makeReq(t *testing.T, method string, path string, data interface{}, options ...reqOption) *http.Request {
	var req *http.Request
	if content, ok := data.([]reqContent); ok {
		var b bytes.Buffer
		w := multipart.NewWriter(&b)
		for _, c := range content {
			assert.NoError(t, c.add(w))
		}
		err := w.Close()
		assert.NoError(t, err)
		req = httptest.NewRequest(method, path, &b)
		req.Header.Set("Content-Type", w.FormDataContentType())
	} else {
		j, err := json.Marshal(data)
		assert.NoError(t, err)
		req = httptest.NewRequest(method, path, bytes.NewReader(j))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}
	for _, option := range options {
		option.make(req)
	}
	return req
}

func makeResp(req *http.Request) *http.Response {
	rec := httptest.NewRecorder()
	base.Echo.ServeHTTP(rec, req)
	return rec.Result()
}

func b64Encode(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}

func b64Encodef(format string, a ...interface{}) string {
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(format, a...)))
}

func getPresignedURLContent(t *testing.T, presignedUrl string) (content string) {
	resp, err := http.Get(presignedUrl)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	p, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	return string(p)
}

func getObjectContent(t *testing.T, bucketName, objectName string) (content []byte) {
	obj, err := base.Storage.GetObject(context.Background(), bucketName, objectName, minio.GetObjectOptions{})
	assert.NoError(t, err)
	content, err = ioutil.ReadAll(obj)
	assert.NoError(t, err)
	return
}

func checkObjectNonExist(t *testing.T, bucketName, objectName string) {
	obj, err := base.Storage.GetObject(context.Background(), bucketName, objectName, minio.GetObjectOptions{})
	assert.NoError(t, err)
	_, err = ioutil.ReadAll(obj)
	assert.NotNil(t, err)
	assert.Equal(t, "The specified key does not exist.", err.Error())
}

func TestMain(m *testing.M) {
	defer database.SetupDatabaseForTest()()
	defer exit.SetupExitForTest()()
	viper.SetConfigType("yaml")
	configFile := bytes.NewBufferString(`debug: true
server:
  port: 8080
  origin:
    - http://127.0.0.1:8000
judger:
  token: judger_token
`)
	err := viper.ReadConfig(configFile)
	judgerAuthorize = headerOption{
		"Authorization": []string{viper.GetString("judger.token")},
		"Judger-Name":   []string{"test_judger"},
	}
	if err != nil {
		panic(err)
	}

	base.Echo = echo.New()
	base.Echo.Validator = validator.NewEchoValidator()
	app.Register(base.Echo)
	base.Echo.Use(setUserForTest)
	initGeneralTestingUsers()
	// fake s3
	faker := gofakes3.New(s3mem.New()) // in-memory s3 server.
	ts := httptest.NewServer(faker.Server())
	defer ts.Close()
	base.Storage, err = minio.New(ts.URL[7:], &minio.Options{
		Creds:  credentials.NewStaticV4("accessKey", "secretAccessKey", ""),
		Secure: false,
	})
	if err != nil {
		panic(err)
	}
	_, err = base.Storage.ListBuckets(context.Background())
	if err != nil {
		panic(err)
	}
	if err := utils.CreateBucket("images"); err != nil {
		panic(err)
	}
	if err := utils.CreateBucket("scripts"); err != nil {
		panic(err)
	}
	if err := utils.CreateBucket("problems"); err != nil {
		panic(err)
	}
	if err := utils.CreateBucket("submissions"); err != nil {
		panic(err)
	}

	//log.Disable()

	os.Exit(m.Run())
}

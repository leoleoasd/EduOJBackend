package controller_test

import (
	"github.com/EduOJ/backend/app/request"
	"github.com/EduOJ/backend/app/response"
	"github.com/EduOJ/backend/app/response/resource"
	"github.com/EduOJ/backend/base"
	"github.com/EduOJ/backend/base/utils"
	"github.com/EduOJ/backend/database/models"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

type testClass struct {
	ID uint `gorm:"primaryKey" json:"id"`
}

func (c testClass) TypeName() string {
	return "test_class"
}

func (c testClass) GetID() uint {
	return c.ID
}

func TestLogin(t *testing.T) {
	t.Parallel()
	// strip monotonic time

	userTestingWrongPassword := models.User{
		Username: "test_login_wrong_password",
		Nickname: "test_login_wrong_password_nick",
		Email:    "test_login_wrong_password@e.e",
		Password: utils.HashPassword("test_login_password"),
	}
	assert.NoError(t, base.DB.Create(&userTestingWrongPassword).Error)

	failTests := []failTest{
		{
			name:   "WithoutParams",
			method: "POST",
			path:   base.Echo.Reverse("auth.login"),
			req: request.LoginRequest{
				UsernameOrEmail: "",
				Password:        "",
			},
			reqOptions: nil,
			statusCode: http.StatusBadRequest,
			resp: response.Response{
				Message: "VALIDATION_ERROR",
				Error: []interface{}{
					map[string]interface{}{
						"field":       "UsernameOrEmail",
						"reason":      "required",
						"translation": "用户名为必填字段",
					},
					map[string]interface{}{
						"field":       "Password",
						"reason":      "required",
						"translation": "密码为必填字段",
					},
				},
				Data: nil,
			},
		},
		{
			name:   "NotFound",
			method: "POST",
			path:   base.Echo.Reverse("auth.login"),
			req: request.LoginRequest{
				UsernameOrEmail: "test_login_1_not_found",
				Password:        "test_login_password",
			},
			reqOptions: nil,
			statusCode: http.StatusNotFound,
			resp:       response.ErrorResp("WRONG_USERNAME", nil),
		},
		{
			name:   "WrongPassword",
			method: "POST",
			path:   base.Echo.Reverse("auth.login"),
			req: request.LoginRequest{
				UsernameOrEmail: "test_login_wrong_password",
				Password:        "test_login_password_wrong",
			},
			reqOptions: nil,
			statusCode: http.StatusForbidden,
			resp:       response.ErrorResp("WRONG_PASSWORD", nil),
		},
	}
	runFailTests(t, failTests, "Login")

	classA := testClass{ID: 1}
	dummy := "test_class"
	adminRole := models.Role{
		Name:   "testLoginAdmin",
		Target: &dummy,
	}
	base.DB.Create(&adminRole)
	adminRole.AddPermission("all")

	successTests := []struct {
		name       string
		user       models.User
		req        request.LoginRequest
		roleName   *string
		roleTarget models.HasRole
	}{
		{
			name: "SuccessWithUsername",
			user: models.User{
				Username: "test_login_1",
				Nickname: "test_login_1_nick",
				Email:    "test_login_1@mail.com",
				Password: utils.HashPassword("test_login_1_pwd"),
			},
			req: request.LoginRequest{
				UsernameOrEmail: "test_login_1",
				Password:        "test_login_1_pwd",
				RememberMe:      false,
			},
			roleName:   nil,
			roleTarget: nil,
		},
		{
			name: "SuccessWithEmail",
			user: models.User{
				Username: "test_login_2",
				Nickname: "test_login_2_nick",
				Email:    "test_login_2@mail.com",
				Password: utils.HashPassword("test_login_2_pwd"),
			},
			req: request.LoginRequest{
				UsernameOrEmail: "test_login_2@mail.com",
				Password:        "test_login_2_pwd",
				RememberMe:      false,
			},
			roleName:   nil,
			roleTarget: nil,
		},
		{
			name: "SuccessWithRememberMe",
			user: models.User{
				Username: "test_login_3",
				Nickname: "test_login_3_nick",
				Email:    "test_login_3@mail.com",
				Password: utils.HashPassword("test_login_3_pwd"),
			},
			req: request.LoginRequest{
				UsernameOrEmail: "test_login_3",
				Password:        "test_login_3_pwd",
				RememberMe:      true,
			},
			roleName:   nil,
			roleTarget: nil,
		},
		{
			name: "SuccessWithRole",
			user: models.User{
				Username: "test_login_4",
				Nickname: "test_login_4_nick",
				Email:    "test_login_4@mail.com",
				Password: utils.HashPassword("test_login_4_pwd"),
			},
			req: request.LoginRequest{
				UsernameOrEmail: "test_login_4",
				Password:        "test_login_4_pwd",
				RememberMe:      false,
			},
			roleName:   &adminRole.Name,
			roleTarget: classA,
		},
	}

	t.Run("testLoginSuccess", func(t *testing.T) {
		t.Parallel()
		for _, test := range successTests {
			test := test
			t.Run("testLogin"+test.name, func(t *testing.T) {
				t.Parallel()
				assert.NoError(t, base.DB.Create(&test.user).Error)
				if test.roleName != nil {
					test.user.GrantRole(*test.roleName, test.roleTarget)
				}
				test.user.LoadRoles()
				httpResp := makeResp(makeReq(t, "POST", base.Echo.Reverse("auth.login"), test.req))
				resp := response.LoginResponse{}
				mustJsonDecode(httpResp, &resp)
				assert.Equal(t, "SUCCESS", resp.Message)
				assert.Nil(t, resp.Error)
				jsonEQ(t, resource.GetUserForAdmin(&test.user), resp.Data.User)
				token := models.Token{}
				assert.NoError(t, base.DB.Preload("User").Where("token = ?", resp.Data.Token).First(&token).Error)
				token.User.LoadRoles()
				jsonEQ(t, test.user, token.User)
				assert.Equal(t, test.req.RememberMe, token.RememberMe)
			})
		}
	})
}

func TestRegister(t *testing.T) {

	dummyUserForConflict := models.User{
		Username: "test_register_conflict",
		Nickname: "test_register_conflict_nick",
		Email:    "test_register_conflict@mail.com",
		Password: utils.HashPassword("test_register_conflict_pwd"),
	}
	assert.NoError(t, base.DB.Create(&dummyUserForConflict).Error)

	failTests := []failTest{
		{
			name:   "WithoutParams",
			method: "POST",
			path:   base.Echo.Reverse("auth.register"),
			req: request.RegisterRequest{
				Username: "",
				Nickname: "",
				Email:    "",
				Password: "",
			},
			reqOptions: nil,
			statusCode: http.StatusBadRequest,
			resp: response.Response{
				Message: "VALIDATION_ERROR",
				Error: []interface{}{
					map[string]interface{}{
						"field":       "Username",
						"reason":      "required",
						"translation": "用户名为必填字段",
					},
					map[string]interface{}{
						"field":       "Nickname",
						"reason":      "required",
						"translation": "昵称为必填字段",
					},
					map[string]interface{}{
						"field":       "Email",
						"reason":      "required",
						"translation": "邮箱为必填字段",
					},
					map[string]interface{}{
						"field":       "Password",
						"reason":      "required",
						"translation": "密码为必填字段",
					},
				},
				Data: nil,
			},
		},
		{
			name:   "ConflictUsername",
			method: "POST",
			path:   base.Echo.Reverse("auth.register"),
			req: request.RegisterRequest{
				Username: "test_register_conflict",
				Nickname: "test_register_1_nick",
				Email:    "test_register_1@mail.com",
				Password: "test_register_1_pwd",
			},
			reqOptions: nil,
			statusCode: http.StatusConflict,
			resp:       response.ErrorResp("CONFLICT_USERNAME", nil),
		},
		{
			name:   "ConflictEmail",
			method: "POST",
			path:   base.Echo.Reverse("auth.register"),
			req: request.RegisterRequest{
				Username: "test_register_1",
				Nickname: "test_register_1_nick",
				Email:    "test_register_conflict@mail.com",
				Password: "test_register_1_pwd",
			},
			reqOptions: nil,
			statusCode: http.StatusConflict,
			resp:       response.ErrorResp("CONFLICT_EMAIL", nil),
		},
	}

	runFailTests(t, failTests, "Register")

	t.Run("testRegisterSuccess", func(t *testing.T) {
		t.Parallel()
		reqUser := models.User{
			Username: "test_register_3",
			Nickname: "test_register_3_nick",
			Email:    "test_register_3@mail.com",
			Password: "test_register_3_pwd",
		}
		httpResp := makeResp(makeReq(t, "POST", base.Echo.Reverse("auth.register"), request.RegisterRequest{
			Username: reqUser.Username,
			Nickname: reqUser.Nickname,
			Email:    reqUser.Email,
			Password: reqUser.Password,
		}))
		assert.Equal(t, http.StatusCreated, httpResp.StatusCode)
		resp := response.RegisterResponse{}
		mustJsonDecode(httpResp, &resp)
		assert.Equal(t, "SUCCESS", resp.Message)
		assert.Nil(t, resp.Error)
		// req == resp
		assert.Equal(t, reqUser.Username, resp.Data.User.Username)
		assert.Equal(t, reqUser.Nickname, resp.Data.User.Nickname)
		assert.Equal(t, reqUser.Email, resp.Data.User.Email)
		databaseUser := models.User{}
		assert.NoError(t, base.DB.Where("email = ?", "test_register_3@mail.com").First(&databaseUser).Error)
		// resp == database
		jsonEQ(t, resp.Data.User, resource.GetUserForAdmin(&databaseUser))
		token := models.Token{}
		assert.NoError(t, base.DB.Where("token = ?", resp.Data.Token).Last(&token).Error)
		// token == database
		assert.Equal(t, databaseUser.ID, token.UserID)
	})
}

func TestEmailRegistered(t *testing.T) {
	t.Parallel()
	user := models.User{
		Username: "test_email_registered_username",
		Nickname: "test_email_registered_nickname",
		Email:    "test_email_registered@e.com",
		Password: "test_email_registered_passwd",
	}
	base.DB.Create(&user)
	t.Run("testEmailRegisteredConflict", func(t *testing.T) {
		t.Parallel()
		httpResp := makeResp(makeReq(t, "GET", base.Echo.Reverse("auth.emailRegistered"), request.EmailRegistered{
			Email: "test_email_registered@e.com",
		}))
		resp := response.Response{}
		mustJsonDecode(httpResp, &resp)
		assert.Equal(t, http.StatusConflict, httpResp.StatusCode)
		assert.Equal(t, response.Response{
			Message: "EMAIL_REGISTERED",
			Error:   nil,
			Data:    nil,
		}, resp)
	})
	t.Run("testEmailRegisteredSuccess", func(t *testing.T) {
		t.Parallel()
		httpResp := makeResp(makeReq(t, "GET", base.Echo.Reverse("auth.emailRegistered"), request.EmailRegistered{
			Email: "test_email_registered_0@e.com",
		}))
		resp := response.Response{}
		mustJsonDecode(httpResp, &resp)
		assert.Equal(t, http.StatusOK, httpResp.StatusCode)
		assert.Equal(t, response.Response{
			Message: "SUCCESS",
			Error:   nil,
			Data:    nil,
		}, resp)
	})
}

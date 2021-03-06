package controller

import (
	"bytes"
	"encoding/json"
	"github.com/EduOJ/backend/app/response"
	"github.com/EduOJ/backend/app/response/resource"
	"github.com/EduOJ/backend/base"
	"github.com/EduOJ/backend/base/log"
	"github.com/EduOJ/backend/base/utils"
	"github.com/EduOJ/backend/database/models"
	"github.com/duo-labs/webauthn/protocol"
	"github.com/duo-labs/webauthn/webauthn"
	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var cac = cache.New(5*time.Minute, 10*time.Minute)

func BeginRegistration(c echo.Context) error {
	user := c.Get("user").(models.User)
	log.Debug(base.WebAuthn.Config)
	options, sessionData, err := base.WebAuthn.BeginRegistration(&user)
	if err != nil {
		panic(errors.Wrap(err, "could not begin register webauthn"))
	}
	cac.Set("register"+user.Username, sessionData, cache.DefaultExpiration)
	return c.JSON(http.StatusOK, response.Response{
		Message: "SUCCESS",
		Error:   nil,
		Data:    options,
	})
}

func FinishRegistration(c echo.Context) error {
	user := c.Get("user").(models.User)
	sessionData, found := cac.Get("register" + user.Username)
	if !found {
		panic(errors.New("not registered"))
	}
	b, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		panic(errors.Wrap(err, "could not read body"))
	}
	parsedResponse, err := protocol.ParseCredentialCreationResponseBody(bytes.NewReader(b))
	if err != nil {
		panic(errors.Wrap(err, "could not parse credential from body"))
	}
	credential, err := base.WebAuthn.CreateCredential(&user, *sessionData.(*webauthn.SessionData), parsedResponse)
	if err != nil {
		panic(errors.Wrap(err, "could not create credential"))
	}
	webauthnCredential := models.WebauthnCredential{}
	m, err := json.Marshal(credential)
	if err != nil {
		panic(errors.Wrap(err, "could not marshal credential to json"))
	}
	webauthnCredential.Content = string(m)
	if err := base.DB.Model(&user).Association("Credentials").Append(&webauthnCredential); err != nil {
		panic(errors.Wrap(err, "could not save credentials"))
	}
	return c.JSON(http.StatusOK, response.Response{
		Message: "SUCCESS",
		Error:   nil,
		Data:    nil,
	})
}

func BeginLogin(c echo.Context) error {
	var id string
	err := echo.QueryParamsBinder(c).String("username", &id).BindError()
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResp("WRONG_USERNAME", nil))
	}
	user, err := utils.FindUser(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
		}
		panic(errors.Wrap(err, "could not find user"))
	}
	options, sessionData, err := base.WebAuthn.BeginLogin(user)
	if err != nil {
		var perr *protocol.Error
		if errors.As(err, &perr) {
			return c.JSON(http.StatusBadRequest, response.ErrorResp(strings.ToUpper(perr.Type), perr))
		}
		panic(errors.Wrap(err, "could not begin login webauthn"))
	}
	cac.Set("login"+user.Username, sessionData, cache.DefaultExpiration)
	return c.JSON(http.StatusOK, response.Response{
		Message: "SUCCESS",
		Error:   nil,
		Data:    options,
	})
}

func FinishLogin(c echo.Context) error {
	var id string
	err := echo.QueryParamsBinder(c).String("username", &id).BindError()
	if err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResp("WRONG_USERNAME", nil))
	}
	var user models.User
	err = base.DB.Where("email = ? or username = ?", id, id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, response.ErrorResp("NOT_FOUND", nil))
		}
		panic(errors.Wrap(err, "could not find user"))
	}
	sessionData, found := cac.Get("login" + user.Username)
	if !found {
		panic(errors.New("not registered"))
	}
	b, err := ioutil.ReadAll(c.Request().Body)
	parsedResponse, err := protocol.ParseCredentialRequestResponseBody(bytes.NewReader(b))
	if err != nil {
		panic(errors.Wrap(err, "could not parse response"))
	}
	_, err = base.WebAuthn.ValidateLogin(&user, *sessionData.(*webauthn.SessionData), parsedResponse)
	if err != nil {
		panic(errors.Wrap(err, "could not validate login webauthn"))
	}
	token := models.Token{
		Token:      utils.RandStr(32),
		User:       user,
		RememberMe: false,
	}
	utils.PanicIfDBError(base.DB.Create(&token), "could not create token for users")
	if !user.RoleLoaded {
		user.LoadRoles()
	}
	return c.JSON(http.StatusOK, response.RegisterResponse{
		Message: "SUCCESS",
		Error:   nil,
		Data: struct {
			User  resource.UserForAdmin `json:"user"`
			Token string                `json:"token"`
		}{
			User:  *resource.GetUserForAdmin(&user),
			Token: token.Token,
		},
	})
}

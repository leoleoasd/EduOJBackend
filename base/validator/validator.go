package validator

import (
	"github.com/EduOJ/backend/base/log"
	zhTranslations "github.com/EduOJ/backend/base/validator/translations/zh"
	zhLocal "github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"regexp"
)

type echoValidator struct {
	V *validator.Validate
}

func (cv *echoValidator) Validate(i interface{}) error {
	return cv.V.Struct(i)
}

var Trans ut.Translator

func init() {
	zh := zhLocal.New()
	uni := ut.New(zh, zh)
	var found bool
	Trans, found = uni.GetTranslator("zh")
	if !found {
		log.Fatal("could not found zh translator")
		panic("could not found zh translator")
	}
}

func New() *validator.Validate {
	v := validator.New()
	// add custom translation here
	if err := zhTranslations.RegisterDefaultTranslations(v, Trans); err != nil {
		log.Fatal(errors.Wrap(err, "could not register default translations"))
		panic(errors.Wrap(err, "could not register default translations"))
	}
	err := v.RegisterValidation("username", validateUsername)
	if err != nil {
		log.Fatal(errors.Wrap(err, "could not register validation"))
		panic(errors.Wrap(err, "could not register validation"))
	}
	return v
}

func NewEchoValidator() *echoValidator {
	return &echoValidator{
		V: New(),
	}
}

var usernameRegex = regexp.MustCompile("^[a-zA-Z0-9_]+$")

func validateUsername(fl validator.FieldLevel) bool {
	return usernameRegex.MatchString(fl.Field().String())
}

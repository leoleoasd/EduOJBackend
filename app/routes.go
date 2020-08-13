package app

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/leoleoasd/EduOJBackend/app/controller"
	adminController "github.com/leoleoasd/EduOJBackend/app/controller/admin"
	"github.com/leoleoasd/EduOJBackend/app/middleware"
	"github.com/leoleoasd/EduOJBackend/base/config"
	"github.com/leoleoasd/EduOJBackend/base/log"
	"github.com/pkg/errors"
	"net/http"
	"regexp"
)

var usernameRegex = regexp.MustCompile("^[a-zA-Z0-9_]+$")

func validateUsername(fl validator.FieldLevel) bool {
	return usernameRegex.MatchString(fl.Field().String())
}

func Register(e *echo.Echo) {
	v := validator.New()
	err := v.RegisterValidation("username", validateUsername)
	if err != nil {
		log.Fatal(errors.Wrap(err, "could not register validation"))
		panic(err)
	}
	e.Validator = &Validator{
		v: v,
	}

	e.Use(middleware.Recover)
	var origins []string
	if n, err := config.Get("server.origin"); err == nil {
		for _, v := range n.(*config.SliceNode).S {
			if vv, ok := v.Value().(string); ok {
				origins = append(origins, vv)
			} else {
				log.Fatal("Illegal origin name" + v.String())
				panic("Illegal origin name" + v.String())
			}
		}
	} else {
		log.Fatal("Illegal origin config", err)
		panic("Illegal origin config")
	}
	e.Use(echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
		AllowOrigins: origins,
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	api := e.Group("/api", middleware.Authentication)

	auth := api.Group("/auth", middleware.Auth)
	auth.POST("/login", controller.Login).Name = "auth.login"
	auth.POST("/register", controller.Register).Name = "auth.register"

	loginCheck := api.Group("/", middleware.LoginCheck)
	_ = loginCheck

	admin := api.Group("/admin")
	admin.POST("/user", adminController.PostUser)
	admin.PUT("/user/:id", adminController.PutUser)
	admin.DELETE("/user/:id", adminController.DeleteUser)
	admin.GET("/user/:id", adminController.GetUser)
	admin.GET("/users", adminController.GetUsers)

	api.GET("/user/:id", controller.GetUser)
	api.GET("/users", controller.GetUsers)

	// TODO: routes.
}

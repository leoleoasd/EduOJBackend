package app

import (
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/leoleoasd/EduOJBackend/app/controller"
	"github.com/leoleoasd/EduOJBackend/app/middleware"
	"github.com/leoleoasd/EduOJBackend/base/config"
	"github.com/leoleoasd/EduOJBackend/base/log"
	"net/http"
)

func Register(e *echo.Echo) {
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
	auth.GET("/email_registered", controller.EmailRegistered)

	admin := api.Group("/admin", middleware.Logged)
	// TODO: add HasPermission
	admin.POST("/user", controller.AdminCreateUser)
	admin.PUT("/user/:id", controller.AdminUpdateUser)
	admin.DELETE("/user/:id", controller.AdminDeleteUser)
	admin.GET("/user/:id", controller.AdminGetUser)
	admin.GET("/users", controller.AdminGetUsers)

	api.GET("/user/me", controller.GetUserMe, middleware.Logged)
	api.PUT("/user/me", controller.UpdateUserMe, middleware.Logged)
	api.GET("/user/:id", controller.GetUser)
	api.GET("/users", controller.GetUsers)

	api.POST("/user/change_password", controller.ChangePassword, middleware.Logged)

	// TODO: routes.
}

package main

import (
	"context"
	"fmt"
	"github.com/EduOJ/backend/app"
	"github.com/EduOJ/backend/base"
	"github.com/EduOJ/backend/base/event"
	"github.com/EduOJ/backend/base/exit"
	"github.com/EduOJ/backend/base/log"
	"github.com/EduOJ/backend/base/utils"
	"github.com/EduOJ/backend/base/validator"
	"github.com/EduOJ/backend/database"
	runEvent "github.com/EduOJ/backend/event/run"
	submissionEvent "github.com/EduOJ/backend/event/submission"
	"github.com/duo-labs/webauthn/webauthn"
	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
)

func readConfig() {
	log.Debug("Reading config.")
	viper.SetConfigName("config")         // name of config file (without extension)
	viper.AddConfigPath("/etc/eduoj/")    // path to look for the config file in
	viper.AddConfigPath("$HOME/.appname") // call multiple times to add many search paths
	viper.AddConfigPath(".")              // optionally look for config in the working directory
	err := viper.ReadInConfig()           // Find and read the config file
	if err != nil {                       // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}

func initLog() {
	log.Debug("Initializing log.")
	err := log.InitFromConfig()
	if err != nil {
		log.Fatal(errors.Wrap(err, "could not init log"))
		os.Exit(-1)
	}
	log.Debug("Logging initialized.")
}

func initEvent() {
	log.Debug("Initializing Event System.")
	event.RegisterListener("run", runEvent.NotifyGetSubmissionPoll)
	event.RegisterListener("submission", submissionEvent.UpdateGrade)
}

func startEcho() {
	log.Debug("Starting echo server.")
	port := viper.GetInt("server.port")
	base.Echo = echo.New()
	base.Echo.Logger = &log.EchoLogger{}
	base.Echo.HideBanner = true
	base.Echo.HidePort = true
	base.Echo.Use(middleware.Recover())
	base.Echo.Server.Addr = fmt.Sprintf(":%d", port)
	base.Echo.Validator = validator.NewEchoValidator()
	app.Register(base.Echo)
	exit.QuitWG.Add(1)
	go func() {
		err := base.Echo.StartServer(base.Echo.Server)
		if err != nil {
			log.Fatal(errors.Wrap(err, "server closed"))
		}
	}()
	log.Fatal("Server started at ", base.Echo.Server.Addr)

	// When server closes, closes web server.
	go func() {
		<-exit.BaseContext.Done()
		err := base.Echo.Shutdown(context.Background())
		if err != nil {
			if err.Error() == "context canceled" {
				log.Fatal("Force quitting.")
			} else {
				log.Fatal(err)
			}
		}
		exit.QuitWG.Done()
	}()
}

func initRedis() {
	log.Debug("Starting redis client.")
	base.Redis = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprint(viper.Get("redis.host"), ":", viper.GetInt("redis.port")),
		Username: viper.GetString("redis.username"),
		Password: viper.GetString("redis.password"),
	})
	// test connection.
	_, err := base.Redis.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal(errors.Wrap(err, "could not init redis"))
		os.Exit(-1)
	}
	log.Debug("Redis client started.")
}

func initGorm(toMigrate ...bool) {
	log.Debug("Starting database client.")
	var err error
	switch viper.GetString("database.dialect") {
	case "mysql":
		base.DB, err = gorm.Open(mysql.Open(viper.GetString("database.uri")), &gorm.Config{
			Logger: log.GormLogger{},
		})
	case "postgres":
		base.DB, err = gorm.Open(postgres.Open(viper.GetString("database.uri")), &gorm.Config{
			Logger: log.GormLogger{},
		})
	case "sqlite":
		base.DB, err = gorm.Open(sqlite.Open(viper.GetString("database.uri")), &gorm.Config{
			Logger:                                   log.GormLogger{},
			DisableForeignKeyConstraintWhenMigrating: true,
		})
	default:
		log.Fatal("unsupported database dialect")
		os.Exit(-1)
	}
	if err != nil {
		log.Fatal(errors.Wrap(err, "could not init database"))
		os.Exit(-1)
	}
	if len(toMigrate) == 0 || toMigrate[0] {
		database.Migrate()
	}
	log.Debug("Database client started.")

	// Cause we need to wait until all logs are wrote to the db
	// So we dont close db connection here.
}

func initStorage() {
	log.Debug("Starting storage client.")
	var err error
	base.Storage, err = minio.New(viper.GetString("storage.endpoint"), &minio.Options{
		Region: viper.GetString("storage.region"),
		Creds:  credentials.NewStaticV4(viper.GetString("storage.access_key_id"), viper.GetString("storage.access_key_secret"), ""),
		Secure: viper.GetBool("storage.false"),
	})
	if err != nil {
		log.Fatal(errors.Wrap(err, "could not connect to minio server."))
		panic(err)
	}
	_, err = base.Storage.ListBuckets(context.Background())
	if err != nil {
		log.Fatal(errors.Wrap(err, "could not connect to minio server."))
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
	log.Debug("Storage client initialized.")
}

func initWebAuthn() {
	var err error
	base.WebAuthn, err = webauthn.New(&webauthn.Config{
		RPDisplayName: viper.GetString("webauthn.display_name"),
		RPID:          viper.GetString("webauthn.domain"),
		RPOrigin:      viper.GetString("webauthn.origin"),
		RPIcon:        viper.GetString("webauthn.icon"),
	})
	if err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}
}

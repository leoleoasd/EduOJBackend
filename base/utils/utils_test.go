package utils

import (
	"bytes"
	"context"
	"github.com/EduOJ/backend/base"
	"github.com/EduOJ/backend/base/log"
	"github.com/EduOJ/backend/database"
	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"net/http/httptest"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	defer database.SetupDatabaseForTest()()
	if err := base.DB.Migrator().AutoMigrate(&TestObject{}); err != nil {
		panic(errors.Wrap(err, "could not create table for test object"))
	}

	configFile := bytes.NewBufferString(`debug: true
server:
  port: 8080
  origin:
    - http://127.0.0.1:8000
`)
	viper.SetConfigType("yaml")
	if err := viper.ReadConfig(configFile); err != nil {
		panic(err)
	}

	// fake s3
	faker := gofakes3.New(s3mem.New()) // in-memory s3 server.
	ts := httptest.NewServer(faker.Server())
	defer ts.Close()
	var err error
	base.Storage, err = minio.New(ts.URL[7:], &minio.Options{
		Creds:  credentials.NewStaticV4("accessKey", "secretAccessKey", ""),
		Secure: false,
	})
	if err != nil {
		panic(err)
	}
	err = base.Storage.MakeBucket(context.Background(), "test-bucket", minio.MakeBucketOptions{
		Region: viper.GetString("storage.region"),
	})
	if err != nil {
		panic(err)
	}
	log.Disable()

	os.Exit(m.Run())
}

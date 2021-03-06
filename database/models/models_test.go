package models

import (
	"github.com/EduOJ/backend/base/exit"
	"github.com/EduOJ/backend/database"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	defer database.SetupDatabaseForTest()()
	defer exit.SetupExitForTest()()
	os.Exit(m.Run())
}

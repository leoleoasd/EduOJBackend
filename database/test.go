package database

import (
	"fmt"
	"github.com/EduOJ/backend/base"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SetupDatabaseForTest setups a in-memory db for testing purpose.
// Shouldn't be called out of test.
func SetupDatabaseForTest() func() {
	oldDB := base.DB
	x, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Silent),
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		fmt.Print(err)
		panic(err)
	}
	base.DB = x
	sqlDB, err := base.DB.DB()
	if err != nil {
		fmt.Print(err)
		panic(err)
	}
	sqlDB.SetMaxOpenConns(1)
	Migrate()
	return func() {
		base.DB = oldDB
	}
}

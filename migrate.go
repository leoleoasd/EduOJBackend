package main

import (
	"fmt"
	"github.com/EduOJ/backend/base/log"
	"github.com/EduOJ/backend/database"
)

func doMigrate() {

	if len(args) == 1 {
		readConfig()
		initGorm()
		log.Fatal("Migrate succeed!")
	} else {
		switch args[1] {
		case "help":
			fmt.Println(`Usage:
./backend migrate <command>
command: The command to run. default: migrate
Commands:
migrate: run migrations.
rollback: rollback all migrations.
rollback_last: rollback last migration.`)
		case "migrate":
			readConfig()
			initGorm()
			log.Fatal("Migrate succeed!")
		case "rollback":
			readConfig()
			initGorm(false)
			m := database.GetMigration()
			err := m.RollbackTo("start")
			if err != nil {
				log.Error(err)
			} else {
				log.Fatal("Migrate succeed!")
			}
		case "rollback_last":
			readConfig()
			initGorm(false)
			m := database.GetMigration()
			err := m.RollbackLast()
			if err != nil {
				log.Error(err)
			} else {
				log.Fatal("Migrate succeed!")
			}
		}
	}
}

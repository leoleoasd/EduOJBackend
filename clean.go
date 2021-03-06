package main

import (
	"github.com/EduOJ/backend/base/exit"
	"github.com/EduOJ/backend/base/log"
	"github.com/EduOJ/backend/base/utils"
	"os"
)

func clean() {
	readConfig()
	initGorm()
	initLog()
	err := utils.CleanUpExpiredTokens()
	if err != nil {
		log.Error(err)
		os.Exit(-1)
	}
	exit.Close()
	exit.QuitWG.Wait()
	log.Fatal("Clean succeed!")
}

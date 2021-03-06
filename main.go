package main

import (
	"github.com/EduOJ/backend/base/log"
	"os"
)

func main() {
	parse()
	if len(args) < 1 {
		log.Fatal("Please specific a command to run.")
		// TODO: output usage
		os.Exit(-1)
	}
	switch args[0] {
	case "test-config":
		testConfig()
	case "serve", "server", "http", "run":
		serve()
	case "migrate", "migration":
		doMigrate()
	case "clean", "clean-up", "clean-db":
		clean()
	case "permission", "perm":
		permission()
	}
}

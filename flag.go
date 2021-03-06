package main

import (
	"github.com/EduOJ/backend/base/log"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	"os"
)

type _opt struct {
	Config string `short:"c" long:"verbose" description:"Config file name" default:"config.yml"`
}

var parser *flags.Parser
var opt _opt
var open = os.Open
var args []string

func init() {
	parser = flags.NewNamedParser("eduOJ server", flags.HelpFlag|flags.PassDoubleDash)
	_, _ = parser.AddGroup("Application", "Application Options", &opt)
}

func parse() {
	var err error
	// TODO: remove useless logs for parser debugging.
	log.Debug("Parsing command-line arguments.")
	args, err = parser.Parse()
	if err != nil {
		log.Fatal(errors.Wrap(err, "could not parse argument "))
		os.Exit(-1)
	}
	log.Debug(args, err)
	log.Debug(opt)
}

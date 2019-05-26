package main

import (
	"fmt"
	"log"
	"os"

	"github.com/docopt/docopt-go"

	"github.com/c4milo/slackbot"
)

// Version is defined in compilation time.
var (
	Version string
)

const usage = `
SlackBot
Primitive configuration management tool

Usage:
  slackbot run <slackbook_file>
  slackbot -h | --help
  slackbot -v | --version

Commands:
  run                   Applies the state declared in the given slackbook YAML file

Options:
  -V --verbose          Turns on verbose output
  -h --help             Displays this help
  -v --version          Displays version string
`

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func main() {
	args, err := docopt.Parse(usage, nil, false, "", false, false)
	if err != nil {
		fmt.Printf(usage)
		os.Exit(1)
	}

	if args["--version"].(bool) {
		fmt.Println(Version)
		return
	}

	if args["--help"].(bool) {
		fmt.Println(usage)
		return
	}

	slackbookPath := args["<slackbook_file>"].(string)
	if slackbookPath == "" {
		fmt.Printf(" ! invalid slackbook_path argument \n")
		os.Exit(1)
	}

	slackbook := new(slackbot.SlackBook)
	if err := slackbook.Decode(slackbookPath); err != nil {
		fmt.Printf(" ! failed parsing slackbook file: %+v\n", err)
		os.Exit(1)
	}

	// Uncomment to debug config AST
	// repr.Println(slackbook)

	if err := slackbook.Run(); err != nil {
		fmt.Printf(" ! failed applying slackbook: %+v\n", err)
		os.Exit(1)
	}

	fmt.Println("Done 🎉")
}

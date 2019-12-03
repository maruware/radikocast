package main

import (
	"log"
	"os"

	"github.com/maruware/radikocast"
	"github.com/mitchellh/cli"
)

func main() {
	c := cli.NewCLI("radikocast", "1.0.0")
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"rec":          radikocast.RecCommandFactory,
		"rec_schedule": radikocast.RecScheduleCommandFactory,
		"rss":          radikocast.RssCommandFactory,
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)
}

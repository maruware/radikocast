package radikocast

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/mitchellh/cli"
)

type recScheduleCommand struct {
	ui cli.Ui
}

func (c *recScheduleCommand) Run(args []string) int {
	var stationID, day, at, areaID, bucket string

	f := flag.NewFlagSet("rec_schedule", flag.ContinueOnError)
	f.StringVar(&stationID, "id", "", "id")
	f.StringVar(&day, "day", "", "day")
	f.StringVar(&at, "at", "", "at")
	f.StringVar(&areaID, "area", "", "area")
	f.StringVar(&areaID, "a", "", "area")
	f.StringVar(&bucket, "bucket", "", "bucket")

	f.Usage = func() { c.ui.Error(c.Help()) }

	if err := f.Parse(args); err != nil {
		return 1
	}

	if stationID == "" {
		c.ui.Error("StationID is empty.")
		return 1
	}
	if day == "" {
		c.ui.Error("day is empty")
		return 1
	}
	if at == "" {
		c.ui.Error("at is empty")
		return 1
	}

	start, err := findLastProgram(day, at, time.Now())
	if err != nil {
		c.ui.Error(fmt.Sprintf("Bad day or at: %v", err))
		return 1
	}

	c.ui.Output("Now downloading.. ")
	code, err := RecProgram(stationID, start, areaID, bucket)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Failed to rec: %s", err))
		return 1
	}
	c.ui.Output(fmt.Sprintf("Completed!\n%s", *code))
	return 0
}

func (c *recScheduleCommand) Synopsis() string {
	return "Record a radiko program"
}

func (c *recScheduleCommand) Help() string {
	return strings.TrimSpace(`
Usage: radikocast rec [options]
  Record a radiko program.
Options:
  -id=name                 Station id
  -day=day_expression      Day expression (ex. monday)
  -area,a=name             Area id
  -bucket=bucketname	   S3 bucket name
`)
}

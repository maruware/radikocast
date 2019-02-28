package radikocast

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mitchellh/cli"
)

type recCommand struct {
	ui cli.Ui
}

func (c *recCommand) Run(args []string) int {
	var stationID, start, areaID string
	var configPath string

	f := flag.NewFlagSet("rec", flag.ContinueOnError)
	f.StringVar(&stationID, "id", "", "id")
	f.StringVar(&start, "start", "", "start")
	f.StringVar(&start, "s", "", "start")
	f.StringVar(&areaID, "area", "", "area")
	f.StringVar(&areaID, "a", "", "area")
	f.StringVar(&configPath, "config", defaultConfigPath, "config")
	f.StringVar(&configPath, "c", defaultConfigPath, "config")
	f.Usage = func() { c.ui.Error(c.Help()) }
	if err := f.Parse(args); err != nil {
		return 1
	}
	config, err := LoadConfig(configPath)
	if err != nil {
		c.ui.Error(fmt.Sprintf(
			"Failed to load config"))
		return 1
	}

	if stationID == "" {
		c.ui.Error("StationID is empty.")
		return 1
	}
	c.ui.Output("Now downloading.. ")
	code, err := RecProgram(stationID, start, areaID, config.Workspace.OutputDirAbs())
	if err != nil {
		c.ui.Error(fmt.Sprintf("Failed to rec: %s", err))
		return 1
	}
	c.ui.Output(fmt.Sprintf("Completed!\n%s", *code))
	return 0
}

func (c *recCommand) Synopsis() string {
	return "Record a radiko program"
}

func (c *recCommand) Help() string {
	return strings.TrimSpace(`
Usage: radikocast rec [options]
  Record a radiko program.
Options:
  -id=name                 Station id
  -start,s=201610101000    Start time
  -area,a=name             Area id
  -config,c=filepath	   Config file path
`)
}

func fileSize(path string) int64 {
	fi, err := os.Stat(path)
	if err != nil {
		log.Fatal(err)
	}

	return fi.Size()
}
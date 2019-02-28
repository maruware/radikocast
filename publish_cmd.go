package radikocast

import (
	"flag"
	"fmt"
	"strings"

	"github.com/mitchellh/cli"
)

type publishCommand struct {
	ui cli.Ui
}

func (c *publishCommand) Run(args []string) int {
	var configPath string

	f := flag.NewFlagSet("publish", flag.ContinueOnError)
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
	// TODO: support other publish type
	err = SyncDirToS3(config.Workspace.OutputDirAbs(), config.Publish.Bucket)
	if err != nil {
		c.ui.Error(fmt.Sprintf(
			"Failed to publish s3"))
		return 1
	}

	return 0
}

func (c *publishCommand) Synopsis() string {
	return "Publish podcast"
}

func (c *publishCommand) Help() string {
	return strings.TrimSpace(`
Usage: radikocast publish
  Publish podcast
Options:
  -config,c=filepath                 Config file path (default: config.yml)
`)
}

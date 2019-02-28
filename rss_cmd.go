package radikocast

import (
	"flag"
	"fmt"
	"strings"

	"github.com/mitchellh/cli"
)

type rssCommand struct {
	ui cli.Ui
}

func (c *rssCommand) Run(args []string) int {
	var configPath string
	f := flag.NewFlagSet("rss", flag.ContinueOnError)
	f.StringVar(&configPath, "config", defaultConfigPath, "config")
	f.StringVar(&configPath, "c", defaultConfigPath, "config")
	f.Usage = func() { c.ui.Error(c.Help()) }
	if err := f.Parse(args); err != nil {
		return 1
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		c.ui.Error(fmt.Sprintf(
			"Failed to load config %s", configPath))
		return 1
	}
	rss := GenerateRss(config.Podcast, config.Workspace.OutputDir)
	err = WriteRss(rss, config.Workspace.FeedPath())
	if err != nil {
		c.ui.Error(fmt.Sprintf("Failed to write rss %s", err))
		return 1
	}

	c.ui.Output(fmt.Sprintf("Updated %s", config.Workspace.FeedPath()))

	return 0
}

func (c *rssCommand) Synopsis() string {
	return "Generate podcast RSS"
}

func (c *rssCommand) Help() string {
	return strings.TrimSpace(`
Usage: radikocast rss [options]
  Generate podcast RSS
Options:
  -config,c=filepath	   Config file path
`)
}

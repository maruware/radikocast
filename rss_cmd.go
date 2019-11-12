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
	var title, host, image, bucket, feed string
	f := flag.NewFlagSet("rss", flag.ContinueOnError)
	f.StringVar(&title, "title", "", "title")
	f.StringVar(&host, "host", "", "host")
	f.StringVar(&image, "image", "", "image")
	f.StringVar(&bucket, "bucket", "", "bucket")
	f.StringVar(&feed, "feed", "", "feed")

	f.Usage = func() { c.ui.Error(c.Help()) }
	if err := f.Parse(args); err != nil {
		return 1
	}

	rss, err := GenerateRss(title, host, image, bucket)
	if err != nil {
		return 1
	}
	err = PutRss(rss, bucket, feed)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Failed to write rss %s", err))
		return 1
	}

	c.ui.Output(fmt.Sprintf("Updated %s", feed))

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
  -config,c=filepath	   Config file path (default: config.yml)
`)
}

package radikocast

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bamzi/jobrunner"
	"github.com/gin-gonic/gin"
	"github.com/mitchellh/cli"
)

type scheduleCommand struct {
	ui cli.Ui
}

func (c *scheduleCommand) Run(args []string) int {
	var configPath string
	f := flag.NewFlagSet("schedule", flag.ContinueOnError)
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

	jobrunner.Start()
	for _, schedule := range config.Schedules {
		jobrunner.Schedule(schedule.Cron(), RecRssPublish{
			Schedule: schedule,
			Config:   config,
		})
	}
	gin.SetMode(gin.ReleaseMode)
	routes := gin.Default()
	routes.GET("/jobrunner/json", JobJson)
	routes.Run(":8080")
	return 0
}

type RecRssPublish struct {
	Schedule *ConfigSchedule
	Config   *Config
}

func (job RecRssPublish) Run() {
	now := time.Now()
	station := job.Schedule.Station
	area := job.Schedule.Area
	start := job.Schedule.StartCode(&now)

	output := job.Config.Workspace.OutputDirAbs()

	code, err := RecProgram(station, start, area, output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to rec")
		return
	}
	fmt.Printf("Rec %s\n", *code)

	rss := GenerateRss(job.Config.Podcast, output)
	err = WriteRss(rss, job.Config.Workspace.FeedPath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write rss")
		return
	}

	// TODO: support other publish type
	err = SyncDirToS3(output, job.Config.Publish.Bucket)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to publish s3")
		return
	}
	fmt.Println("Rss published")
}

func JobJson(c *gin.Context) {
	// returns a map[string]interface{} that can be marshalled as JSON
	c.JSON(200, jobrunner.StatusJson())
}

func JobHtml(c *gin.Context) {
	// Returns the template data pre-parsed
	c.HTML(200, "", jobrunner.StatusPage())
}

func (c *scheduleCommand) Synopsis() string {
	return "Schedule programs"
}

func (c *scheduleCommand) Help() string {
	return strings.TrimSpace(`
Usage: radikocast schedule [options]
  Schedule programs
Options:
  -config,c=filepath	   Config file path (default: config.yml)
`)
}

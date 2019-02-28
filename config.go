package radikocast

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Podcast   *ConfigPodcast    `yaml:"podcast"`
	Schedules []*ConfigSchedule `yaml:"schedules"`
	Publish   *ConfigPublish    `yaml:"publish"`
	Workspace *ConfigWorkspace  `yaml:"workspace"`
}

type ConfigWorkspace struct {
	OutputDir string `yaml:"output_dir"`
}

func (c *ConfigWorkspace) OutputDirAbs() string {
	if !filepath.IsAbs(c.OutputDir) {
		wd, _ := os.Getwd()
		return filepath.Join(wd, c.OutputDir)
	}
	return c.OutputDir
}
func (c *ConfigWorkspace) FeedPath() string {
	return filepath.Join(c.OutputDir, "feed.xml")
}

type ConfigPodcast struct {
	Title string `yaml:"title"`
	Host  string `yaml:"host"`
	Image string `yaml:"image"`
}
type ConfigSchedule struct {
	Day     string           `yaml:"day"`
	At      ConfigScheduleAt `yaml:"at"`
	Station string           `yaml:"station"`
	Area    string           `yaml:"area"`
}

type ConfigPublish struct {
	PublishType string `yaml:"type"`
	Bucket      string `yaml:"bucket"`
}

type ConfigScheduleAt string

func LoadConfig(configPath string) (*Config, error) {
	buf, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(buf, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

var daysOfWeek = map[string]time.Weekday{}
var pattern *regexp.Regexp

func init() {
	for d := time.Sunday; d <= time.Saturday; d++ {
		daysOfWeek[strings.ToLower(d.String())] = d
	}
	pattern, _ = regexp.Compile(`^(\d+):(\d+)-(\d+):(\d+)`)
}

type TimeSpan struct {
	StartH int
	StartM int
	EndH   int
	EndM   int
}

func (at *ConfigScheduleAt) ParseToTimeSpan() (*TimeSpan, error) {
	matches := pattern.FindStringSubmatch(string(*at))
	nums := make([]int, 4)
	for i, m := range matches[1:] {
		n, err := strconv.Atoi(m)
		if err != nil {
			return nil, err
		}
		nums[i] = n
	}
	span := TimeSpan{
		StartH: nums[0],
		StartM: nums[1],
		EndH:   nums[2],
		EndM:   nums[3],
	}
	return &span, nil
}

func (schedule *ConfigSchedule) Cron() string {
	var dow string
	switch schedule.Day {
	case "everyday":
		dow = "*"
	case "weekday":
		dow = "1-5"
	default:
		w := daysOfWeek[schedule.Day]
		dow = fmt.Sprintf("%d", w)
	}

	span, _ := schedule.At.ParseToTimeSpan()

	return fmt.Sprintf("%d %d * * %s", span.EndM+5, span.EndH, dow)
}

func (schedule *ConfigSchedule) StartCode(baseTime *time.Time) string {
	span, _ := schedule.At.ParseToTimeSpan()
	t := time.Date(baseTime.Year(), baseTime.Month(), baseTime.Day(), span.StartH, span.StartM, 0, 0, time.UTC)
	return t.Format("20060102150405")
}

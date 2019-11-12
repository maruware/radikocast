package radikocast

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Title  string `yaml:"title"`
	Host   string `yaml:"host"`
	Image  string `yaml:"image"`
	Bucket string `yaml:"bucket"`
}

func (c *Config) FeedPath() string {
	return "feed.xml"
}

type ConfigSchedule struct {
	Day     string           `yaml:"day"`
	At      ConfigScheduleAt `yaml:"at"`
	Station string           `yaml:"station"`
	Area    string           `yaml:"area"`
}

type ConfigScheduleAt string

func NewConfig(title string, host string, image string, bucket string) *Config {
	return &Config{
		Title:  title,
		Host:   host,
		Image:  image,
		Bucket: bucket,
	}
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
	span, _ := schedule.At.ParseToTimeSpan()

	var dow string
	switch schedule.Day {
	case "everyday":
		dow = "*"
	case "weekday":
		dow = "1-5"
	default:
		w := int(daysOfWeek[schedule.Day])
		w = (w + (span.EndH / 24)) % 7
		dow = strconv.Itoa(w)
	}

	m := span.EndM + 5 // wait few minutes after end
	h := span.EndH % 24

	return fmt.Sprintf("%d %d * * %s", m, h, dow)
}

// StartCode is assumed executing after ending program.
func (schedule *ConfigSchedule) StartCode(baseTime *time.Time) string {
	span, _ := schedule.At.ParseToTimeSpan()
	t := time.Date(baseTime.Year(), baseTime.Month(), baseTime.Day(), span.StartH%24, span.StartM, 0, 0, time.UTC)
	return t.Format("20060102150405")
}

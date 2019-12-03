package radikocast

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/now"
)

var pattern *regexp.Regexp
var daysOfWeek = map[string]time.Weekday{}

func init() {
	for d := time.Sunday; d <= time.Saturday; d++ {
		daysOfWeek[strings.ToLower(d.String())] = d
	}
	pattern, _ = regexp.Compile(`^(\d+):(\d+)$`)
}

func CalcDate(day string, baseTime time.Time) (time.Time, error) {
	oneDay := time.Hour * 24
	sunday := now.With(baseTime).BeginningOfWeek()

	w := baseTime.Weekday()

	switch day {
	case "everyday":
		return baseTime, nil
	case "weekday":
		if w == time.Saturday || w == time.Sunday {
			return now.With(baseTime).Monday().Add(oneDay * 4), nil
		}
		return baseTime, nil
	default:
		if d, ok := daysOfWeek[day]; ok {
			date := sunday.Add(oneDay * time.Duration(d))
			if date.After(baseTime) {
				date = date.Add(oneDay * time.Duration(-7))
			}
			return date, nil
		}
	}
	return time.Time{}, fmt.Errorf("Bad day format: %s", day)
}

// day: ex. thursday, everyday, weekday
// at: ex. 13:42-14:38
func findLastProgram(day string, at string, baseTime time.Time) (string, error) {
	matches := pattern.FindStringSubmatch(at)
	if len(matches) != 3 {
		return "", fmt.Errorf("Bad at format: %v", at)
	}
	nums := make([]int, 2)
	for i, m := range matches[1:] {
		n, err := strconv.Atoi(m)
		if err != nil {
			return "", err
		}
		nums[i] = n
	}
	h := nums[0] % 24
	m := nums[1]

	date, err := CalcDate(day, baseTime)
	if err != nil {
		return "", err
	}

	t := time.Date(date.Year(), date.Month(), date.Day()+nums[0]/24, h, m, 0, 0, time.UTC)
	return t.Format("20060102150405"), nil
}

package radikocast

import (
	"testing"
	"time"
)

func TestCron(t *testing.T) {
	var schedule ConfigSchedule
	schedule = ConfigSchedule{Day: "weekday", At: "01:00-02:00", Station: "LFR"}
	expect := "5 2 * * 1-5"
	if actual := schedule.Cron(); actual != expect {
		t.Errorf("cron is expected %s, but %s", expect, actual)
	}
	schedule = ConfigSchedule{Day: "everyday", At: "02:00-03:00", Station: "LFR"}
	expect = "5 3 * * *"
	if actual := schedule.Cron(); actual != expect {
		t.Errorf("cron is expected %s, but %s", expect, actual)
	}
	schedule = ConfigSchedule{Day: "wednesday", At: "03:00-04:00", Station: "LFR"}
	expect = "5 4 * * 3"
	if actual := schedule.Cron(); actual != expect {
		t.Errorf("cron is expected %s, but %s", expect, actual)
	}
}

func TestStartCode(t *testing.T) {
	var schedule ConfigSchedule
	baseTime := time.Date(2010, 2, 3, 2, 5, 0, 0, time.UTC)
	schedule = ConfigSchedule{Day: "weekday", At: "01:15-02:00", Station: "LFR"}
	expect := "20100203011500"
	if actual := schedule.StartCode(&baseTime); actual != expect {
		t.Errorf("start code is expected %s, but %s", expect, actual)
	}
}

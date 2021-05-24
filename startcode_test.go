package radikocast

import (
	"testing"
	"time"
)

func TestStartCode(t *testing.T) {
	baseTime := time.Date(2010, 2, 3, 2, 5, 0, 0, time.UTC) // wednesday

	day := "weekday"
	at := "01:15"
	expect := "20100203011500"

	s, err := findLastProgram(day, at, baseTime)
	if err != nil {
		t.Errorf("findLastPrograrm err: %v", err)

	}
	if s != expect {
		t.Errorf("start code is expected %s, but %s", expect, s)
	}

	baseTime = time.Date(2010, 2, 7, 3, 5, 0, 0, time.UTC) //sunday
	day = "saturday"
	at = "25:00"
	expect = "20100207010000"

	s, err = findLastProgram(day, at, baseTime)
	if err != nil {
		t.Errorf("findLastPrograrm err: %v", err)

	}
	if s != expect {
		t.Errorf("start code is expected %s, but %s", expect, s)
	}
}

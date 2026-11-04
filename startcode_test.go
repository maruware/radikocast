package radikocast

import (
	"strings"
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

func TestRadikoProgramURL(t *testing.T) {
	start := time.Date(2025, 2, 3, 1, 15, 0, 0, location)
	want := "https://radiko.jp/#!/ts/TBS/20250203011500"
	if got := radikoProgramURL("TBS", start); got != want {
		t.Fatalf("radikoProgramURL(%q, %s) = %s, want %s", "TBS", start.Format(time.RFC3339), got, want)
	}
}

func TestUseGoRadikoDirectAlwaysEnabled(t *testing.T) {
	if !useGoRadikoDirect() {
		t.Fatal("useGoRadikoDirect() = false; want true")
	}
}

func TestProgramInfoFromScheduleXML(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<radiko>
  <stations>
    <station id="FMT">
      <progs>
        <prog id="14122033" ft="20261104050000" to="20261104060000">
          <title>Memories＆Discoveries</title>
          <url>https://jfn-pods.com/program/27337</url>
          <desc></desc>
        </prog>
      </progs>
    </station>
  </stations>
</radiko>`

	got, err := decodeProgramInfoXML([]byte(xmlData), "FMT", time.Date(2026, 11, 4, 5, 0, 0, 0, location))
	if err != nil {
		t.Fatalf("decodeProgramInfoXML() error = %v", err)
	}
	if got.Title != "Memories＆Discoveries" {
		t.Fatalf("decodeProgramInfoXML().Title = %q, want %q", got.Title, "Memories＆Discoveries")
	}
	if got.URL != "https://jfn-pods.com/program/27337" {
		t.Fatalf("decodeProgramInfoXML().URL = %q, want %q", got.URL, "https://jfn-pods.com/program/27337")
	}
	wantEnd, err := time.ParseInLocation(datetimeLayout, "20261104060000", location)
	if err != nil {
		t.Fatalf("time.ParseInLocation() error = %v", err)
	}
	if end := got.EndTime(); !end.Equal(wantEnd) {
		t.Fatalf("decodeProgramInfoXML().EndTime() = %s, want %s", end.Format(datetimeLayout), wantEnd.Format(datetimeLayout))
	}
}

func TestProgramTimeFreePlaylistUsesEndTimeWindow(t *testing.T) {
	start := time.Date(2026, 11, 4, 23, 0, 0, 0, location)
	end := start.Add(30 * time.Minute)
	got, err := directPlaylistURL("https://dr-wowza.radiko-cf.com/tf/playlist.m3u8", "TBS", "dummy", start, end, true)
	if err != nil {
		t.Fatalf("directPlaylistURL() error = %v", err)
	}
	if !strings.Contains(got, "l=300") {
		t.Fatalf("timefree playlist URL = %q, want l=300", got)
	}
	if !strings.Contains(got, "end_at=20261104233000") || !strings.Contains(got, "to=20261104233000") {
		t.Fatalf("timefree playlist URL = %q, want program end_at/to set to 20261104233000", got)
	}
}

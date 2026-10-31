package radikocast

import (
	"os"
	"path/filepath"
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

func TestRadikoURLForYtDlp(t *testing.T) {
	start := time.Date(2025, 2, 3, 1, 15, 0, 0, location)
	want := "https://radiko.jp/#!/ts/TBS/20250203011500"
	if got := radikoURLForYtDlp("TBS", start); got != want {
		t.Fatalf("radikoURLForYtDlp(%q, %s) = %s, want %s", "TBS", start.Format(time.RFC3339), got, want)
	}
}

func TestDownloadRadikoAudioRetriesOnTransientFailure(t *testing.T) {
	oldPath := os.Getenv("PATH")
	binDir := t.TempDir()
	counterPath := filepath.Join(binDir, "attempts.txt")
	script := "#!/bin/bash\ncount=0\nif [ -f \"$YTDLP_FAKE_COUNTER\" ]; then count=$(cat \"$YTDLP_FAKE_COUNTER\") ; fi\ncount=$((count + 1))\nprintf '%s' \"$count\" > \"$YTDLP_FAKE_COUNTER\"\nout=\"\"\nwhile [ \"$#\" -gt 0 ]; do\n  if [ \"$1\" = \"--output\" ]; then\n    shift\n    out=\"$1\"\n  fi\n  shift\ndone\nif [ \"$count\" -eq 1 ]; then\n  echo 'ERROR: url must be a string' >&2\n  exit 1\nfi\nactual=\"${out//%(ext)s/m4a}\"\ntouch \"$actual\"\nexit 0\n"
	if err := os.WriteFile(filepath.Join(binDir, "yt-dlp"), []byte(script), 0o755); err != nil {
		t.Fatalf("write fake yt-dlp: %v", err)
	}
	if err := os.Setenv("PATH", binDir+":"+oldPath); err != nil {
		t.Fatalf("set PATH: %v", err)
	}
	if err := os.Setenv("YTDLP_FAKE_COUNTER", counterPath); err != nil {
		t.Fatalf("set YTDLP_FAKE_COUNTER: %v", err)
	}
	defer func() {
		_ = os.Setenv("PATH", oldPath)
		_ = os.Unsetenv("YTDLP_FAKE_COUNTER")
	}()

	dir := t.TempDir()
	start := time.Date(2025, 2, 3, 1, 15, 0, 0, location)
	got, err := downloadRadikoAudio(t.Context(), "TBS", start, dir)
	if err != nil {
		t.Fatalf("downloadRadikoAudio() returned error after transient failure: %v", err)
	}
	if filepath.Ext(got) != ".m4a" {
		t.Fatalf("downloadRadikoAudio() returned %q, want .m4a output", got)
	}
}

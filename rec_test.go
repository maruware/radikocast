package radikocast

import (
	"os"
	"testing"
	"time"
)

func TestRecProgram(t *testing.T) {
	stationID := "TBS"

	start, err := findLastProgram("sunday", "24:00", time.Now())
	if err != nil {
		t.Fatalf("failed to findLastProgram: %v", err)
	}
	r, err := recProgram(stationID, start, "")
	if err != nil {
		t.Fatalf("failed to recProgram: %v", err)
	}
	defer r.Dispose()

	if _, err := os.Stat(r.audioPath); err != nil {
		t.Errorf("not found audio: %s", r.audioPath)
	}
	if _, err := os.Stat(r.metadataPath); err != nil {
		t.Errorf("not found metadata: %s", r.metadataPath)
	}
}

package radikocast

import (
	"context"
	"fmt"
	"time"
)

// ProgramInfo holds the basic metadata associated with a Radiko program.
type ProgramInfo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"webpage_url"`
	FT          string `json:"ft"`
	Timestamp   int64  `json:"timestamp"`
}

func radikoURLForYtDlp(stationID string, start time.Time) string {
	return fmt.Sprintf("https://radiko.jp/#!/ts/%s/%s", stationID, start.In(location).Format(datetimeLayout))
}

func getProgramInfo(ctx context.Context, stationID string, startTime time.Time) (*ProgramInfo, error) {
	url := radikoURLForYtDlp(stationID, startTime)
	return &ProgramInfo{
		Title:       stationID,
		Description: "",
		URL:         url,
		FT:          startTime.In(location).Format(datetimeLayout),
	}, nil
}

func downloadRadikoAudio(ctx context.Context, stationID string, startTime time.Time, dir string) (string, error) {
	return downloadRadikoAudioGo(ctx, stationID, startTime, dir)
}

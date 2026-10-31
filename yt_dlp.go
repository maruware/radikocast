package radikocast

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ProgramInfo holds the basic metadata that yt-dlp extracts for a Radiko program.
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

func findYtDlpBinary() (string, error) {
	path, err := exec.LookPath("yt-dlp")
	if err != nil {
		return "", fmt.Errorf("yt-dlp is required for Radiko downloads and must be on PATH: %w", err)
	}
	return path, nil
}

func getProgramInfo(ctx context.Context, stationID string, startTime time.Time) (*ProgramInfo, error) {
	path, err := findYtDlpBinary()
	if err != nil {
		return nil, err
	}

	url := radikoURLForYtDlp(stationID, startTime)
	cmd := exec.CommandContext(ctx, path, "--no-playlist", "--skip-download", "--dump-json", url)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var info ProgramInfo
	if err := json.Unmarshal(out, &info); err != nil {
		return nil, err
	}
	if info.Title == "" {
		info.Title = stationID
	}
	if info.URL == "" {
		info.URL = url
	}
	if info.FT == "" {
		info.FT = startTime.In(location).Format(datetimeLayout)
	}
	return &info, nil
}

func cleanupRadikoOutputFiles(dir string) {
	matches, err := filepath.Glob(filepath.Join(dir, "radiko-output.*"))
	if err != nil {
		return
	}
	for _, match := range matches {
		_ = os.Remove(match)
	}
}

func downloadRadikoAudio(ctx context.Context, stationID string, startTime time.Time, dir string) (string, error) {
	path, err := findYtDlpBinary()
	if err != nil {
		return "", err
	}

	outputTemplate := filepath.Join(dir, "radiko-output.%(ext)s")
	url := radikoURLForYtDlp(stationID, startTime)
	const maxAttempts = 3

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		cleanupRadikoOutputFiles(dir)

		cmd := exec.CommandContext(
			ctx,
			path,
			"--no-playlist",
			"--no-warnings",
			"--format",
			"bestaudio/best",
			"--output",
			outputTemplate,
			url,
		)
		output, err := cmd.CombinedOutput()
		if err == nil {
			matches, err := filepath.Glob(filepath.Join(dir, "radiko-output.*"))
			if err != nil {
				return "", err
			}
			for _, match := range matches {
				if strings.HasSuffix(match, ".part") || strings.HasSuffix(match, ".ytdl") {
					continue
				}
				return match, nil
			}
			lastErr = fmt.Errorf("yt-dlp did not produce a Radiko audio file")
			continue
		}

		combined := strings.TrimSpace(string(output))
		lastErr = fmt.Errorf("yt-dlp failed: %w: %s", err, combined)
		if !strings.Contains(combined, "url must be a string") {
			return "", lastErr
		}
		if attempt < maxAttempts {
			continue
		}
		return "", lastErr
	}

	return "", lastErr
}

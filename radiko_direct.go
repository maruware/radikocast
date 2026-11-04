package radikocast

import (
	"context"
	"crypto/rand"
	_ "embed"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

//go:embed radiko_aSmartPhone7a.bin
var radikoKey []byte

type radikoAuth struct {
	Token  string
	User   string
	AreaID string
}

type stationRegionDocument struct {
	Stations []struct {
		Station []struct {
			ID     string `xml:"id"`
			AreaID string `xml:"area_id"`
		} `xml:"station"`
	} `xml:"stations"`
}

type stationStreamDocument struct {
	URLs []struct {
		TimeFree bool   `xml:"timefree,attr"`
		AreaFree bool   `xml:"areafree,attr"`
		Playlist string `xml:"playlist_create_url"`
	} `xml:"url"`
}

const (
	timefreeMaxChunkSeconds = 300
)

func useGoRadikoDirect() bool {
	return true
}

func randHex(n int) string {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return strings.TrimLeft(base64.RawURLEncoding.EncodeToString(buf), "=")[:n]
}

func stationRegion(stationID string) (string, error) {
	resp, err := http.Get("https://radiko.jp/v3/station/region/full.xml")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("region xml request failed: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var doc stationRegionDocument
	if err := xml.Unmarshal(body, &doc); err != nil {
		return "", err
	}
	for _, stations := range doc.Stations {
		for _, station := range stations.Station {
			if strings.EqualFold(station.ID, stationID) {
				if station.AreaID == "" {
					return "JP13", nil
				}
				return station.AreaID, nil
			}
		}
	}
	return "", fmt.Errorf("station %s not found in Radiko region metadata", stationID)
}

func fetchRadikoAuth(stationID string) (*radikoAuth, error) {
	region, err := stationRegion(stationID)
	if err != nil {
		return nil, err
	}

	authReq, err := http.NewRequest("GET", "https://radiko.jp/v2/api/auth1", nil)
	if err != nil {
		return nil, err
	}
	user := randHex(32)
	authReq.Header.Set("X-Radiko-App", "aSmartPhone7a")
	authReq.Header.Set("X-Radiko-App-Version", "7.5.0")
	authReq.Header.Set("X-Radiko-Device", "24.SM-G935F")
	authReq.Header.Set("X-Radiko-User", user)
	authReq.Header.Set("User-Agent", "Dalvik/2.1.0 (Linux; U; Android 7.0.0;SM-G935F/NBD92Q)")

	authResp, err := http.DefaultClient.Do(authReq)
	if err != nil {
		return nil, err
	}
	defer authResp.Body.Close()
	if authResp.StatusCode < 200 || authResp.StatusCode >= 300 {
		return nil, fmt.Errorf("radiko auth1 failed: %s", authResp.Status)
	}

	token := authResp.Header.Get("X-Radiko-AuthToken")
	keyLen, err := strconv.Atoi(authResp.Header.Get("X-Radiko-KeyLength"))
	if err != nil {
		return nil, fmt.Errorf("parse X-Radiko-KeyLength: %w", err)
	}
	keyOff, err := strconv.Atoi(authResp.Header.Get("X-Radiko-KeyOffset"))
	if err != nil {
		return nil, fmt.Errorf("parse X-Radiko-KeyOffset: %w", err)
	}
	partialKey := base64.StdEncoding.EncodeToString(radikoKey[keyOff : keyOff+keyLen])

	coords := map[string]string{
		"JP1":  "43.064615,141.346807,gps",
		"JP2":  "40.824308,140.739998,gps",
		"JP3":  "39.703619,141.152684,gps",
		"JP4":  "38.268837,140.8721,gps",
		"JP5":  "39.718614,140.102364,gps",
		"JP6":  "38.240436,140.363633,gps",
		"JP7":  "37.750299,140.467551,gps",
		"JP8":  "36.341811,140.446793,gps",
		"JP9":  "36.565725,139.883565,gps",
		"JP10": "36.390668,139.060406,gps",
		"JP11": "35.856999,139.648849,gps",
		"JP12": "35.605057,140.123306,gps",
		"JP13": "35.689488,139.691706,gps",
		"JP14": "35.447507,139.642345,gps",
		"JP15": "37.902552,139.023095,gps",
	}
	location := coords[region]
	if location == "" {
		location = "35.689488,139.691706,gps"
	}

	auth2Req, err := http.NewRequest("GET", "https://radiko.jp/v2/api/auth2", nil)
	if err != nil {
		return nil, err
	}
	auth2Req.Header.Set("X-Radiko-App", "aSmartPhone7a")
	auth2Req.Header.Set("X-Radiko-App-Version", "7.5.0")
	auth2Req.Header.Set("X-Radiko-Device", "24.SM-G935F")
	auth2Req.Header.Set("X-Radiko-User", user)
	auth2Req.Header.Set("X-Radiko-AuthToken", token)
	auth2Req.Header.Set("X-Radiko-Location", location)
	auth2Req.Header.Set("X-Radiko-Connection", "wifi")
	auth2Req.Header.Set("X-Radiko-Partialkey", partialKey)
	auth2Req.Header.Set("User-Agent", "Dalvik/2.1.0 (Linux; U; Android 7.0.0;SM-G935F/NBD92Q)")

	auth2Resp, err := http.DefaultClient.Do(auth2Req)
	if err != nil {
		return nil, err
	}
	defer auth2Resp.Body.Close()
	if auth2Resp.StatusCode < 200 || auth2Resp.StatusCode >= 300 {
		return nil, fmt.Errorf("radiko auth2 failed: %s", auth2Resp.Status)
	}
	body, err := io.ReadAll(auth2Resp.Body)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(strings.TrimSpace(string(body)), ",")
	if len(parts) < 2 {
		return nil, fmt.Errorf("unexpected auth2 response: %q", string(body))
	}
	return &radikoAuth{Token: token, User: user, AreaID: parts[0]}, nil
}

func preferredPlaylistScore(rawURL string) int {
	if strings.Contains(rawURL, "dr-wowza.radiko-cf.com") || strings.Contains(rawURL, "rd-wowza-radiko.radiko-cf.com") {
		return 0
	}
	if strings.Contains(rawURL, "radiko.jp") || strings.Contains(rawURL, "f-radiko.smartstream.ne.jp") {
		return 1
	}
	if strings.Contains(rawURL, "tf-f-rpaa-radiko.smartstream.ne.jp") || strings.Contains(rawURL, "tf-c-rpaa-radiko.smartstream.ne.jp") {
		return 2
	}
	return 3
}

func directPlaylistBaseURL(stationID string, timefree bool) (string, error) {
	streamURL := fmt.Sprintf("https://radiko.jp/v3/station/stream/pc_html5/%s.xml", stationID)
	resp, err := http.Get(streamURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("stream XML request failed: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var doc stationStreamDocument
	if err := xml.Unmarshal(body, &doc); err != nil {
		return "", err
	}

	candidates := make([]string, 0, len(doc.URLs))
	for _, u := range doc.URLs {
		if u.TimeFree == timefree && !u.AreaFree && u.Playlist != "" {
			candidates = append(candidates, u.Playlist)
		}
	}
	if len(candidates) == 0 {
		return "", fmt.Errorf("no matching stream playlist was found for %s", stationID)
	}

	best := candidates[0]
	bestScore := preferredPlaylistScore(best)
	for _, candidate := range candidates[1:] {
		score := preferredPlaylistScore(candidate)
		if score < bestScore {
			best = candidate
			bestScore = score
		}
	}
	return best, nil
}

func directPlaylistURL(basePlaylistURL, stationID, lsid string, startTime, endTime time.Time, timefree bool) (string, error) {
	if endTime.IsZero() || endTime.Before(startTime) {
		endTime = startTime.Add(15 * time.Minute)
	}

	parsed, err := url.Parse(basePlaylistURL)
	if err != nil {
		return "", err
	}
	q := parsed.Query()
	q.Set("station_id", stationID)
	q.Set("lsid", lsid)
	q.Set("type", "b")
	if timefree {
		l := int(endTime.Sub(startTime).Seconds())
		if l <= 0 {
			l = timefreeMaxChunkSeconds
		}
		if l > timefreeMaxChunkSeconds {
			l = timefreeMaxChunkSeconds
		}
		q.Set("l", strconv.Itoa(l))
		q.Set("start_at", startTime.In(location).Format(datetimeLayout))
		q.Set("ft", startTime.In(location).Format(datetimeLayout))
		q.Set("end_at", endTime.In(location).Format(datetimeLayout))
		q.Set("to", endTime.In(location).Format(datetimeLayout))
	} else {
		q.Set("l", "15")
	}
	parsed.RawQuery = q.Encode()
	return parsed.String(), nil
}

func fetchWithRadikoHeaders(urlString string, auth *radikoAuth, extra map[string]string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, urlString, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Radiko-AreaId", auth.AreaID)
	req.Header.Set("X-Radiko-AuthToken", auth.Token)
	req.Header.Set("X-Radiko-User", auth.User)
	req.Header.Set("X-Radiko-Connection", "wifi")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	for k, v := range extra {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %s for %s", resp.Status, urlString)
	}
	return io.ReadAll(resp.Body)
}

func resolveMediaPlaylistURL(playlistURL string, auth *radikoAuth) (string, error) {
	body, err := fetchWithRadikoHeaders(playlistURL, auth, nil)
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
			return line, nil
		}
	}
	return "", fmt.Errorf("no media playlist URL in %s", playlistURL)
}

func mediaSegmentURLs(mediaPlaylistURL string, auth *radikoAuth) ([]string, error) {
	body, err := fetchWithRadikoHeaders(mediaPlaylistURL, auth, nil)
	if err != nil {
		return nil, err
	}
	segments := make([]string, 0, 256)
	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
			segments = append(segments, line)
		}
	}
	if len(segments) == 0 {
		return nil, fmt.Errorf("no media segments in %s", mediaPlaylistURL)
	}
	return segments, nil
}

func downloadRadikoAudioGo(ctx context.Context, stationID string, startTime, endTime time.Time, dir string) (string, error) {
	if endTime.IsZero() || endTime.Before(startTime) {
		endTime = startTime.Add(15 * time.Minute)
	}

	auth, err := fetchRadikoAuth(stationID)
	if err != nil {
		return "", fmt.Errorf("radiko auth failed: %w", err)
	}

	basePlaylistURL, err := directPlaylistBaseURL(stationID, true)
	if err != nil {
		return "", err
	}

	segmentURLSet := map[string]struct{}{}
	segments := make([]string, 0, 1024)
	for chunkStart := startTime; chunkStart.Before(endTime); {
		chunkEnd := chunkStart.Add(timefreeMaxChunkSeconds * time.Second)
		if chunkEnd.After(endTime) {
			chunkEnd = endTime
		}

		masterPlaylistURL, err := directPlaylistURL(basePlaylistURL, stationID, auth.User, chunkStart, chunkEnd, true)
		if err != nil {
			return "", err
		}
		mediaPlaylistURL, err := resolveMediaPlaylistURL(masterPlaylistURL, auth)
		if err != nil {
			return "", err
		}
		chunkSegments, err := mediaSegmentURLs(mediaPlaylistURL, auth)
		if err != nil {
			return "", err
		}
		for _, segment := range chunkSegments {
			if _, exists := segmentURLSet[segment]; exists {
				continue
			}
			segmentURLSet[segment] = struct{}{}
			segments = append(segments, segment)
		}
		chunkStart = chunkEnd
	}

	segmentFiles := make([]string, 0, len(segments))
	for i, segmentURL := range segments {
		parsed, err := url.Parse(segmentURL)
		if err != nil {
			return "", err
		}
		segmentPath := filepath.Join(dir, fmt.Sprintf("segment-%03d-%s", i, filepath.Base(parsed.Path)))
		segmentData, err := fetchWithRadikoHeaders(segmentURL, auth, nil)
		if err != nil {
			return "", err
		}
		if err := os.WriteFile(segmentPath, segmentData, 0o600); err != nil {
			return "", err
		}
		segmentFiles = append(segmentFiles, segmentPath)
	}

	if len(segmentFiles) == 0 {
		return "", fmt.Errorf("no Radiko audio segments were downloaded")
	}

	listPath := filepath.Join(dir, "segments.txt")
	listContent := strings.Builder{}
	for _, segmentFile := range segmentFiles {
		listContent.WriteString("file '" + strings.ReplaceAll(segmentFile, "'", "'\\''") + "'\n")
	}
	if err := os.WriteFile(listPath, []byte(listContent.String()), 0o600); err != nil {
		return "", err
	}

	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		return "", fmt.Errorf("ffmpeg is required for Go direct Radiko download: %w", err)
	}
	outPath := filepath.Join(dir, "radiko-output.m4a")
	cmd := exec.CommandContext(ctx,
		ffmpegPath,
		"-hide_banner",
		"-loglevel", "error",
		"-y",
		"-f", "concat",
		"-safe", "0",
		"-i", listPath,
		"-c:a", "copy",
		outPath,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("go direct Radiko download failed: %w: %s", err, strings.TrimSpace(string(output)))
	}
	if _, err := os.Stat(outPath); err != nil {
		return "", err
	}
	return outPath, nil
}

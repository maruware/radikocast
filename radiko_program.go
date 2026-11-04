package radikocast

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ProgramInfo holds the basic metadata associated with a Radiko program.
type ProgramInfo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"webpage_url"`
	FT          string `json:"ft"`
	To          string `json:"to"`
	Timestamp   int64  `json:"timestamp"`
}

func (p *ProgramInfo) EndTime() time.Time {
	if p == nil || p.To == "" {
		return time.Time{}
	}
	parsed, err := time.ParseInLocation(datetimeLayout, p.To, location)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

type radikoProgramScheduleDocument struct {
	Stations []struct {
		Station []struct {
			ID    string `xml:"id,attr"`
			Progs struct {
				Programs []struct {
					FT          string `xml:"ft,attr"`
					To          string `xml:"to,attr"`
					Title       string `xml:"title"`
					URL         string `xml:"url"`
					Description string `xml:"desc"`
					Info        string `xml:"info"`
				} `xml:"prog"`
			} `xml:"progs"`
		} `xml:"station"`
	} `xml:"stations"`
}

func radikoProgramURL(stationID string, start time.Time) string {
	return fmt.Sprintf("https://radiko.jp/#!/ts/%s/%s", stationID, start.In(location).Format(datetimeLayout))
}

func decodeProgramInfoXML(body []byte, stationID string, startTime time.Time) (*ProgramInfo, error) {
	var doc radikoProgramScheduleDocument
	if err := xml.Unmarshal(body, &doc); err != nil {
		return nil, err
	}

	wantFT := startTime.In(location).Format(datetimeLayout)
	for _, stations := range doc.Stations {
		for _, station := range stations.Station {
			if !strings.EqualFold(station.ID, stationID) {
				continue
			}
			for _, prog := range station.Progs.Programs {
				if prog.FT != wantFT {
					continue
				}
				return &ProgramInfo{
					Title:       strings.TrimSpace(prog.Title),
					Description: strings.TrimSpace(prog.Description),
					URL:         strings.TrimSpace(prog.URL),
					FT:          prog.FT,
					To:          strings.TrimSpace(prog.To),
				}, nil
			}
		}
	}
	return nil, fmt.Errorf("program info for %s at %s not found", stationID, wantFT)
}

func getProgramInfo(ctx context.Context, stationID string, startTime time.Time) (*ProgramInfo, error) {
	dateKey := startTime.In(location).Format("20060102")
	urlString := fmt.Sprintf("https://radiko.jp/v3/program/station/date/%s/%s.xml", dateKey, stationID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlString, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("program metadata request failed: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return decodeProgramInfoXML(body, stationID, startTime)
}

func downloadRadikoAudio(ctx context.Context, stationID string, startTime, endTime time.Time, dir string) (string, error) {
	return downloadRadikoAudioGo(ctx, stationID, startTime, endTime, dir)
}

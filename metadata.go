package radikocast

import (
	"path/filepath"
	"time"
)

type MetaData struct {
	StartAt       time.Time `json:"start_at"`
	Title         string    `json:"title"`
	Desc          string    `json:"desc"`
	StartCode     string    `json:"start_code"`
	URL           string    `json:"url"`
	AudioFilename string    `json:"audio_filename"`
	AudioSize     int64     `json:"audio_size"`
}

// MetadataFromProg return metadata from prog
func NewMetadata(pg *ProgramInfo, audioPath string) (*MetaData, error) {
	loc := location
	if loc == nil {
		var err error
		loc, err = time.LoadLocation("Asia/Tokyo")
		if err != nil {
			return nil, err
		}
	}

	if pg == nil {
		pg = &ProgramInfo{}
	}

	startCode := pg.FT
	if startCode == "" {
		startCode = time.Now().In(loc).Format(datetimeLayout)
	}
	startAt, err := time.ParseInLocation(datetimeLayout, startCode, loc)
	if err != nil {
		return nil, err
	}

	name := filepath.Base(audioPath)
	size := fileSize(audioPath)
	return &MetaData{
		StartAt:       startAt,
		Title:         pg.Title,
		Desc:          pg.Description,
		StartCode:     startCode,
		URL:           pg.URL,
		AudioFilename: name,
		AudioSize:     size,
	}, nil
}

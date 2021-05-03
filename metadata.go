package radikocast

import (
	"path/filepath"
	"time"

	"github.com/yyoshiki41/go-radiko"
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
func NewMetadata(pg *radiko.Prog, audioPath string) (*MetaData, error) {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return nil, err
	}
	t, err := time.ParseInLocation("20060102150405", pg.Ft, loc)
	if err != nil {
		return nil, err
	}

	name := filepath.Base(audioPath)
	size := fileSize(audioPath)
	return &MetaData{
		StartAt:       t,
		Title:         pg.Title,
		Desc:          pg.Info,
		StartCode:     pg.Ft,
		URL:           pg.URL,
		AudioFilename: name,
		AudioSize:     size,
	}, nil
}

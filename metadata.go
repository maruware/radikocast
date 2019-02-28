package radikocast

import (
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
func MetadataFromProg(pg *radiko.Prog) *MetaData {
	loc, _ := time.LoadLocation("Asia/Tokyo")
	t, _ := time.ParseInLocation("20060102150405", pg.Ft, loc)
	return &MetaData{StartAt: t, Title: pg.Title, Desc: pg.Info, StartCode: pg.Ft, URL: pg.URL}
}

package radikocast

import "time"

const (
	tz                = "Asia/Tokyo"
	datetimeLayout    = "20060102150405"
	defaultConfigPath = "config.yml"
)

var (
	location *time.Location
)

func init() {
	var err error
	location, err = time.LoadLocation(tz)
	if err != nil {
		panic(err)
	}
}

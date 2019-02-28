package radikocast

import (
	"os"

	"github.com/mitchellh/cli"
)

var Ui cli.Ui

const (
	infoPrefix  = "INFO: "
	warnPrefix  = "WARN: "
	errorPrefix = "ERROR: "
)

func init() {
	Ui = &cli.PrefixedUi{
		InfoPrefix:  infoPrefix,
		WarnPrefix:  warnPrefix,
		ErrorPrefix: errorPrefix,
		Ui: &cli.BasicUi{
			Writer: os.Stdout,
		},
	}
}

func RecCommandFactory() (cli.Command, error) {
	return &recCommand{
		ui: Ui,
	}, nil
}
func RssCommandFactory() (cli.Command, error) {
	return &rssCommand{
		ui: Ui,
	}, nil
}
func PublishCommandFactory() (cli.Command, error) {
	return &publishCommand{
		ui: Ui,
	}, nil
}
func ScheduleCommandFactory() (cli.Command, error) {
	return &scheduleCommand{
		ui: Ui,
	}, nil
}

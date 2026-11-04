package radikocast

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mitchellh/cli"
)

type recCommand struct {
	ui cli.Ui
}

func (c *recCommand) Run(args []string) int {
	var stationID, start, areaID, bucket, format, outputDir string

	f := flag.NewFlagSet("rec", flag.ContinueOnError)
	f.StringVar(&stationID, "id", "", "id")
	f.StringVar(&start, "start", "", "start")
	f.StringVar(&start, "s", "", "start")
	f.StringVar(&areaID, "area", "", "area")
	f.StringVar(&areaID, "a", "", "area")
	f.StringVar(&bucket, "bucket", "", "bucket")
	f.StringVar(&format, "format", "mp3", "format")
	f.StringVar(&outputDir, "out", "", "out")
	f.StringVar(&outputDir, "output", "", "output")

	f.Usage = func() { c.ui.Error(c.Help()) }
	if err := f.Parse(args); err != nil {
		return 1
	}

	if stationID == "" {
		c.ui.Error("StationID is empty.")
		return 1
	}
	c.ui.Output("Now downloading.. ")
	var err error
	if bucket == "" {
		audioPath, metadataPath, recErr := RecProgramToLocal(stationID, start, areaID, format, outputDir)
		if recErr != nil {
			c.ui.Error(recErr.Error())
			return 1
		}
		c.ui.Output(fmt.Sprintf("Saved audio: %s", audioPath))
		c.ui.Output(fmt.Sprintf("Saved metadata: %s", metadataPath))
		c.ui.Output(fmt.Sprintf("Completed!"))
		return 0
	}
	err = RecAndUploadProgram(stationID, start, areaID, bucket, format)
	if err != nil {
		c.ui.Error(err.Error())
		return 1
	}
	c.ui.Output(fmt.Sprintf("Completed!"))
	return 0
}

func (c *recCommand) Synopsis() string {
	return "Record a radiko program"
}

func (c *recCommand) Help() string {
	return strings.TrimSpace(`
Usage: radikocast rec [options]
  Record a radiko program.
Options:
  -id=name                 Station id
  -start,s=201610101000    Start time
  -area,a=name             Area id
	-bucket=bucketname	   S3 bucket name (optional)
	-out,-output=dir         Local output directory (used when -bucket is omitted)
`)
}

func fileSize(path string) int64 {
	fi, err := os.Stat(path)
	if err != nil {
		log.Fatal(err)
	}

	return fi.Size()
}

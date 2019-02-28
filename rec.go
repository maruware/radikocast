package radikocast

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/olekukonko/tablewriter"
	"github.com/yyoshiki41/go-radiko"
	"github.com/yyoshiki41/radigo"
)

func RecProgram(stationID string, start string, areaID string, outputDir string) (*string, error) {
	fmt.Printf("Rec %s %s\n", stationID, start)
	startTime, err := time.ParseInLocation(datetimeLayout, start, location)
	if err != nil {
		return nil, err
	}

	output := radigo.OutputConfig{
		DirFullPath:  outputDir,
		FileBaseName: fmt.Sprintf("%s-%s", startTime.In(location).Format(datetimeLayout), stationID),
		FileFormat:   radigo.AudioFormatAAC,
	}

	if err := output.SetupDir(); err != nil {
		return nil, err
	}

	if output.IsExist() {
		return nil, fmt.Errorf("Dup rec: %s", output.AbsPath())
	}

	spin := spinner.New(spinner.CharSets[9], time.Second)
	spin.Start()
	defer spin.Stop()

	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	client, err := getClient(ctx, areaID)
	if err != nil {
		return nil, err
	}

	_, err = client.AuthorizeToken(ctx)
	if err != nil {
		return nil, err
	}

	uri, err := client.TimeshiftPlaylistM3U8(ctx, stationID, startTime)
	if err != nil {
		return nil, err
	}

	chunklist, err := radiko.GetChunklistFromM3U8(uri)
	if err != nil {
		return nil, err
	}

	aacDir, err := output.TempAACDir()
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(aacDir) // clean up

	if err := BulkDownload(chunklist, aacDir); err != nil {
		return nil, err
	}

	concatedFile, err := radigo.ConcatAACFilesFromList(ctx, aacDir)
	if err != nil {
		return nil, err
	}

	retErr := os.Rename(concatedFile, output.AbsPath())
	if retErr != nil {
		return nil, retErr
	}

	// dump metadata
	pg, err := client.GetProgramByStartTime(ctx, stationID, startTime)
	if err != nil {
		ctxCancel()
		return nil, err
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"STATION ID", "TITLE", "DESC"})
	table.Append([]string{stationID, pg.Title, pg.Info})
	fmt.Print("\n")
	table.Render()

	metadata := MetadataFromProg(pg)
	metadata.AudioFilename = filepath.Base(output.AbsPath())
	metadata.AudioSize = fileSize(output.AbsPath())
	metadataPath := strings.Replace(output.AbsPath(), ".aac", ".json", 1)

	jsonByte, _ := json.Marshal(*metadata)
	ioutil.WriteFile(metadataPath, jsonByte, os.ModePerm)

	return &output.FileBaseName, nil
}

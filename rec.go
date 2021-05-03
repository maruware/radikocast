package radikocast

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/olekukonko/tablewriter"
	"github.com/yyoshiki41/go-radiko"
	"github.com/yyoshiki41/radigo"
)

func RecProgram(stationID string, start string, areaID string) error {
	fmt.Printf("Rec %s %s\n", stationID, start)

	r, err := recProgram(stationID, start, areaID)
	if err != nil {
		return fmt.Errorf("rec error: %w", err)
	}
	defer r.Dispose()

	return nil
}

func RecAndUploadProgram(stationID string, start string, areaID string, bucket string) error {
	fmt.Printf("Rec %s %s\n", stationID, start)

	r, err := recProgram(stationID, start, areaID)
	if err != nil {
		return fmt.Errorf("rec error: %w", err)
	}
	defer r.Dispose()

	// put s3
	fmt.Printf("bucket: %s\n", bucket)
	err = putProgramToS3(r.audioPath, r.metadataPath, bucket)
	if err != nil {
		return fmt.Errorf("upload error: %w", err)
	}
	return nil
}

type recResult struct {
	audioPath    string
	metadataPath string

	tmpdir string
}

func (r *recResult) Dispose() {
	os.RemoveAll(r.tmpdir)
}

func recProgram(stationID string, start string, areaID string) (*recResult, error) {
	startTime, err := time.ParseInLocation(datetimeLayout, start, location)
	if err != nil {
		return nil, err
	}

	td, err := ioutil.TempDir("", "radikocast")
	if err != nil {
		return nil, err
	}

	output := radigo.OutputConfig{
		DirFullPath:  td,
		FileBaseName: fmt.Sprintf("%s-%s", startTime.In(location).Format(datetimeLayout), stationID),
		FileFormat:   radigo.AudioFormatMP3,
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	err = radigo.ConvertAACtoMP3(ctx, concatedFile, output.AbsPath())
	if err != nil {
		return nil, err
	}

	// dump metadata
	pg, err := client.GetProgramByStartTime(ctx, stationID, startTime)
	if err != nil {
		return nil, err
	}

	printInfo(stationID, pg.Title, pg.Info)

	metadata, err := NewMetadata(pg, output.AbsPath())
	if err != nil {
		return nil, fmt.Errorf("metadata error: %w", err)
	}

	metadataPath := strings.Replace(output.AbsPath(), ".mp3", ".json", 1)
	if err := writeJson(&metadata, metadataPath); err != nil {
		return nil, fmt.Errorf("failed to write metadata: %w", err)
	}

	return &recResult{
		audioPath:    output.AbsPath(),
		metadataPath: metadataPath,
	}, nil
}

func printInfo(stationID, title, desc string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"STATION ID", "TITLE", "DESC"})
	table.Append([]string{stationID, title, desc})
	fmt.Print("\n")
	table.Render()
}

func writeJson(data interface{}, dst string) error {
	f, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	return enc.Encode(data)
}

func putProgramToS3(audioPath string, metadataPath string, bucket string) error {
	storage := NewS3(bucket)
	err := storage.PutObjectFromFile(audioPath, "audio/mpeg")
	if err != nil {
		return err
	}
	err = storage.PutObjectFromFile(metadataPath, "application/json")
	if err != nil {
		return err
	}

	return nil
}

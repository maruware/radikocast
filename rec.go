package radikocast

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/olekukonko/tablewriter"
)

func RecProgram(stationID string, start string, areaID string, format string) error {
	fmt.Printf("Rec %s %s\n", stationID, start)

	r, err := recProgram(stationID, start, areaID, format)
	if err != nil {
		return fmt.Errorf("rec error: %w", err)
	}
	defer r.Dispose()

	return nil
}

func RecProgramToLocal(stationID string, start string, areaID string, format string, outputDir string) (string, string, error) {
	fmt.Printf("Rec %s %s\n", stationID, start)

	r, err := recProgramToLocal(stationID, start, areaID, format, outputDir)
	if err != nil {
		return "", "", fmt.Errorf("rec error: %w", err)
	}
	defer r.Dispose()

	return r.audioPath, r.metadataPath, nil
}

func RecAndUploadProgram(stationID string, start string, areaID string, bucket string, format string) error {
	fmt.Printf("Rec %s %s\n", stationID, start)

	r, err := recProgram(stationID, start, areaID, format)
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

func recProgram(stationID string, start string, areaID string, format string) (*recResult, error) {
	startTime, err := time.ParseInLocation(datetimeLayout, start, location)
	if err != nil {
		return nil, err
	}

	td, err := os.MkdirTemp("", "radikocast")
	if err != nil {
		return nil, err
	}

	var ff string
	switch format {
	case "m4a":
		ff = AudioFormatM4A
	case "mp3":
		ff = AudioFormatMP3
	default:
		return nil, fmt.Errorf("Bad format: %s", format)
	}

	output := OutputConfig{
		DirFullPath:  td,
		FileBaseName: fmt.Sprintf("%s-%s", startTime.In(location).Format(datetimeLayout), stationID),
		FileFormat:   ff,
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

	pg, err := getProgramInfo(ctx, stationID, startTime)
	if err != nil {
		pg = &ProgramInfo{
			Title: stationID,
			URL:   radikoProgramURL(stationID, startTime),
			FT:    startTime.In(location).Format(datetimeLayout),
			To:    startTime.Add(15 * time.Minute).In(location).Format(datetimeLayout),
		}
	}
	programEndTime := startTime.Add(15 * time.Minute)
	if end := pg.EndTime(); !end.IsZero() {
		programEndTime = end
	}

	downloadedFile, err := downloadRadikoAudio(ctx, stationID, startTime, programEndTime, td)
	if err != nil {
		return nil, err
	}
	defer os.Remove(downloadedFile)

	cv, err := NewConverter(ff)
	if err != nil {
		return nil, err
	}

	if err := cv.Convert(ctx, downloadedFile, output.AbsPath()); err != nil {
		return nil, err
	}

	printInfo(stationID, pg.Title, pg.Description)

	metadata, err := NewMetadata(pg, output.AbsPath())
	if err != nil {
		return nil, fmt.Errorf("metadata error: %w", err)
	}

	metadataPath := strings.Replace(output.AbsPath(),
		fmt.Sprintf(".%s", output.AudioExt()), ".json", 1)
	if err := writeJson(&metadata, metadataPath); err != nil {
		return nil, fmt.Errorf("failed to write metadata: %w", err)
	}

	return &recResult{
		audioPath:    output.AbsPath(),
		metadataPath: metadataPath,
		tmpdir:       td,
	}, nil
}

func recProgramToLocal(stationID string, start string, areaID string, format string, outputDir string) (*recResult, error) {
	startTime, err := time.ParseInLocation(datetimeLayout, start, location)
	if err != nil {
		return nil, err
	}

	td, err := os.MkdirTemp("", "radikocast")
	if err != nil {
		return nil, err
	}

	var ff string
	switch format {
	case "m4a":
		ff = AudioFormatM4A
	case "mp3":
		ff = AudioFormatMP3
	default:
		return nil, fmt.Errorf("Bad format: %s", format)
	}

	outputCfg, err := NewOutputConfig(fmt.Sprintf("%s-%s", startTime.In(location).Format(datetimeLayout), stationID), ff)
	if err != nil {
		return nil, err
	}
	if outputDir != "" {
		if filepath.IsAbs(outputDir) {
			outputCfg.DirFullPath = filepath.Clean(outputDir)
		} else {
			wd, err := os.Getwd()
			if err != nil {
				return nil, err
			}
			outputCfg.DirFullPath = filepath.Join(wd, outputDir)
		}
	}

	if err := outputCfg.SetupDir(); err != nil {
		return nil, err
	}

	if outputCfg.IsExist() {
		return nil, fmt.Errorf("Dup rec: %s", outputCfg.AbsPath())
	}

	spin := spinner.New(spinner.CharSets[9], time.Second)
	spin.Start()
	defer spin.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pg, err := getProgramInfo(ctx, stationID, startTime)
	if err != nil {
		pg = &ProgramInfo{
			Title: stationID,
			URL:   radikoProgramURL(stationID, startTime),
			FT:    startTime.In(location).Format(datetimeLayout),
			To:    startTime.Add(15 * time.Minute).In(location).Format(datetimeLayout),
		}
	}
	programEndTime := startTime.Add(15 * time.Minute)
	if end := pg.EndTime(); !end.IsZero() {
		programEndTime = end
	}

	downloadedFile, err := downloadRadikoAudio(ctx, stationID, startTime, programEndTime, td)
	if err != nil {
		return nil, err
	}
	defer os.Remove(downloadedFile)

	cv, err := NewConverter(ff)
	if err != nil {
		return nil, err
	}

	if err := cv.Convert(ctx, downloadedFile, outputCfg.AbsPath()); err != nil {
		return nil, err
	}

	printInfo(stationID, pg.Title, pg.Description)

	metadata, err := NewMetadata(pg, outputCfg.AbsPath())
	if err != nil {
		return nil, fmt.Errorf("metadata error: %w", err)
	}

	metadataPath := strings.Replace(outputCfg.AbsPath(),
		fmt.Sprintf(".%s", outputCfg.AudioExt()), ".json", 1)
	if err := writeJson(&metadata, metadataPath); err != nil {
		return nil, fmt.Errorf("failed to write metadata: %w", err)
	}

	return &recResult{
		audioPath:    outputCfg.AbsPath(),
		metadataPath: metadataPath,
		tmpdir:       td,
	}, nil
}

func printInfo(stationID, title, desc string) {
	table := tablewriter.NewTable(os.Stdout)
	table.Header([]string{"STATION ID", "TITLE", "DESC"})
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

package radikocast

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/eduncan911/podcast"
)

func GenerateRss(title string, host string, image string, bucket string) (*podcast.Podcast, error) {
	storage := NewS3(bucket)

	contents, err := storage.Scan()
	if err != nil {
		return nil, err
	}
	jsonContents := []*s3.Object{}
	for _, c := range contents {
		if strings.HasSuffix(*c.Key, ".json") {
			jsonContents = append(jsonContents, c)
		}
	}

	now := time.Now()
	feed := podcast.New(title, host, "", nil, &now)

	if len(image) > 0 {
		feed.IImage = &podcast.IImage{HREF: image}
	}
	feed.Language = "ja"
	for _, j := range jsonContents {
		var metadata MetaData
		o, err := storage.GetObject(j)
		if err != nil {
			return nil, err
		}
		defer o.Body.Close()
		d := json.NewDecoder(o.Body)
		if err := d.Decode(&metadata); err != nil {
			return nil, err
		}

		item, err := generateItemNode(&metadata, host)
		if err != nil {
			return nil, err
		}
		if _, err := feed.AddItem(*item); err != nil {
			return nil, fmt.Errorf("Item %s error: %w", metadata.AudioFilename, err)
		}
	}

	return &feed, nil
}

func PutRss(rss *podcast.Podcast, bucket string, feedName string) error {
	buf := bytes.NewBuffer(nil)
	if err := rss.Encode(buf); err != nil {
		return err
	}

	storage := NewS3(bucket)
	err := storage.PutObject(feedName, buf, "application/rss+xml; charset=utf-8")
	if err != nil {
		return err
	}

	return nil
}

func generateItemNode(metadata *MetaData, host string) (*podcast.Item, error) {
	url := fmt.Sprintf("%s/%s", host, metadata.AudioFilename)
	ext := filepath.Ext(metadata.AudioFilename)
	var etype podcast.EnclosureType
	switch ext {
	case ".aac":
		etype = podcast.M4A
	case ".m4a":
		etype = podcast.M4A
	case ".mp3":
		etype = podcast.MP3
	default:
		return nil, fmt.Errorf("no support file format: %s", ext)
	}

	// Description required
	desc := metadata.Desc
	if len(desc) <= 0 {
		desc = metadata.Title
	}
	return &podcast.Item{
		Title: metadata.Title,
		Enclosure: &podcast.Enclosure{
			URL:    url,
			Length: metadata.AudioSize,
			Type:   etype,
		},
		Description: desc,
		PubDate:     &metadata.StartAt,
		GUID:        metadata.AudioFilename,
	}, nil
}

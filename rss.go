package radikocast

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/eduncan911/podcast"
)

func GenerateRss(config *ConfigPodcast, dir string) podcast.Podcast {
	host := config.Host
	metadataPaths, _ := filepath.Glob(filepath.Join(dir, "*.json"))

	now := time.Now()
	feed := podcast.New(config.Title, host, "", nil, &now)
	feed.IImage = &podcast.IImage{HREF: config.Image}
	feed.Language = "ja"
	for _, metadataPath := range metadataPaths {
		var metadata MetaData
		jsonBytes, _ := ioutil.ReadFile(metadataPath)
		json.Unmarshal(jsonBytes, &metadata)

		item := generateItemNode(&metadata, host)
		feed.AddItem(*item)
	}

	return feed
}

func WriteRss(rss podcast.Podcast, dst string) error {
	file, err := os.OpenFile(dst, os.O_WRONLY, os.ModePerm)
	defer file.Close()
	if err != nil {
		return err
	}
	err = rss.Encode(file)
	if err != nil {
		return err
	}
	return nil
}

func generateItemNode(metadata *MetaData, host string) *podcast.Item {
	url := fmt.Sprintf("%s/%s", host, metadata.AudioFilename)
	return &podcast.Item{
		Title: metadata.Title,
		Enclosure: &podcast.Enclosure{
			URL:    url,
			Length: metadata.AudioSize,
		},
		Description: metadata.Desc,
		PubDate:     &metadata.StartAt,
		GUID:        metadata.AudioFilename,
	}
}

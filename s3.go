package radikocast

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func SyncDirToS3(dir string, bucket string) error {
	sess := session.Must(session.NewSession())
	svc := s3.New(sess)

	res, _ := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: &bucket,
	})
	contents := res.Contents

	aacPaths, _ := filepath.Glob(filepath.Join(dir, "*.aac"))
	for _, aacPath := range aacPaths {
		_, err := syncObject(svc, &bucket, aacPath, "audio/aac", contents)
		if err != nil {
			return err
		}
	}

	xmlPaths, _ := filepath.Glob(filepath.Join(dir, "*.xml"))
	for _, xmlPath := range xmlPaths {
		_, err := syncObject(svc, &bucket, xmlPath, "application/rss+xml; charset=utf-8", contents)
		if err != nil {
			return err
		}
	}
	return nil
}

func syncObject(svc *s3.S3, bucket *string, filePath string, contentType string, contents []*s3.Object) (bool, error) {
	key := filepath.Base(filePath)
	md5, err := HashFileMd5(filePath)
	if err != nil {
		return false, err
	}
	etag := fmt.Sprintf("\"%s\"", md5)

	var sameContent *s3.Object = nil
	for _, content := range contents {
		if *content.Key == key && *content.ETag == etag {
			sameContent = content
		}
	}

	if sameContent != nil {
		fmt.Println("[Skip]", key)
		return false, nil
	}

	f, err := os.Open(filePath)
	if err != nil {
		return false, err
	}

	svc.PutObject(&s3.PutObjectInput{
		Bucket:      bucket,
		Key:         &key,
		Body:        f,
		ContentType: &contentType,
		ACL:         aws.String("public-read"),
	})

	fmt.Println("[Put]", key)

	return true, nil
}

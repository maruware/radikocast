package radikocast

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// func SyncDirToS3(dir string, bucket string) error {
// 	sess := session.Must(session.NewSession())
// 	svc := s3.New(sess)

// 	res, _ := svc.ListObjectsV2(&s3.ListObjectsV2Input{
// 		Bucket: &bucket,
// 	})
// 	contents := res.Contents

// 	aacPaths, _ := filepath.Glob(filepath.Join(dir, "*.aac"))
// 	for _, aacPath := range aacPaths {
// 		_, err := syncObject(svc, &bucket, aacPath, "audio/aac", contents)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	jsonPaths, _ := filepath.Glob(filepath.Join(dir, "*.json"))
// 	for _, jsonPath := range jsonPaths {
// 		_, err := syncObject(svc, &bucket, jsonPath, "application/json", contents)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	xmlPaths, _ := filepath.Glob(filepath.Join(dir, "*.xml"))
// 	for _, xmlPath := range xmlPaths {
// 		_, err := syncObject(svc, &bucket, xmlPath, "application/rss+xml; charset=utf-8", contents)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

type S3 struct {
	Bucket string
	svc    *s3.S3
}

func NewS3(bucket string) *S3 {
	sess := session.Must(session.NewSession())
	svc := s3.New(sess)
	return &S3{
		Bucket: bucket,
		svc:    svc,
	}
}

func (s *S3) syncObject(filePath string, contentType string, contents []*s3.Object) (bool, error) {
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

	_, err = s.svc.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(s.Bucket),
		Key:         &key,
		Body:        f,
		ContentType: &contentType,
		ACL:         aws.String("public-read"),
	})
	if err != nil {
		return false, err
	}

	fmt.Println("[Put]", key)

	return true, nil
}

func (s *S3) PutObjectFromFile(filePath string, contentType string) error {
	key := filepath.Base(filePath)

	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = s.svc.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(s.Bucket),
		Key:         &key,
		Body:        f,
		ContentType: &contentType,
		ACL:         aws.String("public-read"),
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *S3) PutObject(key string, content io.Reader, contentType string) error {
	uploader := s3manager.NewUploaderWithClient(s.svc)
	_, err := uploader.Upload(&s3manager.UploadInput{
		ACL:         aws.String("public-read"),
		Bucket:      aws.String(s.Bucket),
		Key:         aws.String(key),
		Body:        content,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *S3) Scan() ([]*s3.Object, error) {
	contents := []*s3.Object{}
	var continuationToken *string = nil
	for {
		res, err := s.svc.ListObjectsV2(&s3.ListObjectsV2Input{
			Bucket:            aws.String(s.Bucket),
			ContinuationToken: continuationToken,
		})
		if err != nil {
			return nil, err
		}
		contents = append(contents, res.Contents...)
		if *res.IsTruncated {
			continuationToken = res.NextContinuationToken
		} else {
			break
		}
	}

	return contents, nil
}

func (s *S3) GetObject(o *s3.Object) (*s3.GetObjectOutput, error) {
	sess := session.Must(session.NewSession())
	svc := s3.New(sess)

	r, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    o.Key,
	})
	if err != nil {
		return nil, err
	}
	return r, nil
}

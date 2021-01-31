// Package manager ...
package manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// S3Session -
type S3Session struct {
	session *session.Session
}

// S3UploadObject -
type S3UploadObject struct {
	Data []byte       `json:"Data"`
	Meta S3UploadMeta `json:"Meta"`
}

// S3UploadMeta -
type S3UploadMeta struct {
	Hash string `json:"Hash"`
	Name string `json:"Name"`
	Path string `json:"Path"`
}

// NewS3Session - Initialize S3 Connection
func NewS3Session() *S3Session {
	var s S3Session = S3Session{}
	s.initializeSession()
	return &s
}

func (s *S3Session) initializeSession() {

	// Initialize S3 Client
	s3Client, err := session.NewSession(
		&aws.Config{
			Region:                        aws.String(os.Getenv("AWS_DEFAULT_REGION")),
			CredentialsChainVerboseErrors: aws.Bool(true),
			Credentials:                   credentials.NewEnvCredentials(),
		},
	)

	if err != nil {
		log.Fatal(err)
	}

	s.session = s3Client
}

// DownloadFeatureFromS3 - Could return pointer to bytes instead...
func (s *S3Session) DownloadFeatureFromS3(sourceBucket string, fileKey string) ([]byte, error) {

	var err error
	svc := s3.New(session.Must(s.session, err))

	output, err := svc.GetObject(
		&s3.GetObjectInput{
			Bucket: aws.String(sourceBucket), // Environment Variable
			Key:    aws.String(fileKey),
		},
	)

	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(output.Body)
	if err != nil {
		log.Fatal(err)
	}

	return body, nil

}

// StartS3UploadWorker - Originally Boosted from: https://golangcode.com/uploading-a-file-to-s3/
// AddObjectToS3 will upload a single file to S3, it will require a pre-built aws session
// and will set file info like content type and encryption on the uploaded file.
// TODO: Error logging channel
func (s *S3Session) StartS3UploadWorker(i int, targetBucket string, jobs <-chan *S3UploadObject, wg *sync.WaitGroup) {
	defer wg.Done()

	var err error

	// Init Session if not Initialized
	if s.session == nil {
		s.initializeSession()
	}

	svc := s3.New(session.Must(s.session, err))

	// Check IF File Exists
	for s3Upload := range jobs {
		log.Info("Worker %d Recieved: %+v\n", i, s3Upload.Meta)
		var fileKey string = fmt.Sprintf("%s.geojson", s3Upload.Meta.Hash)

		// Check if Head Exists
		output, _ := svc.HeadObject(
			&s3.HeadObjectInput{
				Bucket: aws.String(targetBucket), // Environment Variable
				Key:    aws.String(fileKey),
			},
		)

		// If file DNE - Send to S3
		if output.ContentLength == nil {

			// Send the Main Content to Main Folder
			_, err = svc.PutObject(&s3.PutObjectInput{
				Bucket:               aws.String(targetBucket),
				Key:                  aws.String(fileKey),
				Body:                 bytes.NewReader(s3Upload.Data), // QUESTION: Does this waste space???
				ContentLength:        aws.Int64(int64(len(s3Upload.Data))),
				ContentType:          aws.String(http.DetectContentType(s3Upload.Data)),
				ContentDisposition:   aws.String("attachment"),
				ServerSideEncryption: aws.String("AES256"),
				ACL:                  aws.String("private"),
			})
			if err != nil {
				log.Warn(err)
			}

			// Send the Meta to a Meta Folder

			// Generate Metadata Object
			metaContent, _ := json.Marshal(s3Upload.Meta)

			_, err = svc.PutObject(&s3.PutObjectInput{
				Bucket:               aws.String(targetBucket),
				Key:                  aws.String(fmt.Sprintf("meta/%s_meta.json", s3Upload.Meta.Hash)),
				Body:                 bytes.NewReader(metaContent), // QUESTION: Does this waste space???
				ContentLength:        aws.Int64(int64(len(metaContent))),
				ContentType:          aws.String(http.DetectContentType(metaContent)),
				ContentDisposition:   aws.String("attachment"),
				ServerSideEncryption: aws.String("AES256"),
				ACL:                  aws.String("private"),
			})
			if err != nil {
				log.Warn(err)
			}
		}
	}
}

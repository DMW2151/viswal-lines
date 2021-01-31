// package comment...
package main

/*
NOTES:
	- boosted from https://github.com/aws/aws-lambda-go/blob/master/events/README_S3.md
*/

import (
	"aws-lambda-viswal/pkg/manager"
	"aws-lambda-viswal/pkg/viswal"
	"context"
	"crypto/md5"
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	log "github.com/sirupsen/logrus"
)

var (
	s3Region       string = os.Getenv("S3_SHAPES_DEFAULT_REGION")
	s3SourceBucket string = os.Getenv("S3_SHAPES_SRC_BUCKET")
	s3TargetBucket string = os.Getenv("S3_SHAPES_TARGET_BUCKET")
)

func handler(ctx context.Context, s3Event events.S3Event) (string, error) {

	var s3 events.S3Entity
	var s3Object manager.S3UploadObject

	// For each event, download to memory, decompose to features
	// and upload as new source...
	for _, record := range s3Event.Records {

		s3 = record.S3

		// Download the object from S3...
		b, err := s.DownloadFeatureFromS3(s3.Bucket.Name, s3.Object.Key)

		if err != nil {
			log.WithFields(log.Fields{"Key": s3.Object.Key}).Fatal("Failed S3 Download")
		}

		// Begin Feature Processing
		fc, _ := viswal.BatchReduceGEOJSON(b)

		// Send results to S3 Upload Workers
		for _, feature := range fc.Features {

			featureData, err := feature.MarshalJSON()
			if err != nil {
				log.Warn(err)
			}

			// Get Name - Safely
			featureName := feature.Properties["name"]
			if featureName == nil {
				featureName = ""
			}

			// Send object...
			s3Object = manager.S3UploadObject{
				Data: featureData,
				Meta: manager.S3UploadMeta{
					Hash: fmt.Sprintf("%x", md5.Sum(featureData)),
					Name: featureName.(string),
				},
			}

			workerPool <- &s3Object
		}

		close(workerPool)

	}

	// Token Return - For "Fun"
	return "Finished", nil
}

// Initialize S3 Connection && Pool to Communicate Uploads on...
var s = manager.NewS3Session()
var workerPool = make(chan *manager.S3UploadObject, 10)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.WarnLevel)
}

func main() {
	//Set Feature S3 Upload Concurrency & Start N workers...
	workerConcurrency, _ := strconv.Atoi(os.Getenv("S3_WORKER_CONCURRENCY"))

	for i := 0; i < workerConcurrency; i++ {
		go s.StartS3UploadWorker(s3TargetBucket, workerPool)
	}

	// Make the handler available for Remote Procedure Call by AWS Lambda
	lambda.Start(handler)
}

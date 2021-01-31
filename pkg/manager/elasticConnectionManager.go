// Package manager ...
package manager

import (
	"context"
	"encoding/json"
	"os"

	log "github.com/sirupsen/logrus"

	elastic "github.com/olivere/elastic/v7"
)

var (
	elasticDefaultIndex string = os.Getenv("ELASTIC_DEFAULT_INDEX")
	elasticHost         string = os.Getenv("ELASTIC_HOST")
)

// ElasticClient - Used to connect to ES and make Bulk Requests
type ElasticClient struct {
	ConnectionParams map[string]string
	Client           *elastic.Client
	Context          context.Context
	Request          *elastic.BulkService
	IndexName        string
}

// NewElasticClient -
func NewElasticClient() ElasticClient {

	// Init Elastic Client...
	e := ElasticClient{
		ConnectionParams: map[string]string{
			"Host": elasticHost,
		},
		IndexName: elasticDefaultIndex,
	}

	e.initializeClient()

	return e

}

func (e *ElasticClient) initializeClient() {

	client, err := elastic.NewClient(
		elastic.SetURL(e.ConnectionParams["Host"]),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false),
	)

	if err != nil {
		log.Fatal(err)
	}

	e.Client = client

}

// ExecuteRequest -
func (e *ElasticClient) ExecuteRequest() {

	if !(e.hasContext()) {
		e.initializeContext()
	}

	if _, err := e.Request.Do(e.Context); err != nil {
		log.WithFields(
			log.Fields{
				"Host":  e.ConnectionParams["Host"],
				"Index": e.IndexName,
			},
		).Error("Elasticsearch Request Error", err)
	}

}

func (e *ElasticClient) hasRequest() bool {
	return e.Request != nil
}

func (e *ElasticClient) hasContext() bool {
	return e.Context != nil
}

func (e *ElasticClient) hasClient() bool {
	return e.Client != nil
}

func (e *ElasticClient) initializeContext() {
	// NOTE: Learn More about what Each context Means...
	if e.Context == nil {
		e.Context = context.Background()
	}
}

func (e *ElasticClient) initializeBulkRequest() {
	e.Request = e.Client.
		Bulk().
		Index(e.IndexName)
}

// PrepareInsert -
func (e *ElasticClient) PrepareInsert(entry *S3UploadMeta) {

	// Initialize Request if Not Yet Exists - Send to Elastic
	if !(e.hasRequest()) {
		e.newRequest()
	}

	b, err := json.Marshal(entry)
	if err != nil {
		// Literally Should NEVER happen...
		log.WithFields(
			log.Fields{
				"Path": entry.Path,
				"Name": entry.Name,
			},
		).Warn(
			"Failed To Add Feature to Index Request", err,
		)
	} else {
		var id string = entry.Hash
		// Add to ES Request...
		e.Request.Add(
			elastic.NewBulkIndexRequest().Id(id).Doc(string(b)),
		)
	}

}

func (e *ElasticClient) newRequest() {

	// Initialize New Request
	if !(e.hasClient()) {
		e.initializeClient()
	}

	// Add Each New Document
	if !(e.hasRequest()) {
		e.initializeBulkRequest()
	}

}

package main

import (
	manager "aws-lambda-viswal/pkg/manager"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-lambda-go/events"
	elastic "github.com/olivere/elastic/v7"
)

var tpl *template.Template

var (
	elasticDefaultIndex string = os.Getenv("ELASTIC_DEFAULT_INDEX")
	elasticHost         string = os.Getenv("ELASTIC_HOST")
)

// QueryMSG  - From frontend...
type QueryMSG struct {
	QueryString string `json:"queryString"`
}

// Index - Handler for Index Page...
func index(w http.ResponseWriter, req *http.Request) {
	tpl.ExecuteTemplate(w, "index.html", nil)
}

// _subscription - Handler for Subscribed Endpoint
// a lot of notes about handling SNS vs S3 forwarded messages...
// IDK If I need to verify anything here??
func _subscription(w http.ResponseWriter, req *http.Request) {

	// Check Header "X-Amz-Sns-Message-Type" for Action...
	msgType := req.Header.Get("X-Amz-Sns-Message-Type")
	if msgType == "Notification" {
		_handleS3Event(w, req)
	}

	// Only expect Notifications; If Request != Notification
	// Either a subscription message or a mistake...
	w.WriteHeader(http.StatusBadRequest) // TODO - Update this Error...
}

// get location of the metafile and insert into elastic - Treat as S3 Event
func _handleS3Event(w http.ResponseWriter, req *http.Request) {

	var events events.S3Event
	var e manager.S3UploadMeta

	// Unmarshall S3 Event...
	content, _ := ioutil.ReadAll(req.Body)
	_ = json.Unmarshal(content, &events)

	// Get Contents of S3 Meta, Download the Meta File, & Write to Elastic
	for _, record := range events.Records {
		b, _ := s3m.DownloadFeatureFromS3(record.S3.Bucket.Name, record.S3.Object.Key)
		json.Unmarshal(b, &e)
		log.Info("%+v\n", e)

		esm.PrepareInsert(&e)
		esm.ExecuteRequest()
		w.WriteHeader(http.StatusOK)
	}

}

func _autocomplete(w http.ResponseWriter, r *http.Request) {

	var (
		q       QueryMSG
		entry   manager.S3UploadMeta
		err     error
		entries = []string{}
	)

	decoder := json.NewDecoder(r.Body)
	_ = decoder.Decode(&q)

	// TODO - Check if QueryString is null, Shouldn't matter,
	// not even worth Info Logging, really...
	ctx := context.Background()

	searchSuggester := elastic.
		NewCompletionSuggester(elasticDefaultIndex).
		Text(q.QueryString).
		Field("Name").
		Size(5)

	searchSource := elastic.NewSearchSource().
		Suggester(searchSuggester)

	searchResult, err := esm.Client.Search().
		Index(elasticDefaultIndex).
		SearchSource(searchSource).
		Do(ctx)

	if err != nil {
		// TODO: Differentiate between "NoResultsError" and Something Significant...
		fmt.Println("Failed...", err)
	} else {
		if searchResult.Suggest != nil {
			// Get results and save to a slice...
			for _, ops := range searchResult.Suggest[elasticDefaultIndex] {
				for _, op := range ops.Options {
					_ = json.Unmarshal(op.Source, &entry)
					entries = append(entries, entry.Name)
				}
			}
		}
	}

	// Send results back to responseWriter...
	json.NewEncoder(w).Encode(entries)
}

// Start Elastic Manager and S3 Client Manager...
var s3m = manager.NewS3Session()
var esm = manager.NewElasticClient()

func init() {
	fmt.Println("Init")
	tpl = template.Must(template.ParseGlob("./api/templates/*"))
}

func main() {
	// Add routes to serve home and download pages
	http.HandleFunc("/", index)
	http.HandleFunc("/search", _autocomplete)
	http.HandleFunc("/_sub", _subscription)
	http.Handle("/favicon.ico", http.NotFoundHandler())
	http.ListenAndServe(":8081", nil)
}

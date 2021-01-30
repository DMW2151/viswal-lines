package main

import (
	viswal "aws-lambda-viswal/pkg/viswal"
	"fmt"
	"io/ioutil"
	"os"
)

var r *viswal.Reducer
var err error

func main() {

	// Read content
	content, _ := os.Open("/Users/dustinwilson/Desktop/Personal/viswal/data/chicago.geojson")
	b, _ := ioutil.ReadAll(content)

	// Reduce - Batch
	fc, _ := viswal.BatchReduceGEOJSON(b)
	fmt.Println(fc.Features[0].Properties["Order"])

	// // Reduce Normal
	// fc1, _ := geojson.UnmarshalFeatureCollection(b)

	// // Initialize Reducer
	// r = &viswal.Reducer{
	// 	Data: fc1.Features,
	// }

}

// Package viswal -
package viswal

import (
	"log"
	"sync"

	geojson "github.com/paulmach/go.geojson"
)

// Reducer - Reads data from some source,
type Reducer struct {
	Data []*geojson.Feature
}

// reduceFeaturesAsyncWrapper
func (r *Reducer) reduceFeaturesAsyncWrapper(index int, wg *sync.WaitGroup) {
	defer wg.Done()
	err := r.ReduceFeature(index)
	if err != nil {
		log.Fatal(err)
	}
}

// ReduceFeature - wraper around geom. reducing method
func (r *Reducer) ReduceFeature(index int) error {

	// Reduce geometry - route to corret type
	order, err := ReduceGeometry(r.Data[index].Geometry)
	if err != nil {
		return err
	}

	// Set `Order` to the reducer's feature
	r.Data[index].Properties["Order"] = order
	return nil
}

// ReduceGeometry - Find the shape of the polygon and reduce
/*
NOTES: 2D Polygons fall into one of the following categories:
	- [][][][]float64 -> MultiPolygon,
	- [][][]float64 -> Polygon, MultiLineString
	- [][]float64 -> Linestring
There are also Geometry.Type == []*Geometry; Handle uniquely...
*/
func ReduceGeometry(geom *geojson.Geometry) ([][]float64, error) {

	switch polygonType := geom.Type; polygonType {

	case "MultiPolygon": // Send each Polygon of the MultiPolygon...
		var multiPolygonOrder = make([][]float64, len(geom.MultiPolygon))

		for i, polygon := range geom.MultiPolygon {
			pq := priorityQueueFromPolygon(polygon)
			multiPolygonOrder[i] = pq.getQueuePriorityOrder()
		}
		return multiPolygonOrder, nil

	case "Polygon": // Send Polygon to run
		var polygonOrder = make([][]float64, 1)
		pq := priorityQueueFromPolygon(geom.Polygon)

		polygonOrder[0] = pq.getQueuePriorityOrder()
		return polygonOrder, nil

	case "MultiLineString": // Send MultiLineString to run
		var multiLineStringOrder = make([][]float64, len(geom.MultiLineString))
		var geomLinestring [][][]float64

		for i, linestring := range geom.MultiLineString {
			geomLinestring = [][][]float64{linestring}
			pq := priorityQueueFromPolygon(geomLinestring)
			multiLineStringOrder[i] = pq.getQueuePriorityOrder()
		}
		return multiLineStringOrder, nil

	case "LineString": // Send LineString to run
		var linestringOrder = make([][]float64, 1)

		geomLinestring := [][][]float64{geom.LineString}
		pq := priorityQueueFromPolygon(geomLinestring)

		linestringOrder[0] = pq.getQueuePriorityOrder()
		return linestringOrder, nil

	// Nested GeometryCollection - Ugh
	case "GeometryCollection":
		var geomCollectionOrder = make([][]float64, len(geom.Geometries))

		for i, geom := range geom.Geometries {
			result, err := ReduceGeometry(geom)
			if err != nil {
				// TODO: Warn here...
				result = [][]float64{}
				log.Fatal(1)
			}
			geomCollectionOrder[i] = result[0] // TODO: Check this...
		}

		return geomCollectionOrder, nil

	// Do nothing; Type Not Yet Supported...
	default:
		return [][]float64{}, nil
	}

}

// BatchReduceGEOJSON - Read's a reducer's (shape) data
// and kick of reducing jobs.
func BatchReduceGEOJSON(b []byte) (*geojson.FeatureCollection, error) {

	var r Reducer
	var wg sync.WaitGroup

	// NOTE: BIG Assumption Here -> ioutil.ReadAll puts everything
	// Into memory, assumes files we read won't be too large.
	fc1, err := geojson.UnmarshalFeatureCollection(b)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize Reducer
	r = Reducer{
		Data: fc1.Features,
	}

	// Calculate polygon priority
	for idx := range r.Data {
		wg.Add(1)
		go r.reduceFeaturesAsyncWrapper(idx, &wg)
	}

	wg.Wait()

	return fc1, nil
}

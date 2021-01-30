// Package viswal -
package viswal

import (
	"math"
)

/*
Point - Atomic Unit for Viswal Algorithm
- X, Y: coords of the Point
- leftPoint, rightPoint: pointers to neighboring points
- alive: boolean indicating if the node is still active
- currentArea: The currentArea the priority of the point in the queue
	under Viswal-Whyatt. `currentArea` represents the area of the triangle
	(p-1, p, p+1)
- index: Location in the queue, maintained by the heap.Interface{} methods
*/
type Point struct {
	id                    int
	X, Y                  float64
	leftPoint, rightPoint *Point
	alive                 bool
	currentArea           float64
	index                 int
}

// newPoint - Create a new `viswsal.Point` from id/index
// and coordinates. Initializes with infinite area until
// `leftNode` and `rightNode` assigned
func newPoint(idx int, X float64, Y float64) *Point {

	return &Point{
		id:          idx,
		X:           X,
		Y:           Y,
		currentArea: math.Inf(1),
		index:       idx,
		alive:       true,
	}
}

// calcAreas - Calculates the area of the triangle formed by
// a point (p2) and the neighboring 2 points (p1, p3)
func (p2 *Point) area(p1 *Point, p3 *Point) {
	p2.currentArea = math.Abs((p1.X*p2.Y)+(p2.X*p3.Y)+(p3.X*p1.Y)-(p1.X*p3.Y)-(p2.X*p1.Y)-(p3.X*p2.Y)) / 2
}

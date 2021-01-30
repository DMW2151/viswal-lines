// Package viswal -
package viswal

import (
	"container/heap"
)

// heap.Inerface{} methods from container/heap docs, see:
// 	- https://golang.org/pkg/container/heap/

// PriorityQueue - implements heap.Interface and holds Points
type PriorityQueue []*Point

/*
Functions required to Implement the "container/heap" Heap Interface
*/

// Swap -
func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index, pq[j].index = i, j
}

// Less -
func (pq PriorityQueue) Less(i, j int) bool {
	// Pop gives us the lowest, not highest, priority. We use less than here.
	if i < pq.Len() && j < pq.Len() {
		return pq[i].currentArea < pq[j].currentArea
	}
	return false
}

// Len -
func (pq PriorityQueue) Len() int {
	return len(pq)
}

// Push -
func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	point := x.(*Point)
	point.index = n
	*pq = append(*pq, point)
}

// Pop -
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	point := old[n-1]
	old[n-1] = nil // Avoid memory leak
	*pq = old[0 : n-1]
	return point
}

// priorityQueueFromPolygon - Generates a New `PriorityQueue` from a list of coordinates
func priorityQueueFromPolygon(polygon [][][]float64) *PriorityQueue {

	// Initialize with same length as input polygon
	var pq = make(PriorityQueue, len(polygon[0]))

	// For each point in the Polygon insert into the PQ
	for i, p := range polygon[0] {
		pq[i] = newPoint(i, p[0], p[1])
	}

	// Set pointers in each point's `leftPoint` and `rightPoint` field
	for i := range pq {
		if i > 0 {
			pq[i].leftPoint = pq[i-1]
		}
		if i < pq.Len()-1 {
			pq[i].rightPoint = pq[i+1]
		}
	}

	return &pq
}

func (pq *PriorityQueue) getPointArea(point *Point) {
	if point.leftPoint != nil && point.rightPoint != nil {
		point.area(point.leftPoint, point.rightPoint)
	}
}

// getQueuePriorityOrder - calculates the least significant
// remaining pops a point, pops from the heap, and assigns a
// value on [0, 1] for that point. Continues while pq.Len() > 2
func (pq *PriorityQueue) getQueuePriorityOrder() []float64 {

	var countPoints = pq.Len()
	var priorityOrder = make([]float64, countPoints)
	var point interface{}

	// Assign the Area of All Current Points
	for _, p := range *pq {
		pq.getPointArea(p)
		heap.Fix(pq, p.index)
	}

	// Reduce the geometry to it's start & end point
	// Assign a normalize value to priority order
	for droppedCtr := countPoints; droppedCtr > 2; droppedCtr-- {
		point = heap.Pop(pq)
		priorityOrder[point.(*Point).id] = (float64(droppedCtr) / float64(countPoints))

		// Update adjacent triangles
		pq.update(point.(*Point).leftPoint)
		pq.update(point.(*Point).rightPoint)
	}

	return priorityOrder
}

// Update modifies the priority and value of an Point in the queue.
func (pq *PriorityQueue) update(point *Point) {

	var leftNode = point.leftPoint
	var rightNode = point.rightPoint

	// Get New Left and Right Nodes
	for (leftNode != nil) && (rightNode != nil) {
		for !leftNode.alive { // While left node is dead - continue to left
			leftNode = leftNode.leftPoint
		}
		for !rightNode.alive { // While right node is dead - continue to right
			rightNode = rightNode.rightPoint
		}

		// Recalc Areas && Reset Nodes
		point.area(leftNode, rightNode)
		leftNode.rightPoint, rightNode.leftPoint = rightNode, leftNode

		// call to heap.Fix - implementation from heap/container
		heap.Fix(pq, point.index)
		return
	}

}

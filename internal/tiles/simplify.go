package tiles

import "math"

const simplifyTolerance = 0.000015

func simplifyLine(points [][]float64) [][]float64 {
	if len(points) < 3 {
		return points
	}

	keep := make([]bool, len(points))
	keep[0] = true
	keep[len(points)-1] = true

	douglasPeucker(points, 0, len(points)-1, simplifyTolerance, keep)

	simplified := make([][]float64, 0, len(points))
	for i, k := range keep {
		if k {
			simplified = append(simplified, points[i])
		}
	}
	return simplified
}

func douglasPeucker(points [][]float64, start, end int, tolerance float64, keep []bool) {
	if end-start < 2 {
		return
	}

	maxDist := 0.0
	maxIdx := -1

	for i := start + 1; i < end; i++ {
		dist := perpendicularDistance(points[i], points[start], points[end])
		if dist > maxDist {
			maxDist = dist
			maxIdx = i
		}
	}

	if maxDist > tolerance && maxIdx != -1 {
		keep[maxIdx] = true
		douglasPeucker(points, start, maxIdx, tolerance, keep)
		douglasPeucker(points, maxIdx, end, tolerance, keep)
	}
}

func perpendicularDistance(p, lineStart, lineEnd []float64) float64 {
	x, y := p[0], p[1]
	x1, y1 := lineStart[0], lineStart[1]
	x2, y2 := lineEnd[0], lineEnd[1]

	dx := x2 - x1
	dy := y2 - y1

	if dx == 0 && dy == 0 {
		// lineStart and lineEnd are the same point.
		return math.Hypot(x-x1, y-y1)
	}

	// Project p onto the line, clamped to the segment.
	t := ((x-x1)*dx + (y-y1)*dy) / (dx*dx + dy*dy)
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}

	closestX := x1 + t*dx
	closestY := y1 + t*dy

	return math.Hypot(x-closestX, y-closestY)
}
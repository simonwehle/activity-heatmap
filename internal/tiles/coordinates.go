package tiles

import "math"

func roundCoord(v float64) float64 {
	const precision = 1e6 // 6 decimal places
	return math.Round(v*precision) / precision
}

func roundCoordPair(pair []float64) []float64 {
	return []float64{roundCoord(pair[0]), roundCoord(pair[1])}
}
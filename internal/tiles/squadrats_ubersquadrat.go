package tiles

// UbersquadratResult holds the computed Übersquadrat: the largest
// n×n block where every tile is visited. Connectivity to a yard is
// not required — any visited tiles count.
type UbersquadratResult struct {
	Size    int
	TopLeft TileXY
	BotRight TileXY
}

// findUbersquadrat finds the largest n×n square of visited tiles using
// the classic "maximal square in a binary matrix" DP algorithm.
// dp[i][j] = side length of the largest all-visited square whose
// bottom-right corner is at grid position (i, j).
func findUbersquadrat(visited map[TileXY]bool) *UbersquadratResult {
	if len(visited) == 0 {
		return nil
	}

	minX, minY := 0, 0
	maxX, maxY := 0, 0
	first := true
	for t := range visited {
		if first {
			minX, maxX = t.X, t.X
			minY, maxY = t.Y, t.Y
			first = false
		} else {
			if t.X < minX { minX = t.X }
			if t.X > maxX { maxX = t.X }
			if t.Y < minY { minY = t.Y }
			if t.Y > maxY { maxY = t.Y }
		}
	}

	width  := maxX - minX + 1
	height := maxY - minY + 1

	// Allocate DP table as a flat slice for cache efficiency.
	dp := make([]int, width*height)
	idx := func(i, j int) int { return i*height + j }

	bestSize := 0
	bestBR := TileXY{}

	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			tx := minX + i
			ty := minY + j
			if !visited[TileXY{X: tx, Y: ty}] {
				dp[idx(i, j)] = 0
				continue
			}
			if i == 0 || j == 0 {
				dp[idx(i, j)] = 1
			} else {
				a := dp[idx(i-1, j)]
				b := dp[idx(i, j-1)]
				c := dp[idx(i-1, j-1)]
				m := a
				if b < m { m = b }
				if c < m { m = c }
				dp[idx(i, j)] = m + 1
			}
			if dp[idx(i, j)] > bestSize {
				bestSize = dp[idx(i, j)]
				bestBR = TileXY{X: tx, Y: ty}
			}
		}
	}

	if bestSize == 0 {
		return nil
	}

	return &UbersquadratResult{
		Size:     bestSize,
		TopLeft:  TileXY{X: bestBR.X - bestSize + 1, Y: bestBR.Y - bestSize + 1},
		BotRight: bestBR,
	}
}

// ubersquadratPolygon converts an UbersquadratResult into the single
// closed polygon ring that outlines the entire square block — used as
// a GeoJSON Polygon outline layer (fill-opacity: 0, only the border
// is visible).
func ubersquadratPolygon(u *UbersquadratResult, zoom int) [][]float64 {
	// Northwest corner of top-left tile
	nwLat, nwLon := tileToLatLon(u.TopLeft.X, u.TopLeft.Y, zoom)
	// Southeast corner of bottom-right tile
	seLat, seLon := tileToLatLon(u.BotRight.X+1, u.BotRight.Y+1, zoom)

	return [][]float64{
		{nwLon, nwLat},
		{seLon, nwLat},
		{seLon, seLat},
		{nwLon, seLat},
		{nwLon, nwLat},
	}
}

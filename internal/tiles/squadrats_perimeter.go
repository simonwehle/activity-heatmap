package tiles

// borderEdge represents one edge of the outer perimeter as two
// tile-corner coordinates (in tile-grid integer space, not lat/lon).
// Converted to lat/lon when building the GeoJSON.
type borderEdge struct {
	X1, Y1, X2, Y2 int
}

// findPerimeterEdges returns all outer border edges of the visited tile
// set — an edge exists where a visited tile borders a non-visited tile.
// Internal edges (between two visited tiles) are excluded, so the
// result is purely the outer boundary.
func findPerimeterEdges(visited map[TileXY]bool) []borderEdge {
	var edges []borderEdge
	for t := range visited {
		tx, ty := t.X, t.Y
		// North: neighbour (tx, ty-1) not visited
		if !visited[TileXY{tx, ty - 1}] {
			edges = append(edges, borderEdge{tx, ty, tx + 1, ty})
		}
		// South: neighbour (tx, ty+1) not visited
		if !visited[TileXY{tx, ty + 1}] {
			edges = append(edges, borderEdge{tx, ty + 1, tx + 1, ty + 1})
		}
		// West: neighbour (tx-1, ty) not visited
		if !visited[TileXY{tx - 1, ty}] {
			edges = append(edges, borderEdge{tx, ty, tx, ty + 1})
		}
		// East: neighbour (tx+1, ty) not visited
		if !visited[TileXY{tx + 1, ty}] {
			edges = append(edges, borderEdge{tx + 1, ty, tx + 1, ty + 1})
		}
	}
	return edges
}

// perimeterToGeoJSON converts border edges into a GeoJSON
// FeatureCollection of LineString features (one per edge), suitable
// for a "line" layer with controllable line-width and color.
// Each edge is a short line segment in lat/lon space.
func perimeterToGeoJSON(edges []borderEdge, zoom int) map[string]interface{} {
	features := make([]interface{}, 0, len(edges))
	for _, e := range edges {
		lat1, lon1 := tileToLatLon(e.X1, e.Y1, zoom)
		lat2, lon2 := tileToLatLon(e.X2, e.Y2, zoom)
		features = append(features, map[string]interface{}{
			"type": "Feature",
			"geometry": map[string]interface{}{
				"type":        "LineString",
				"coordinates": [][]float64{{lon1, lat1}, {lon2, lat2}},
			},
			"properties": map[string]interface{}{},
		})
	}
	return map[string]interface{}{
		"type":     "FeatureCollection",
		"features": features,
	}
}

package tiles

// PolygonGeometry is a GeoJSON Polygon geometry. Coordinates is an
// array of linear rings — the first is the exterior ring, any
// further rings would be holes (unused here, since tile squares have
// no holes). This is intentionally a separate type from
// GeoJSONGeometry (used for LineStrings elsewhere in this package),
// because Polygon coordinates are one nesting level deeper
// ([][][]float64, not [][]float64) — reusing the LineString struct
// here would produce invalid GeoJSON.
type PolygonGeometry struct {
	Type        string        `json:"type"`
	Coordinates [][][]float64 `json:"coordinates"`
}

// PolygonFeature is a GeoJSON Feature wrapping a PolygonGeometry.
type PolygonFeature struct {
	Type       string                 `json:"type"`
	Geometry   PolygonGeometry        `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
}

// PolygonFeatureCollection is a GeoJSON FeatureCollection of
// PolygonFeatures — the squadrats tile grid's source data shape.
type PolygonFeatureCollection struct {
	Type     string           `json:"type"`
	Features []PolygonFeature `json:"features"`
}

// collectVisitedTiles walks every coordinate in the given segments and
// returns the set of unique tiles ([x,y] at squadratsZoom) that any
// point falls into. Using a map deduplicates automatically — a tile
// crossed by 50 different runs still counts once.
func collectVisitedTiles(segments []LineSegment) map[TileXY]bool {
	visited := make(map[TileXY]bool)

	for _, seg := range segments {
		for _, coord := range seg.Coordinates {
			lon, lat := coord[0], coord[1]
			tile := latLonToTile(lat, lon, squadratsZoom)
			visited[tile] = true
		}
	}

	return visited
}

// visitedTilesToGeoJSON converts a set of visited tiles into a
// FeatureCollection of square Polygon features, one per tile.
func visitedTilesToGeoJSON(visited map[TileXY]bool) PolygonFeatureCollection {
	fc := PolygonFeatureCollection{
		Type:     "FeatureCollection",
		Features: []PolygonFeature{},
	}

	for tile := range visited {
		fc.Features = append(fc.Features, PolygonFeature{
			Type: "Feature",
			Geometry: PolygonGeometry{
				Type:        "Polygon",
				Coordinates: [][][]float64{tileBoundsPolygon(tile, squadratsZoom)},
			},
			Properties: map[string]interface{}{
				"tile_x": tile.X,
				"tile_y": tile.Y,
			},
		})
	}

	return fc
}

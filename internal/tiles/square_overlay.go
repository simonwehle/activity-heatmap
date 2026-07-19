package tiles

import (
	"encoding/json"
	"net/http"
)

// cachedSquadratsGeoJSON holds the computed tile-grid FeatureCollection,
// built once whenever Generate() runs (same trigger as the heatmap
// cache), guarded by the same cacheMutex already used elsewhere in
// this package.
var cachedSquadratsGeoJSON *PolygonFeatureCollection

// buildSquadratsCache computes the visited-tile grid from the given
// segments and stores it in cachedSquadratsGeoJSON. Call this from
// Generate(), right after cachedSegments is set, while still holding
// cacheMutex — see integration note below.
func buildSquadratsCache(segments []LineSegment) {
	visited := collectVisitedTiles(segments)
	fc := visitedTilesToGeoJSON(visited)
	cachedSquadratsGeoJSON = &fc
}

// SquadratsLayer is a minimal {id, type, paint} layer entry, defined
// independently here rather than reusing a type from heatmap_overlay.go,
// since that file's current version doesn't define a named layer
// struct (it reads layers as []interface{} straight from JSON).
type SquadratsLayer struct {
	ID    string                 `json:"id"`
	Type  string                 `json:"type"`
	Paint map[string]interface{} `json:"paint"`
}

// SquadratsOverlayResponse mirrors the same {source, layers} shape as
// the heatmap overlay response, so the Flutter/web client-side code
// can reuse the same addOverlay()-style logic for this new route —
// just pointed at a different URL, with a "fill" layer instead of
// "line".
type SquadratsOverlayResponse struct {
	Source interface{}      `json:"source"`
	Layers []SquadratsLayer `json:"layers"`
}

// squadratsOverlayLayers defines the single fill layer used to render
// visited tiles. Kept minimal for the first version: one flat color,
// no clustering/yard distinction yet.
var squadratsOverlayLayers = []SquadratsLayer{
	{
		ID:   "squadrats-tiles",
		Type: "fill",
		Paint: map[string]interface{}{
			"fill-color":         "#ff4d00",
			"fill-opacity":       0.35,
			"fill-outline-color": "#ff4d00",
		},
	},
}

// GetSquadratsOverlay serves the visited-tile grid as its own route,
// separate from /api/heatmap-overlay — both are available
// simultaneously, neither replaces the other.
func GetSquadratsOverlay(w http.ResponseWriter, r *http.Request) {
	cacheMutex.Lock()
	geojson := cachedSquadratsGeoJSON
	cacheMutex.Unlock()

	if geojson == nil {
		http.Error(w, "Squadrats data not ready", http.StatusServiceUnavailable)
		return
	}

	response := SquadratsOverlayResponse{
		Source: geojson,
		Layers: squadratsOverlayLayers,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

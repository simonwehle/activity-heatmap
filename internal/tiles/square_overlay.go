package tiles

import (
	"encoding/json"
	"net/http"
)

var cachedSquadratsGeoJSON *PolygonFeatureCollection
var cachedYardGeoJSON      *PolygonFeatureCollection
var cachedUbersquadrat     *UbersquadratResult

// buildSquadratsCache computes visited tiles, yard (fully enclosed
// interior tiles), and Übersquadrat (largest n×n filled square).
// Called from Generate() while holding cacheMutex.
func buildSquadratsCache(segments []LineSegment) {
	visited := collectVisitedTiles(segments)
	yard    := findYard(visited)
	uber    := findUbersquadrat(visited)

	allFC  := visitedTilesToGeoJSON(visited)
	yardFC := visitedTilesToGeoJSON(yard)

	cachedSquadratsGeoJSON = &allFC
	cachedYardGeoJSON      = &yardFC
	cachedUbersquadrat     = uber
}

// GetSquadratsOverlay serves the standalone squadrats overlay
// (kept for debugging; /api/heathunt is the active combined route).
func GetSquadratsOverlay(w http.ResponseWriter, r *http.Request) {
	cacheMutex.Lock()
	all  := cachedSquadratsGeoJSON
	yard := cachedYardGeoJSON
	uber := cachedUbersquadrat
	cacheMutex.Unlock()

	if all == nil || yard == nil {
		http.Error(w, "Squadrats data not ready", http.StatusServiceUnavailable)
		return
	}

	sources, layers := buildSquadratsSourcesAndLayers(all, yard, uber)

	type response struct {
		Sources map[string]interface{} `json:"sources"`
		Layers  []interface{}          `json:"layers"`
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response{Sources: sources, Layers: layers})
}

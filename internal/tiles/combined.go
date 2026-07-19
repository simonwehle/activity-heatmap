package tiles

import (
	"encoding/json"
	"net/http"
)

type CombinedOverlayResponse struct {
	Sources map[string]interface{} `json:"sources"`
	Layers  []interface{}          `json:"layers"`
}

func GetCombinedOverlay(w http.ResponseWriter, r *http.Request) {
	cacheMutex.Lock()
	heatmapGeojson := cachedGeoJSON
	squadratsGeojson := cachedSquadratsGeoJSON
	cacheMutex.Unlock()

	if heatmapGeojson == nil || squadratsGeojson == nil {
		http.Error(w, "Overlay data not ready", http.StatusServiceUnavailable)
		return
	}

	trackLayers, err := loadStyleLayers()
	if err != nil {
		http.Error(w, "failed to load style layers", http.StatusInternalServerError)
		return
	}

	// squadratsOverlayLayers doesn't carry a "source" field (it was
	// designed for the single-source overlay shape), so add one here
	// per layer for the combined, multi-source shape.
	squadratsLayers := make([]interface{}, len(squadratsOverlayLayers))
	for i, l := range squadratsOverlayLayers {
		squadratsLayers[i] = map[string]interface{}{
			"id":     l.ID,
			"type":   l.Type,
			"source": "squadrats-source",
			"paint":  l.Paint,
		}
	}

	allLayers := make([]interface{}, 0, len(trackLayers)+len(squadratsLayers))
	allLayers = append(allLayers, trackLayers...)
	allLayers = append(allLayers, squadratsLayers...)

	response := CombinedOverlayResponse{
		Sources: map[string]interface{}{
			"tracks-source": map[string]interface{}{
				"type": "geojson",
				"data": heatmapGeojson,
			},
			"squadrats-source": map[string]interface{}{
				"type": "geojson",
				"data": squadratsGeojson,
			},
		},
		Layers: allLayers,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
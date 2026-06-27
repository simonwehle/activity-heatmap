package tiles

import (
	"encoding/json"
	"net/http"
	"os"
)

type HeatmapOverlayResponse struct {
	Source interface{}   `json:"source"`
	Layers []interface{} `json:"layers"`
}

var styleLayers []interface{}

func loadStyleLayers() ([]interface{}, error) {
	if styleLayers != nil {
		return styleLayers, nil
	}

	data, err := os.ReadFile("./style/heatmap.json")
	if err != nil {
		return nil, err
	}

	var style struct {
		Layers []interface{} `json:"layers"`
	}
	if err := json.Unmarshal(data, &style); err != nil {
		return nil, err
	}

	styleLayers = style.Layers
	return styleLayers, nil
}

func GetHeatmapOverlay(w http.ResponseWriter, r *http.Request) {
	cacheMutex.Lock()
	geojson := cachedGeoJSON
	cacheMutex.Unlock()

	if geojson == nil {
		http.Error(w, "Heatmap not ready", http.StatusServiceUnavailable)
		return
	}

	layers, err := loadStyleLayers()
	if err != nil {
		http.Error(w, "failed to load style layers", http.StatusInternalServerError)
		return
	}

	response := HeatmapOverlayResponse{
		Source: geojson,
		Layers: layers,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
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

func loadHeathuntStyleLayers() ([]interface{}, error) {
	layers, err := loadStyleLayers()
	if err != nil {
		return nil, err
	}

	pink := map[string][]interface{}{
		"tracks-glow": {
			"interpolate",
			[]interface{}{"linear"},
			[]interface{}{"get", "count"},
			1, "rgb(244, 75, 254)",
			10, "rgb(247, 120, 255)",
			50, "rgb(249, 170, 255)",
			100, "rgb(253, 217, 255)",
		},
		"tracks-core": {
			"interpolate",
			[]interface{}{"linear"},
			[]interface{}{"get", "count"},
			1, "rgb(244, 75, 254)",
			10, "rgb(239, 110, 250)",
			50, "rgb(245, 150, 254)",
			100, "rgb(247, 190, 255)",
		},
	}

	for _, layer := range layers {
		m := layer.(map[string]interface{})
		if color, ok := pink[m["id"].(string)]; ok {
			m["paint"].(map[string]interface{})["line-color"] = color
		}
	}
	return layers, nil
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
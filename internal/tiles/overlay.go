package tiles

import (
	"encoding/json"
	"net/http"
)

type HeatmapOverlayResponse struct {
	Source interface{}           `json:"source"`
	Layers []HeatmapOverlayLayer `json:"layers"`
}

type HeatmapOverlayLayer struct {
	ID    string                 `json:"id"`
	Type  string                 `json:"type"`
	Paint map[string]interface{} `json:"paint"`
}

func GetHeatmapOverlay(w http.ResponseWriter, r *http.Request) {
	cacheMutex.Lock()
	geojson := cachedGeoJSON
	cacheMutex.Unlock()

	if geojson == nil {
		http.Error(w, "Heatmap not ready", http.StatusServiceUnavailable)
		return
	}

	response := HeatmapOverlayResponse{
		Source: geojson,
		Layers: []HeatmapOverlayLayer{
			{
				ID:   "tracks-glow",
				Type: "line",
				Paint: map[string]interface{}{
					"line-color": []interface{}{
						"interpolate", []interface{}{"linear"}, []interface{}{"get", "count"},
						1, "rgb(255, 77, 0)",
						10, "rgb(255, 130, 60)",
						50, "rgb(255, 180, 130)",
						100, "rgb(255, 220, 190)",
					},
					"line-width": []interface{}{
						"interpolate", []interface{}{"linear"}, []interface{}{"zoom"},
						8, []interface{}{"interpolate", []interface{}{"linear"}, []interface{}{"get", "count"}, 1, 1.5, 100, 4},
						12, []interface{}{"interpolate", []interface{}{"linear"}, []interface{}{"get", "count"}, 1, 2.5, 100, 6},
						16, []interface{}{"interpolate", []interface{}{"linear"}, []interface{}{"get", "count"}, 1, 3.5, 100, 8},
					},
					"line-opacity": []interface{}{
						"interpolate", []interface{}{"linear"}, []interface{}{"get", "count"},
						1, 0.35,
						20, 0.5,
						100, 0.6,
					},
					"line-blur": []interface{}{
						"interpolate", []interface{}{"linear"}, []interface{}{"zoom"},
						8, 1.4,
						12, 1.0,
						18, 0.8,
					},
				},
			},
			{
				ID:   "tracks-core",
				Type: "line",
				Paint: map[string]interface{}{
					"line-color": []interface{}{
						"interpolate", []interface{}{"linear"}, []interface{}{"get", "count"},
						1, "rgb(255, 77, 0)",
						10, "rgb(255, 110, 40)",
						50, "rgb(255, 150, 90)",
						100, "rgb(255, 195, 150)",
					},
					"line-width": []interface{}{
						"interpolate", []interface{}{"linear"}, []interface{}{"zoom"},
						8, []interface{}{"interpolate", []interface{}{"linear"}, []interface{}{"get", "count"}, 1, 0.6, 100, 2.2},
						12, []interface{}{"interpolate", []interface{}{"linear"}, []interface{}{"get", "count"}, 1, 0.9, 100, 2.8},
						16, []interface{}{"interpolate", []interface{}{"linear"}, []interface{}{"get", "count"}, 1, 1.2, 100, 3.4},
					},
					"line-opacity": []interface{}{
						"interpolate", []interface{}{"linear"}, []interface{}{"get", "count"},
						1, 0.65,
						20, 0.8,
						100, 0.95,
					},
					"line-blur": 0,
				},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

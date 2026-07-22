package tiles

import (
	"encoding/json"
	"net/http"
)

type CombinedOverlayResponse struct {
	Sources map[string]interface{} `json:"sources"`
	Layers  []interface{}          `json:"layers"`
}

// buildSquadratsSourcesAndLayers constructs sources and layers for the
// squadrats visualization:
//   - All visited tiles:  #FCD7AE fill
//   - Yard tiles:         #FFFFFF fill (fully enclosed interior only)
//   - Outer perimeter:    black line outline around visited tile cluster
//   - Übersquadrat:       #FF4D00 thick line outline
func buildSquadratsSourcesAndLayers(
	all       *PolygonFeatureCollection,
	yard      *PolygonFeatureCollection,
	uber      *UbersquadratResult,
	perimeter map[string]interface{},
) (map[string]interface{}, []interface{}) {

	sources := map[string]interface{}{
		"squadrats-all":       map[string]interface{}{"type": "geojson", "data": all},
		"squadrats-yard":      map[string]interface{}{"type": "geojson", "data": yard},
		"squadrats-perimeter": map[string]interface{}{"type": "geojson", "data": perimeter},
	}

	layers := []interface{}{
		// 1) All visited tiles — peach base
		map[string]interface{}{
			"id":     "squadrats-all-fill",
			"type":   "fill",
			"source": "squadrats-all",
			"paint": map[string]interface{}{
				"fill-color":   "#FCD7AE",
				"fill-opacity": 0.6,
			},
		},
		// 2) Yard tiles — white interior
		map[string]interface{}{
			"id":     "squadrats-yard-fill",
			"type":   "fill",
			"source": "squadrats-yard",
			"paint": map[string]interface{}{
				"fill-color":   "#FFFFFF",
				"fill-opacity": 0.7,
			},
		},
		// 3) Outer perimeter — black outline around entire visited cluster
		map[string]interface{}{
			"id":     "squadrats-perimeter-line",
			"type":   "line",
			"source": "squadrats-perimeter",
			"paint": map[string]interface{}{
				"line-color":   "#000000",
				"line-width":   1.5,
				"line-opacity": 0.8,
			},
			"layout": map[string]interface{}{
				"line-join": "round",
				"line-cap":  "square",
			},
		},
	}

	// 4) Übersquadrat — thick orange outline
	if uber != nil {
		ring := ubersquadratPolygon(uber, squadratsZoom)
		uberGeoJSON := map[string]interface{}{
			"type": "FeatureCollection",
			"features": []interface{}{
				map[string]interface{}{
					"type": "Feature",
					"geometry": map[string]interface{}{
						"type":        "LineString",
						"coordinates": ring,
					},
					"properties": map[string]interface{}{
						"size": uber.Size,
					},
				},
			},
		}
		sources["squadrats-uber"] = map[string]interface{}{
			"type": "geojson",
			"data": uberGeoJSON,
		}
		layers = append(layers, map[string]interface{}{
			"id":     "squadrats-uber-outline",
			"type":   "line",
			"source": "squadrats-uber",
			"paint": map[string]interface{}{
				"line-color":   "#FF4D00",
				"line-width":   4,
				"line-opacity": 1.0,
			},
			"layout": map[string]interface{}{
				"line-join": "round",
				"line-cap":  "round",
			},
		})
	}

	return sources, layers
}

func GetCombinedOverlay(w http.ResponseWriter, r *http.Request) {
	cacheMutex.Lock()
	heatmapGeojson := cachedGeoJSON
	all            := cachedSquadratsGeoJSON
	yard           := cachedYardGeoJSON
	uber           := cachedUbersquadrat
	perimeter      := cachedPerimeterGeoJSON
	cacheMutex.Unlock()

	if heatmapGeojson == nil || all == nil || yard == nil {
		http.Error(w, "Overlay data not ready", http.StatusServiceUnavailable)
		return
	}

	trackLayers, err := loadStyleLayers()
	if err != nil {
		http.Error(w, "failed to load style layers", http.StatusInternalServerError)
		return
	}

	squadratsSources, squadratsLayers := buildSquadratsSourcesAndLayers(all, yard, uber, perimeter)

	allSources := map[string]interface{}{
		"tracks-source": map[string]interface{}{
			"type": "geojson",
			"data": heatmapGeojson,
		},
	}
	for k, v := range squadratsSources {
		allSources[k] = v
	}

	allLayers := make([]interface{}, 0, len(squadratsLayers)+len(trackLayers))
	allLayers = append(allLayers, squadratsLayers...)
	allLayers = append(allLayers, trackLayers...)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CombinedOverlayResponse{
		Sources: allSources,
		Layers:  allLayers,
	})
}

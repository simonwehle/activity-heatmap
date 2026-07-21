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
// three-level squadrats visualization:
//   - All visited tiles:  #FCD7AE fill (squadrats color)
//   - Yard tiles:         #FFFFFF fill on top (fully enclosed tiles only)
//   - Übersquadrat:       #FF4D00 thick line outline (no fill)
func buildSquadratsSourcesAndLayers(
	all  *PolygonFeatureCollection,
	yard *PolygonFeatureCollection,
	uber *UbersquadratResult,
) (map[string]interface{}, []interface{}) {

	sources := map[string]interface{}{
		"squadrats-all":  map[string]interface{}{"type": "geojson", "data": all},
		"squadrats-yard": map[string]interface{}{"type": "geojson", "data": yard},
	}

	layers := []interface{}{
		// 1) All visited tiles — Squadrats peach base color
		map[string]interface{}{
			"id":     "squadrats-all-fill",
			"type":   "fill",
			"source": "squadrats-all",
			"paint": map[string]interface{}{
				"fill-color":   "#FCD7AE",
				"fill-opacity": 0.6,
			},
		},
		// 2) Yard tiles — white, on top of peach
		map[string]interface{}{
			"id":     "squadrats-yard-fill",
			"type":   "fill",
			"source": "squadrats-yard",
			"paint": map[string]interface{}{
				"fill-color":   "#FFFFFF",
				"fill-opacity": 0.7,
			},
		},
	}

	// 3) Übersquadrat — thick orange outline as a line layer.
	// Using a LineString (closed ring) instead of fill-outline-color
	// so line-width is fully controllable.
	if uber != nil {
		ring := ubersquadratPolygon(uber, squadratsZoom)
		// Close the ring as a LineString (same coords, just without
		// the Polygon wrapper) so we can use a line layer.
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
		layers = append(layers,
			map[string]interface{}{
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
			},
		)
	}

	return sources, layers
}

func GetCombinedOverlay(w http.ResponseWriter, r *http.Request) {
	cacheMutex.Lock()
	heatmapGeojson := cachedGeoJSON
	all  := cachedSquadratsGeoJSON
	yard := cachedYardGeoJSON
	uber := cachedUbersquadrat
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

	squadratsSources, squadratsLayers := buildSquadratsSourcesAndLayers(all, yard, uber)

	allSources := map[string]interface{}{
		"tracks-source": map[string]interface{}{
			"type": "geojson",
			"data": heatmapGeojson,
		},
	}
	for k, v := range squadratsSources {
		allSources[k] = v
	}

	// Layer order: squadrats fills → übersquadrat outline → track lines on top
	allLayers := make([]interface{}, 0, len(squadratsLayers)+len(trackLayers))
	allLayers = append(allLayers, squadratsLayers...)
	allLayers = append(allLayers, trackLayers...)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CombinedOverlayResponse{
		Sources: allSources,
		Layers:  allLayers,
	})
}

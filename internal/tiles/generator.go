package tiles

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"activity-heatmap/internal/parser"
)

type GeoJSONFeature struct {
	Type       string                 `json:"type"`
	Geometry   GeoJSONGeometry        `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
}

type GeoJSONGeometry struct {
	Type        string      `json:"type"`
	Coordinates [][]float64 `json:"coordinates"`
}

type FeatureCollection struct {
	Type     string          `json:"type"`
	Features []GeoJSONFeature `json:"features"`
}

type LineSegment struct {
	Coordinates [][]float64
	Count       int
}

var (
	cachedGeoJSON *FeatureCollection
	cachedSegments []LineSegment
	cacheMutex    sync.Mutex
)

func Generate() error {
	log.Println("Starting heatmap generation...")

	lineSegments, err := extractLineSegments()
	if err != nil {
		return err
	}
	log.Printf("Extracted %d line segments", len(lineSegments))

	geojson := segmentsToGeoJSON(lineSegments)
	log.Printf("Created GeoJSON with %d line features", len(geojson.Features))

	cacheMutex.Lock()
	cachedSegments = lineSegments
	cachedGeoJSON = &geojson
	cacheMutex.Unlock()

	log.Println("✓ Heatmap generation complete")
	return nil
}

func extractLineSegments() ([]LineSegment, error) {
	var allSegments []LineSegment

	files, err := os.ReadDir("./activities")
	if err != nil {
		return nil, err
	}

	processedFiles := 0
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		name := strings.ToLower(file.Name())
		filePath := filepath.Join("./activities", file.Name())

		var segments []parser.Segment
		var err error

		switch {
		case strings.HasSuffix(name, ".gpx"):
			gpx, e := parser.ParseGPXFile(filePath)
			if e != nil {
				log.Printf("Warning: skipping %s: %v", file.Name(), e)
				continue
			}
			for _, track := range gpx.Tracks {
				for _, seg := range track.Segments {
					if len(seg.Points) < 2 {
						continue
					}
					var pts []parser.Point
					for _, p := range seg.Points {
						pts = append(pts, parser.Point{Lat: p.Lat, Lon: p.Lon})
					}
					segments = append(segments, parser.Segment{Points: pts})
				}
			}
		case strings.HasSuffix(name, ".fit"):
			segments, err = parser.ParseFITFile(filePath)
			if err != nil {
				log.Printf("Warning: skipping %s: %v", file.Name(), err)
				continue
			}
		default:
			continue
		}

		for _, seg := range segments {
			if len(seg.Points) < 2 {
				continue
			}
			coords := make([][]float64, len(seg.Points))
			for i, p := range seg.Points {
				coords[i] = []float64{p.Lon, p.Lat}
			}
			allSegments = append(allSegments, LineSegment{
				Coordinates: coords,
				Count:       1,
			})
		}
		processedFiles++
	}

	log.Printf("Processed %d GPX files", processedFiles)
	return allSegments, nil
}

func segmentsToGeoJSON(segments []LineSegment) FeatureCollection {
	fc := FeatureCollection{
		Type:     "FeatureCollection",
		Features: []GeoJSONFeature{},
	}

	for i, segment := range segments {
		feature := GeoJSONFeature{
			Type: "Feature",
			Geometry: GeoJSONGeometry{
				Type:        "LineString",
				Coordinates: segment.Coordinates,
			},
			Properties: map[string]interface{}{
				"count": segment.Count,
				"id":    i,
			},
		}
		fc.Features = append(fc.Features, feature)
	}

	return fc
}

func GetHeatmapGeoJSON(w http.ResponseWriter, r *http.Request) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	if cachedGeoJSON == nil {
		http.Error(w, "Heatmap not ready", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cachedGeoJSON)
}

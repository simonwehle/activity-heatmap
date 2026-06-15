package tiles

import (
	"bytes"
	"fmt"
	"image/png"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/fogleman/gg"
)

var (
	tileCache      = map[string][]byte{}
	tileCacheMutex sync.Mutex
)

func ClearTileCache() {
	tileCacheMutex.Lock()
	tileCache = map[string][]byte{}
	tileCacheMutex.Unlock()
}

func tileBounds(z, x, y int) (minLon, minLat, maxLon, maxLat float64) {
	n := math.Pow(2, float64(z))
	minLon = float64(x)/n*360 - 180
	maxLon = float64(x+1)/n*360 - 180
	maxLat = math.Atan(math.Sinh(math.Pi*(1-2*float64(y)/n))) * 180 / math.Pi
	minLat = math.Atan(math.Sinh(math.Pi*(1-2*float64(y+1)/n))) * 180 / math.Pi
	return
}

func lonLatToPixel(lon, lat, minLon, minLat, maxLon, maxLat float64, tileSize int) (px, py float64) {
	px = (lon-minLon)/(maxLon-minLon) * float64(tileSize)
	py = (maxLat-lat)/(maxLat-minLat) * float64(tileSize)
	return
}

func ServePNGTile(w http.ResponseWriter, r *http.Request) {
	// path: /tiles/{z}/{x}/{y}.png
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		http.Error(w, "bad path", http.StatusBadRequest)
		return
	}
	z, err1 := strconv.Atoi(parts[1])
	x, err2 := strconv.Atoi(parts[2])
	yStr := strings.TrimSuffix(parts[3], ".png")
	y, err3 := strconv.Atoi(yStr)
	if err1 != nil || err2 != nil || err3 != nil {
		http.Error(w, "bad coords", http.StatusBadRequest)
		return
	}

	key := fmt.Sprintf("%d/%d/%d", z, x, y)

	tileCacheMutex.Lock()
	if cached, ok := tileCache[key]; ok {
		tileCacheMutex.Unlock()
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(cached)
		return
	}
	tileCacheMutex.Unlock()

	const tileSize = 256

	minLon, minLat, maxLon, maxLat := tileBounds(z, x, y)
	dc := gg.NewContext(tileSize, tileSize)
	dc.SetRGBA(0, 0, 0, 0)
	dc.Clear()

	cacheMutex.Lock()
	segments := make([]LineSegment, len(cachedSegments))
	copy(segments, cachedSegments)
	cacheMutex.Unlock()

	for _, seg := range segments {
		if len(seg.Coordinates) < 2 {
			continue
		}

		segMinLon, segMinLat, segMaxLon, segMaxLat := segmentBounds(seg)
		if segMaxLon < minLon || segMinLon > maxLon || segMaxLat < minLat || segMinLat > maxLat {
			continue
		}

		width := math.Min(1.5+float64(seg.Count)*0.3, 6.0)

		// glow layer
		dc.SetRGBA(0.11, 0.56, 1.0, math.Min(0.06+float64(seg.Count)*0.015, 0.35))
		dc.SetLineWidth(width * 2.5)
		dc.SetLineCapRound()
		drawSegment(dc, seg, minLon, minLat, maxLon, maxLat, tileSize)
		dc.Stroke()

		// core layer
		dc.SetRGBA(0.118, 0.565, 1.0, 1.0)
		dc.SetLineWidth(width * 0.6)
		dc.SetLineCapRound()
		drawSegment(dc, seg, minLon, minLat, maxLon, maxLat, tileSize)
		dc.Stroke()
	}

	var buf bytes.Buffer
	png.Encode(&buf, dc.Image())
	pngBytes := buf.Bytes()

	tileCacheMutex.Lock()
	tileCache[key] = pngBytes
	tileCacheMutex.Unlock()

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(pngBytes)
}

func drawSegment(dc *gg.Context, seg LineSegment, minLon, minLat, maxLon, maxLat float64, tileSize int) {
	started := false
	var lastPx, lastPy float64
	for _, c := range seg.Coordinates {
		px, py := lonLatToPixel(c[0], c[1], minLon, minLat, maxLon, maxLat, tileSize)
		if !started {
			dc.MoveTo(px, py)
			started = true
			lastPx, lastPy = px, py
			continue
		}

		dx, dy := px-lastPx, py-lastPy
		if math.Sqrt(dx*dx+dy*dy) < 1.0 {
			continue
		}
		dc.LineTo(px, py)
		lastPx, lastPy = px, py
	}
}

func segmentBounds(seg LineSegment) (minLon, minLat, maxLon, maxLat float64) {
	minLon = seg.Coordinates[0][0]
	maxLon = seg.Coordinates[0][0]
	minLat = seg.Coordinates[0][1]
	maxLat = seg.Coordinates[0][1]

	for _, c := range seg.Coordinates[1:] {
		if c[0] < minLon {
			minLon = c[0]
		}
		if c[0] > maxLon {
			maxLon = c[0]
		}
		if c[1] < minLat {
			minLat = c[1]
		}
		if c[1] > maxLat {
			maxLat = c[1]
		}
	}
	return
}
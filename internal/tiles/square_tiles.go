package tiles

import "math"

// squadratsZoom is the tile zoom level used for the explorer-tile grid.
// Zoom 14 is the de facto standard used by Squadrats, StatsHunters, and
// VeloViewer — each tile is roughly 1-3km across depending on latitude
// (size varies with latitude due to Web Mercator projection; e.g. ~1.5km
// near Berlin's latitude). This matches Squadrats' own description of
// a "squadrat" as approximately 1 square mile.
const squadratsZoom = 14

// TileXY is a single Slippy Map tile coordinate.
type TileXY struct {
	X int
	Y int
}

// latLonToTile converts a latitude/longitude pair into the Slippy Map
// tile (x, y) that contains it, at the given zoom level. Standard
// Web Mercator tile formula, used by OpenStreetMap, MapLibre, and
// every other XYZ tile scheme.
func latLonToTile(lat, lon float64, zoom int) TileXY {
	latRad := lat * math.Pi / 180.0
	n := math.Pow(2, float64(zoom))

	x := int((lon + 180.0) / 360.0 * n)
	y := int((1.0 - math.Log(math.Tan(latRad)+1.0/math.Cos(latRad))/math.Pi) / 2.0 * n)

	return TileXY{X: x, Y: y}
}

// tileToLatLon converts a Slippy Map tile (x, y) at the given zoom
// back into the lat/lon of its top-left (northwest) corner. Used to
// compute the four corners of a tile's square footprint.
func tileToLatLon(x, y, zoom int) (lat, lon float64) {
	n := math.Pow(2, float64(zoom))

	lon = float64(x)/n*360.0 - 180.0
	latRad := math.Atan(math.Sinh(math.Pi * (1.0 - 2.0*float64(y)/n)))
	lat = latRad * 180.0 / math.Pi

	return lat, lon
}

// tileBoundsPolygon returns the four corners of a tile's square
// footprint as a closed ring of [lon, lat] pairs, suitable for a
// GeoJSON Polygon's coordinates[0] (exterior ring), ordered
// northwest -> northeast -> southeast -> southwest -> northwest
// (closing the ring by repeating the first point, as GeoJSON requires).
func tileBoundsPolygon(t TileXY, zoom int) [][]float64 {
	nwLat, nwLon := tileToLatLon(t.X, t.Y, zoom)
	seLat, seLon := tileToLatLon(t.X+1, t.Y+1, zoom)

	return [][]float64{
		{nwLon, nwLat}, // northwest
		{seLon, nwLat}, // northeast
		{seLon, seLat}, // southeast
		{nwLon, seLat}, // southwest
		{nwLon, nwLat}, // close the ring
	}
}

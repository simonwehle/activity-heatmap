package tiles

// findYard returns the subset of visited tiles where all 4 cardinal
// neighbours (N, E, S, W) are also visited — i.e. the fully enclosed
// interior tiles. A tile on the edge of the visited area is never
// part of the yard, even if it's connected to many other tiles.
func findYard(visited map[TileXY]bool) map[TileXY]bool {
	neighbours := [4][2]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}}

	yard := make(map[TileXY]bool)
	for tile := range visited {
		enclosed := true
		for _, d := range neighbours {
			nb := TileXY{X: tile.X + d[0], Y: tile.Y + d[1]}
			if !visited[nb] {
				enclosed = false
				break
			}
		}
		if enclosed {
			yard[tile] = true
		}
	}
	return yard
}

package strava

import (
	"encoding/binary"
	"math"
)

// LineStringWKB encodes coords as a WKB LineString (ISO, little-endian).
// coords are [lon, lat] pairs (X, Y order per GeoParquet spec).
// Returns nil if fewer than 2 points.
func LineStringWKB(coords [][2]float64) []byte {
	if len(coords) < 2 {
		return nil
	}
	buf := make([]byte, 9+len(coords)*16)
	buf[0] = 1                                 // byte order: little-endian
	binary.LittleEndian.PutUint32(buf[1:5], 2) // geometry type: LineString
	binary.LittleEndian.PutUint32(buf[5:9], uint32(len(coords)))
	for i, c := range coords {
		off := 9 + i*16
		binary.LittleEndian.PutUint64(buf[off:off+8], math.Float64bits(c[0]))    // X = lon
		binary.LittleEndian.PutUint64(buf[off+8:off+16], math.Float64bits(c[1])) // Y = lat
	}
	return buf
}

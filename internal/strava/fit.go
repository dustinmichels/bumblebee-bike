package strava

import (
	"compress/gzip"
	"fmt"
	"io"

	"github.com/tormoder/fit"
)

// ReadFITGZ decompresses a gzip stream then parses the FIT data inside.
// Returns [lon, lat] pairs (X, Y) for all valid GPS records.
func ReadFITGZ(r io.Reader) ([][2]float64, error) {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("gzip: %w", err)
	}
	defer gr.Close()

	return readFIT(gr)
}

func readFIT(r io.Reader) ([][2]float64, error) {
	f, err := fit.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	activity, err := f.Activity()
	if err != nil {
		return nil, fmt.Errorf("activity: %w", err)
	}

	coords := make([][2]float64, 0, len(activity.Records))
	for _, rec := range activity.Records {
		if rec.PositionLat.Invalid() || rec.PositionLong.Invalid() {
			continue
		}
		coords = append(coords, [2]float64{
			rec.PositionLong.Degrees(), // X = longitude
			rec.PositionLat.Degrees(),  // Y = latitude
		})
	}
	return coords, nil
}

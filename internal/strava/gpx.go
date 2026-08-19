package strava

import (
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
)

// ReadGPX parses a GPX file and returns [lon, lat] pairs (X, Y) for all
// track points across all tracks and segments.
func ReadGPX(r io.Reader) ([][2]float64, error) {
	dec := xml.NewDecoder(r)
	var coords [][2]float64

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("gpx decode: %w", err)
		}

		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "trkpt" {
			continue
		}

		var lat, lon float64
		var hasLat, hasLon bool
		for _, attr := range se.Attr {
			switch attr.Name.Local {
			case "lat":
				if v, err := strconv.ParseFloat(attr.Value, 64); err == nil {
					lat = v
					hasLat = true
				}
			case "lon":
				if v, err := strconv.ParseFloat(attr.Value, 64); err == nil {
					lon = v
					hasLon = true
				}
			}
		}
		if hasLat && hasLon {
			coords = append(coords, [2]float64{lon, lat}) // X=lon, Y=lat
		}
	}

	return coords, nil
}

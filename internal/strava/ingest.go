package strava

import (
	"archive/zip"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/parquet-go/parquet-go"
)

// geoMetadata is the GeoParquet file-level metadata value for the "geo" key.
const geoMetadata = `{"version":"1.1.0","primary_column":"geometry","columns":{"geometry":{"encoding":"WKB","geometry_types":["LineString"]}}}`

// IngestResult holds counts of activities processed during ingest.
type IngestResult struct {
	Total     int
	Parsed    int
	RideCount int
}

// IngestZip reads a Strava bulk-export zip, processes all supported GPS
// activity formats (.fit.gz, .gpx), and writes a geoparquet file to outPath.
// The zip is read in-place; no temporary extraction is performed.
func IngestZip(zipPath, outPath string) (IngestResult, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return IngestResult{}, fmt.Errorf("open zip %q: %w", zipPath, err)
	}
	defer r.Close()

	// Build a name→file map for O(1) lookup.
	fileMap := make(map[string]*zip.File, len(r.File))
	for _, f := range r.File {
		fileMap[f.Name] = f
	}

	// Locate activities.csv (may be under a single top-level directory).
	var csvEntry *zip.File
	for _, f := range r.File {
		if filepath.Base(f.Name) == "activities.csv" {
			csvEntry = f
			break
		}
	}
	if csvEntry == nil {
		return IngestResult{}, fmt.Errorf("activities.csv not found in %s", zipPath)
	}

	// The prefix is the directory path inside the zip, e.g. "strava_export/".
	prefix := strings.TrimSuffix(csvEntry.Name, "activities.csv")

	rc, err := csvEntry.Open()
	if err != nil {
		return IngestResult{}, fmt.Errorf("open activities.csv in zip: %w", err)
	}
	activities, err := ParseActivities(rc)
	rc.Close()
	if err != nil {
		return IngestResult{}, fmt.Errorf("parse activities.csv: %w", err)
	}

	opener := func(filename string) ([][2]float64, error) {
		name := prefix + filename
		zf, ok := fileMap[name]
		if !ok {
			return nil, fmt.Errorf("not found in zip: %s", name)
		}
		rc, err := zf.Open()
		if err != nil {
			return nil, err
		}
		defer rc.Close()
		return routeTrack(filename, rc)
	}

	return processActivities(activities, opener, outPath)
}

// IngestDir reads an already-extracted Strava export directory and writes
// a geoparquet file to outPath. Useful for faster iteration during development.
func IngestDir(dir, outPath string) (IngestResult, error) {
	f, err := os.Open(filepath.Join(dir, "activities.csv"))
	if err != nil {
		return IngestResult{}, fmt.Errorf("open activities.csv: %w", err)
	}
	activities, err := ParseActivities(f)
	f.Close()
	if err != nil {
		return IngestResult{}, fmt.Errorf("parse activities.csv: %w", err)
	}

	opener := func(filename string) ([][2]float64, error) {
		path := filepath.Join(dir, filename)
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		return routeTrack(filename, f)
	}

	return processActivities(activities, opener, outPath)
}

// processActivities is the shared core: filters to supported GPS activity
// formats, parses tracks, and writes one geoparquet row per activity.
func processActivities(
	activities []Activity,
	openTrack func(filename string) ([][2]float64, error),
	outPath string,
) (IngestResult, error) {
	var result IngestResult
	result.Total = len(activities)

	out, err := os.Create(outPath)
	if err != nil {
		return result, fmt.Errorf("create %q: %w", outPath, err)
	}
	defer out.Close()
	writer := parquet.NewGenericWriter[ActivityRow](out,
		parquet.KeyValueMetadata("geo", geoMetadata),
	)

	var written, skipped int
	for _, act := range activities {
		if act.Filename == nil || !isSupportedTrack(*act.Filename) {
			continue
		}

		coords, err := openTrack(*act.Filename)
		if err != nil {
			slog.Warn("skipping: track read error", "id", act.ActivityID, "file", *act.Filename, "err", err)
			skipped++
			continue
		}

		wkb := LineStringWKB(coords)
		if wkb == nil {
			slog.Warn("skipping: no GPS data", "id", act.ActivityID, "file", *act.Filename)
			skipped++
			continue
		}

		row := activityToRow(act, wkb)
		if _, err := writer.Write([]ActivityRow{row}); err != nil {
			// Close writer before returning so the file is not left open.
			writer.Close()
			return result, fmt.Errorf("write row for activity %d: %w", act.ActivityID, err)
		}
		written++
		if row.ActivityType == "Ride" {
			result.RideCount++
		}
	}

	if err := writer.Close(); err != nil {
		return result, fmt.Errorf("close parquet writer: %w", err)
	}

	result.Parsed = written

	slog.Info("geoparquet written",
		"path", outPath,
		"rows", written,
		"skipped", skipped,
	)
	return result, nil
}

// isSupportedTrack reports whether a filename has a supported GPS format.
func isSupportedTrack(filename string) bool {
	return strings.HasSuffix(filename, ".fit.gz") ||
		strings.HasSuffix(filename, ".fit") ||
		strings.HasSuffix(filename, ".gpx.gz") ||
		strings.HasSuffix(filename, ".gpx") ||
		strings.HasSuffix(filename, ".tcx.gz")
}

// routeTrack dispatches to the correct parser based on file extension.
func routeTrack(filename string, r io.Reader) ([][2]float64, error) {
	switch {
	case strings.HasSuffix(filename, ".fit.gz"):
		return ReadFITGZ(r)
	case strings.HasSuffix(filename, ".fit"):
		return ReadFIT(r)
	case strings.HasSuffix(filename, ".gpx.gz"):
		return ReadGPXGZ(r)
	case strings.HasSuffix(filename, ".gpx"):
		return ReadGPX(r)
	case strings.HasSuffix(filename, ".tcx.gz"):
		return ReadTCXGZ(r)
	default:
		return nil, fmt.Errorf("unsupported track format: %s", filename)
	}
}

func activityToRow(act Activity, wkb []byte) ActivityRow {
	row := ActivityRow{
		ActivityID:   act.ActivityID,
		ActivityDate: act.ActivityDate,
		ActivityName: act.ActivityName,
		ActivityType: act.ActivityType,
		Filename:     *act.Filename,
		Geometry:     wkb,
	}
	if act.ElapsedTime != nil {
		row.ElapsedTime = *act.ElapsedTime
	}
	if act.Distance != nil {
		row.Distance = *act.Distance
	}
	if act.MovingTime != nil {
		row.MovingTime = *act.MovingTime
	}
	if act.MaxSpeed != nil {
		row.MaxSpeed = *act.MaxSpeed
	}
	if act.AverageSpeed != nil {
		row.AverageSpeed = *act.AverageSpeed
	}
	if act.ElevationGain != nil {
		row.ElevationGain = *act.ElevationGain
	}
	if act.ElevationLoss != nil {
		row.ElevationLoss = *act.ElevationLoss
	}
	if act.AverageHeartRate != nil {
		row.AverageHeartRate = *act.AverageHeartRate
	}
	return row
}

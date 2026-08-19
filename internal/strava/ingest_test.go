package strava_test

import (
	"os"
	"testing"

	"github.com/dustinmichels/bumblebee-bike/internal/strava"
	"github.com/parquet-go/parquet-go"
)

const (
	zipPath    = "../../data/strava_export.zip"
	exportDir  = "../../data/strava_export"
	outputPath = "../../data/activities.parquet"
)

// TestIngestZip processes the full Strava export zip and produces a geoparquet
// file at data/activities.parquet.  Run with -v to see row/skip counts.
func TestIngestZip(t *testing.T) {
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		t.Skipf("zip not found: %s", zipPath)
	}
	if err := strava.IngestZip(zipPath, outputPath); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("output file missing after ingest: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("output file is empty")
	}
	t.Logf("wrote %s (%d bytes)", outputPath, info.Size())
}

// TestIngestDir uses the already-extracted directory for faster iteration.
func TestIngestDir(t *testing.T) {
	if _, err := os.Stat(exportDir); os.IsNotExist(err) {
		t.Skipf("export dir not found: %s", exportDir)
	}
	if err := strava.IngestDir(exportDir, outputPath); err != nil {
		t.Fatal(err)
	}

	f, err := os.Open(outputPath)
	if err != nil {
		t.Fatalf("output file missing after ingest: %v", err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	pf, err := parquet.OpenFile(f, info.Size())
	if err != nil {
		t.Fatalf("open parquet: %v", err)
	}

	// 944 .fit.gz with GPS + 208 .gpx = 1152 total supported;
	// 26 have no GPS data → expect 1126 rows.
	const wantRows = 1126
	if got := pf.NumRows(); got != wantRows {
		t.Errorf("row count: got %d, want %d", got, wantRows)
	}
	t.Logf("wrote %s (%d bytes, %d rows)", outputPath, info.Size(), pf.NumRows())
}

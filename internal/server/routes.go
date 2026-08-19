package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"

	"github.com/dustinmichels/bumblebee-bike/internal/strava"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func apiRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/health", handleHealth)
	r.Post("/upload", handleUpload)
	r.Post("/filter", handleFilter)
	return r
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	slog.Info("received upload request")

	// Check if we want to use the default parquet file (only allowed in test mode)
	if r.URL.Query().Get("useDefaultParquet") == "true" || r.FormValue("useDefaultParquet") == "true" {
		if !isTestMode() {
			slog.Error("default parquet requested but test mode is not enabled")
			http.Error(w, "test mode not enabled", http.StatusBadRequest)
			return
		}
		if !fileExists("data/activities.parquet") {
			slog.Error("default activities.parquet not found")
			http.Error(w, "default activities.parquet not found", http.StatusNotFound)
			return
		}

		if err := os.MkdirAll("tmp", 0755); err != nil {
			slog.Error("failed to create tmp directory", "err", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		sessionId := uuid.New().String()
		parquetPath := fmt.Sprintf("tmp/activities-%s.parquet", sessionId)

		slog.Info("copying default activities.parquet to tmp", "sessionId", sessionId, "parquetPath", parquetPath)
		if err := copyFile("data/activities.parquet", parquetPath); err != nil {
			slog.Error("failed to copy activities.parquet", "err", err)
			http.Error(w, "failed to copy activities.parquet", http.StatusInternalServerError)
			return
		}

		totalRows, rideCount, err := getParquetStats(parquetPath)
		if err != nil {
			slog.Warn("failed to query parquet stats, using fallbacks", "err", err)
			totalRows = 0
			rideCount = 0
		}

		summary := fmt.Sprintf("Using default activities.parquet. Loaded %d activities. %d are type = Ride.", totalRows, rideCount)
		slog.Info(summary, "sessionId", sessionId)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":    "success",
			"sessionId": sessionId,
			"total":     totalRows,
			"parsed":    totalRows,
			"rideCount": rideCount,
			"summary":   summary,
		})
		return
	}

	// Check if we want to use the default zip file (only allowed in test mode)
	if r.URL.Query().Get("useDefaultZip") == "true" || r.FormValue("useDefaultZip") == "true" {
		if !isTestMode() {
			slog.Error("default zip requested but test mode is not enabled")
			http.Error(w, "test mode not enabled", http.StatusBadRequest)
			return
		}
		if !fileExists("data/strava_export.zip") {
			slog.Error("default strava_export.zip not found")
			http.Error(w, "default strava_export.zip not found", http.StatusNotFound)
			return
		}

		if err := os.MkdirAll("tmp", 0755); err != nil {
			slog.Error("failed to create tmp directory", "err", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		sessionId := uuid.New().String()
		parquetPath := fmt.Sprintf("tmp/activities-%s.parquet", sessionId)

		slog.Info("processing default zip in-place", "sessionId", sessionId, "parquetPath", parquetPath)
		res, err := strava.IngestZip("data/strava_export.zip", parquetPath)
		if err != nil {
			slog.Error("ingest default zip failed", "sessionId", sessionId, "err", err)
			http.Error(w, fmt.Sprintf("ingest failed: %v", err), http.StatusInternalServerError)
			return
		}

		summary := fmt.Sprintf("Succesfully parsed %d / %d  activities from default zip. %d are type = Ride.", res.Parsed, res.Total, res.RideCount)
		slog.Info(summary, "sessionId", sessionId, "parquetPath", parquetPath)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":    "success",
			"sessionId": sessionId,
			"total":     res.Total,
			"parsed":    res.Parsed,
			"rideCount": res.RideCount,
			"summary":   summary,
		})
		return
	}

	// Limit upload size to 500MB
	if err := r.ParseMultipartForm(500 << 20); err != nil {
		slog.Error("failed to parse multipart form", "err", err)
		http.Error(w, "failed to parse multipart form", http.StatusBadRequest)
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		slog.Error("failed to get file from form", "err", err)
		http.Error(w, "invalid file key 'file' in form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if err := os.MkdirAll("tmp", 0755); err != nil {
		slog.Error("failed to create tmp directory", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	sessionId := uuid.New().String()
	zipPath := fmt.Sprintf("tmp/upload-%s.zip", sessionId)
	parquetPath := fmt.Sprintf("tmp/activities-%s.parquet", sessionId)

	out, err := os.Create(zipPath)
	if err != nil {
		slog.Error("failed to create zip file", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		slog.Error("failed to write zip file", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	out.Close()

	slog.Info("processing upload zip", "sessionId", sessionId, "zipPath", zipPath)
	res, err := strava.IngestZip(zipPath, parquetPath)
	os.Remove(zipPath)

	if err != nil {
		slog.Error("ingest failed", "sessionId", sessionId, "err", err)
		http.Error(w, fmt.Sprintf("ingest failed: %v", err), http.StatusInternalServerError)
		return
	}

	summary := fmt.Sprintf("Succesfully parsed %d / %d  activities. %d are type = Ride.", res.Parsed, res.Total, res.RideCount)
	slog.Info(summary, "sessionId", sessionId, "parquetPath", parquetPath)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":    "success",
		"sessionId": sessionId,
		"total":     res.Total,
		"parsed":    res.Parsed,
		"rideCount": res.RideCount,
		"summary":   summary,
	})
}

type FilterRequest struct {
	SessionId string     `json:"sessionId"`
	BBox      [4]float64 `json:"bbox"`
}

func handleFilter(w http.ResponseWriter, r *http.Request) {
	slog.Info("received filter request")

	var req FilterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("failed to decode filter request", "err", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if _, err := uuid.Parse(req.SessionId); err != nil {
		slog.Error("invalid sessionId format", "sessionId", req.SessionId, "err", err)
		http.Error(w, "invalid sessionId format", http.StatusBadRequest)
		return
	}

	parquetPath := fmt.Sprintf("tmp/activities-%s.parquet", req.SessionId)
	if _, err := os.Stat(parquetPath); os.IsNotExist(err) {
		slog.Error("session parquet file not found", "sessionId", req.SessionId)
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}

	outputGeoJSON := fmt.Sprintf("tmp/output-%s-%s.geojson", req.SessionId, uuid.New().String()[:8])

	slog.Info("filtering activities with duckdb", "sessionId", req.SessionId, "bbox", req.BBox)

	query := fmt.Sprintf(`INSTALL spatial;
LOAD spatial;
SET geometry_always_xy = true;
COPY (
  SELECT activity_id, activity_date, activity_name, geometry 
  FROM read_parquet('%s') 
  WHERE ST_Intersects(geometry, ST_MakeEnvelope(%f, %f, %f, %f)) 
    AND activity_type = 'Ride'
) TO '%s' WITH (FORMAT 'GDAL', DRIVER 'GeoJSON');`,
		parquetPath,
		req.BBox[0], req.BBox[1], req.BBox[2], req.BBox[3],
		outputGeoJSON,
	)

	cmd := exec.Command("duckdb", "-c", query)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		slog.Error("duckdb filter query failed", "err", err, "stderr", stderr.String())
		http.Error(w, fmt.Sprintf("filter query failed: %v", err), http.StatusInternalServerError)
		return
	}
	defer os.Remove(outputGeoJSON)

	geoJSONData, err := os.ReadFile(outputGeoJSON)
	if err != nil {
		slog.Error("failed to read generated geojson", "err", err)
		http.Error(w, "failed to read output data", http.StatusInternalServerError)
		return
	}

	slog.Info("filtering completed successfully", "sessionId", req.SessionId, "bytes", len(geoJSONData))
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(geoJSONData)
}

type HealthResponse struct {
	Status            string `json:"status"`
	TestMode          bool   `json:"testMode"`
	HasDefaultZip     bool   `json:"hasDefaultZip"`
	HasDefaultParquet bool   `json:"hasDefaultParquet"`
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	testMode := isTestMode()
	resp := HealthResponse{
		Status:            "ok",
		TestMode:          testMode,
		HasDefaultZip:     testMode && fileExists("data/strava_export.zip"),
		HasDefaultParquet: testMode && fileExists("data/activities.parquet"),
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func isTestMode() bool {
	return os.Getenv("APP_ENV") == "test" || os.Getenv("ENV") == "test" || os.Getenv("TEST_MODE") == "true"
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

type ParquetStats struct {
	TotalRows int64 `json:"total_rows"`
	RideCount int64 `json:"ride_count"`
}

func getParquetStats(parquetPath string) (int, int, error) {
	query := fmt.Sprintf("SELECT COUNT(*) as total_rows, CAST(SUM(CASE WHEN activity_type = 'Ride' THEN 1 ELSE 0 END) AS BIGINT) as ride_count FROM read_parquet('%s')", parquetPath)
	cmd := exec.Command("duckdb", "-json", "-c", query)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return 0, 0, fmt.Errorf("duckdb query failed: %w, stderr: %s", err, stderr.String())
	}

	var results []ParquetStats
	if err := json.Unmarshal(stdout.Bytes(), &results); err != nil {
		return 0, 0, fmt.Errorf("failed to unmarshal duckdb output: %w", err)
	}

	if len(results) == 0 {
		return 0, 0, fmt.Errorf("no results returned from duckdb")
	}

	return int(results[0].TotalRows), int(results[0].RideCount), nil
}

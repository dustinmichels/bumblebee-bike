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
	r.Get("/map-test", handleMapTest)
	r.Post("/upload", handleUpload)
	r.Post("/filter", handleFilter)
	return r
}

func handleMapTest(w http.ResponseWriter, r *http.Request) {
	if !isTestMode() {
		http.NotFound(w, r)
		return
	}

	parquetPath := "data/activities.parquet"
	if !fileExists(parquetPath) {
		slog.Error("map test parquet not found")
		http.Error(w, "data/activities.parquet not found", http.StatusNotFound)
		return
	}

	slog.Info("loading map test data", "parquetPath", parquetPath)
	geoJSONData, err := exportGeoJSONFromParquet(parquetPath, "WHERE geometry IS NOT NULL")
	if err != nil {
		slog.Error("map test export failed", "err", err)
		http.Error(w, "failed to export map test data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(geoJSONData)
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	slog.Info("received upload request")

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

	slog.Info("filtering activities with duckdb", "sessionId", req.SessionId, "bbox", req.BBox)

	whereClause := fmt.Sprintf(
		"WHERE ST_Intersects(geometry, ST_MakeEnvelope(%f, %f, %f, %f)) AND activity_type = 'Ride'",
		req.BBox[0], req.BBox[1], req.BBox[2], req.BBox[3],
	)

	geoJSONData, err := exportGeoJSONFromParquet(parquetPath, whereClause)
	if err != nil {
		slog.Error("duckdb filter query failed", "err", err)
		http.Error(w, fmt.Sprintf("filter query failed: %v", err), http.StatusInternalServerError)
		return
	}

	slog.Info("filtering completed successfully", "sessionId", req.SessionId, "bytes", len(geoJSONData))
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(geoJSONData)
}

func exportGeoJSONFromParquet(parquetPath, whereClause string) ([]byte, error) {
	if err := os.MkdirAll("tmp", 0755); err != nil {
		return nil, err
	}

	outputGeoJSON := fmt.Sprintf("tmp/output-%s.geojson", uuid.New().String()[:8])
	query := fmt.Sprintf(`INSTALL spatial;
LOAD spatial;
SET geometry_always_xy = true;
COPY (
  SELECT activity_id, activity_date, activity_name, activity_type, geometry
  FROM read_parquet('%s')
  %s
) TO '%s' WITH (FORMAT 'GDAL', DRIVER 'GeoJSON');`,
		parquetPath,
		whereClause,
		outputGeoJSON,
	)

	cmd := exec.Command("duckdb", "-c", query)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("duckdb query failed: %w, stderr: %s", err, stderr.String())
	}
	defer os.Remove(outputGeoJSON)

	return os.ReadFile(outputGeoJSON)
}

type HealthResponse struct {
	Status string `json:"status"`
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(HealthResponse{Status: "ok"})
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

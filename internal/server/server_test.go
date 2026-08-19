package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"
)

func TestHealth(t *testing.T) {
	router := apiRouter()
	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["status"] != "ok" {
		t.Errorf("expected status 'ok', got %q", resp["status"])
	}
	if len(resp) != 1 {
		t.Errorf("expected health response to contain only status, got %v", resp)
	}
}

func TestMapTestUnavailableOutsideTestMode(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("ENV", "")
	t.Setenv("TEST_MODE", "")

	router := apiRouter()
	req := httptest.NewRequest("GET", "/map-test", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rr.Code)
	}
}

func TestSPARejectsMapTestOutsideTestMode(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("ENV", "")
	t.Setenv("TEST_MODE", "")

	fsys := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("index")},
	}
	handler := spaHandler(fsys)

	req := httptest.NewRequest("GET", "/map-test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rr.Code)
	}
}

func TestSPAAllowsMapTestInTestMode(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("ENV", "")
	t.Setenv("TEST_MODE", "")

	fsys := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("index")},
	}
	handler := spaHandler(fsys)

	req := httptest.NewRequest("GET", "/map-test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
}

func TestListRenameOpenAndDeleteUploads(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "uploads")
	t.Setenv("MAPTOOLS_DATA_DIR", dataDir)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatal(err)
	}

	parquetPath := filepath.Join(dataDir, "activities-existing.parquet")
	if err := os.WriteFile(parquetPath, []byte("parquet"), 0644); err != nil {
		t.Fatal(err)
	}

	total := 12
	parsed := 11
	rideCount := 9
	createdAt := time.Date(2026, time.January, 15, 12, 30, 0, 0, time.UTC)
	if err := writeUploadMetadata(uploadMetadataPath(parquetPath), uploadMetadata{
		DatasetID:   "dataset-123",
		DisplayName: "Existing Upload",
		CreatedAt:   createdAt,
		Total:       &total,
		Parsed:      &parsed,
		RideCount:   &rideCount,
	}); err != nil {
		t.Fatal(err)
	}

	router := apiRouter()

	reqList := httptest.NewRequest("GET", "/uploads", nil)
	rrList := httptest.NewRecorder()
	router.ServeHTTP(rrList, reqList)

	if rrList.Code != http.StatusOK {
		t.Fatalf("expected list status 200, got %d", rrList.Code)
	}

	var listResp listUploadsResponse
	if err := json.NewDecoder(rrList.Body).Decode(&listResp); err != nil {
		t.Fatalf("failed to decode list response: %v", err)
	}

	if len(listResp.Uploads) != 1 {
		t.Fatalf("expected one upload, got %d", len(listResp.Uploads))
	}

	upload := listResp.Uploads[0]
	if upload.DatasetID != "dataset-123" {
		t.Fatalf("expected datasetId dataset-123, got %q", upload.DatasetID)
	}
	if upload.DisplayName != "Existing Upload" {
		t.Fatalf("expected display name Existing Upload, got %q", upload.DisplayName)
	}
	if upload.Total == nil || *upload.Total != total {
		t.Fatalf("expected total %d, got %#v", total, upload.Total)
	}

	renameBody := bytes.NewBufferString(`{"name":"Boston Routes"}`)
	reqRename := httptest.NewRequest("PATCH", "/uploads/dataset-123", renameBody)
	reqRename.Header.Set("Content-Type", "application/json")
	rrRename := httptest.NewRecorder()
	router.ServeHTTP(rrRename, reqRename)

	if rrRename.Code != http.StatusOK {
		t.Fatalf("expected rename status 200, got %d", rrRename.Code)
	}

	var renamed UploadedDataset
	if err := json.NewDecoder(rrRename.Body).Decode(&renamed); err != nil {
		t.Fatalf("failed to decode rename response: %v", err)
	}

	renamedParquetPath := filepath.Join(dataDir, "Boston Routes.parquet")
	if _, err := os.Stat(renamedParquetPath); err != nil {
		t.Fatalf("expected renamed parquet file to exist: %v", err)
	}
	if renamed.FileName != "Boston Routes.parquet" {
		t.Fatalf("expected renamed file name to be Boston Routes.parquet, got %q", renamed.FileName)
	}
	if renamed.DisplayName != "Boston Routes" {
		t.Fatalf("expected renamed display name to be Boston Routes, got %q", renamed.DisplayName)
	}

	openedPath := ""
	restoreOpenUploadPath := openUploadPath
	openUploadPath = func(path string) error {
		openedPath = path
		return nil
	}
	defer func() {
		openUploadPath = restoreOpenUploadPath
	}()

	reqOpen := httptest.NewRequest("POST", "/uploads/dataset-123/open", nil)
	rrOpen := httptest.NewRecorder()
	router.ServeHTTP(rrOpen, reqOpen)

	if rrOpen.Code != http.StatusNoContent {
		t.Fatalf("expected open status 204, got %d", rrOpen.Code)
	}
	if openedPath != renamedParquetPath {
		t.Fatalf("expected open path %q, got %q", renamedParquetPath, openedPath)
	}

	reqDelete := httptest.NewRequest("DELETE", "/uploads/dataset-123", nil)
	rrDelete := httptest.NewRecorder()
	router.ServeHTTP(rrDelete, reqDelete)

	if rrDelete.Code != http.StatusNoContent {
		t.Fatalf("expected delete status 204, got %d", rrDelete.Code)
	}

	if _, err := os.Stat(renamedParquetPath); !os.IsNotExist(err) {
		t.Fatalf("expected parquet file to be deleted, stat err = %v", err)
	}
	if _, err := os.Stat(uploadMetadataPath(renamedParquetPath)); !os.IsNotExist(err) {
		t.Fatalf("expected metadata file to be deleted, stat err = %v", err)
	}
}

func TestUploadAndFilter(t *testing.T) {
	zipPath := "../../data/strava_export.zip"
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		t.Skip("skipping integration test, test zip file not found")
	}

	dataDir := filepath.Join(t.TempDir(), "uploads")
	t.Setenv("MAPTOOLS_DATA_DIR", dataDir)

	router := apiRouter()

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, err := mw.CreateFormFile("file", filepath.Base(zipPath))
	if err != nil {
		t.Fatal(err)
	}

	zipFile, err := os.Open(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer zipFile.Close()

	if _, err := io.Copy(fw, zipFile); err != nil {
		t.Fatal(err)
	}
	mw.Close()

	reqUpload := httptest.NewRequest("POST", "/upload", &buf)
	reqUpload.Header.Set("Content-Type", mw.FormDataContentType())
	rrUpload := httptest.NewRecorder()
	router.ServeHTTP(rrUpload, reqUpload)

	if rrUpload.Code != http.StatusOK {
		body, _ := io.ReadAll(rrUpload.Body)
		t.Fatalf("upload failed with status %d: %s", rrUpload.Code, body)
	}

	var uploadResp map[string]any
	if err := json.NewDecoder(rrUpload.Body).Decode(&uploadResp); err != nil {
		t.Fatal(err)
	}

	sessionId, _ := uploadResp["sessionId"].(string)
	if sessionId == "" {
		t.Fatal("expected sessionId, got empty string")
	}

	dataset, ok := uploadResp["dataset"].(map[string]any)
	if !ok {
		t.Fatalf("expected dataset payload, got %v", uploadResp["dataset"])
	}
	if dataset["datasetId"] != sessionId {
		t.Fatalf("expected datasetId %q, got %v", sessionId, dataset["datasetId"])
	}

	if uploadResp["total"] == nil || uploadResp["parsed"] == nil || uploadResp["rideCount"] == nil || uploadResp["summary"] == nil {
		t.Errorf("missing statistics in upload response: %v", uploadResp)
	} else {
		total := uploadResp["total"].(float64)
		parsed := uploadResp["parsed"].(float64)
		rideCount := uploadResp["rideCount"].(float64)
		summary := uploadResp["summary"].(string)

		if parsed != 1146 {
			t.Errorf("expected 1146 parsed activities, got %.0f", parsed)
		}
		if total <= 0 {
			t.Errorf("expected total > 0, got %.0f", total)
		}
		if rideCount <= 0 {
			t.Errorf("expected rideCount > 0, got %.0f", rideCount)
		}
		expectedSummary := fmt.Sprintf("Succesfully parsed %.0f / %.0f  activities. %.0f are type = Ride.", parsed, total, rideCount)
		if summary != expectedSummary {
			t.Errorf("expected summary %q, got %q", expectedSummary, summary)
		}
	}

	parquetPath := filepath.Join(dataDir, "activities-"+sessionId+".parquet")
	defer os.Remove(parquetPath)
	defer os.Remove(uploadMetadataPath(parquetPath))
	if _, err := os.Stat(parquetPath); os.IsNotExist(err) {
		t.Fatalf("expected parquet file to exist at %s, but it does not", parquetPath)
	}

	filterReqBody := FilterRequest{
		SessionId: sessionId,
		BBox:      [4]float64{-71.1912, 42.2279, -70.9227, 42.3969},
	}
	bodyBuf, err := json.Marshal(filterReqBody)
	if err != nil {
		t.Fatal(err)
	}

	reqFilter := httptest.NewRequest("POST", "/filter", bytes.NewReader(bodyBuf))
	reqFilter.Header.Set("Content-Type", "application/json")
	rrFilter := httptest.NewRecorder()
	router.ServeHTTP(rrFilter, reqFilter)

	if rrFilter.Code != http.StatusOK {
		body, _ := io.ReadAll(rrFilter.Body)
		t.Fatalf("filter failed with status %d: %s", rrFilter.Code, body)
	}

	var geojson map[string]interface{}
	if err := json.NewDecoder(rrFilter.Body).Decode(&geojson); err != nil {
		t.Fatalf("failed to decode geojson output: %v", err)
	}

	if geojson["type"] != "FeatureCollection" {
		t.Errorf("expected type FeatureCollection, got %v", geojson["type"])
	}

	features, ok := geojson["features"].([]interface{})
	if !ok {
		t.Fatal("geojson features is not a slice")
	}

	if len(features) == 0 {
		t.Error("expected at least some rides to be returned, got 0")
	} else {
		t.Logf("found %d rides matching search criteria", len(features))
	}
}

func TestGetURL(t *testing.T) {
	tests := []struct {
		addr     string
		expected string
	}{
		{":8080", "http://localhost:8080"},
		{"localhost:8080", "http://localhost:8080"},
		{"127.0.0.1:8080", "http://127.0.0.1:8080"},
		{"0.0.0.0:8080", "http://localhost:8080"},
		{"[::]:8080", "http://localhost:8080"},
		{":80", "http://localhost:80"},
		{"invalid-addr", "http://localhostinvalid-addr"},
	}

	for _, tt := range tests {
		actual := getURL(tt.addr)
		if actual != tt.expected {
			t.Errorf("getURL(%q) = %q, expected %q", tt.addr, actual, tt.expected)
		}
	}
}

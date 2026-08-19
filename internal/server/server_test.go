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
}

func TestUploadAndFilter(t *testing.T) {
	// Path to zip test data from root
	zipPath := "../../data/strava_export.zip"
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		t.Skip("skipping integration test, test zip file not found")
	}

	router := apiRouter()

	// Prepare multipart form upload
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

	// Perform Upload
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
	// Make sure the parquet file was created in tmp
	parquetPath := filepath.Join("tmp", "activities-"+sessionId+".parquet")
	defer os.Remove(parquetPath) // cleanup after test
	if _, err := os.Stat(parquetPath); os.IsNotExist(err) {
		t.Fatalf("expected parquet file to exist at %s, but it does not", parquetPath)
	}

	// Perform Filter
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

	// Read and parse geojson output
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

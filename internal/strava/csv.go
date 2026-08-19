package strava

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ParseActivities parses activities.csv into Activity slices.
// Duplicate column names: first occurrence wins (matches the user-visible fields).
func ParseActivities(r io.Reader) ([]Activity, error) {
	cr := csv.NewReader(r)
	cr.LazyQuotes = true
	cr.FieldsPerRecord = -1 // tolerate ragged rows

	headers, err := cr.Read()
	if err != nil {
		return nil, fmt.Errorf("read headers: %w", err)
	}

	// First-occurrence index per column name.
	idx := make(map[string]int, len(headers))
	for i, h := range headers {
		h = strings.TrimSpace(h)
		if _, exists := idx[h]; !exists {
			idx[h] = i
		}
	}

	col := func(row []string, name string) string {
		i, ok := idx[name]
		if !ok || i >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[i])
	}

	var activities []Activity
	for {
		row, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read row: %w", err)
		}

		act := Activity{
			ActivityID:          parseInt64(col(row, "Activity ID")),
			ActivityDate:        col(row, "Activity Date"),
			ActivityName:        col(row, "Activity Name"),
			ActivityType:        col(row, "Activity Type"),
			ActivityDescription: col(row, "Activity Description"),
			ElapsedTime:         optInt64(col(row, "Elapsed Time")),
			Distance:            optFloat64(col(row, "Distance")),
			Filename:            optString(col(row, "Filename")),
			MovingTime:          optInt64(col(row, "Moving Time")),
			MaxSpeed:            optFloat64(col(row, "Max Speed")),
			AverageSpeed:        optFloat64(col(row, "Average Speed")),
			ElevationGain:       optFloat64(col(row, "Elevation Gain")),
			ElevationLoss:       optFloat64(col(row, "Elevation Loss")),
			AverageHeartRate:    optString(col(row, "Average Heart Rate")),
			ActivityClass:       optString(col(row, "Type")),
			StartTime:           optString(col(row, "Start Time")),
			CarbonSaved:         optInt64(col(row, "Carbon Saved")),
			Media:               optString(col(row, "Media")),
		}
		activities = append(activities, act)
	}

	return activities, nil
}

func parseInt64(s string) int64 {
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}

func optInt64(s string) *int64 {
	if s == "" {
		return nil
	}
	if v, err := strconv.ParseInt(s, 10, 64); err == nil {
		return &v
	}
	// CSV stores some integers as floats (e.g. "2410.0").
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		v := int64(f)
		return &v
	}
	return nil
}

func optFloat64(s string) *float64 {
	if s == "" {
		return nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return &v
}

func optString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

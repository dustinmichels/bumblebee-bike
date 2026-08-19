# TODO

## Parse files

Create a function to ingest a bulk download of Strava data, eg, the one at `data/strava_export.zip`.

For testing, can directly use `data/strava_export` (already unzipped), but the function should ultimately take a path to a zip file and unzip it into a temporary directory.

It needs to:

- unzip the file
- find the `activities.csv` file
- iterate over the rows and parse them into a struct, eg:

```go
type Activity struct {
	ActivityID                int64    `json:"Activity ID"`
	ActivityDate              string   `json:"Activity Date"`
	ActivityName              string   `json:"Activity Name"`
	ActivityType              string   `json:"Activity Type"`
	ActivityDescription       string   `json:"Activity Description"`
	ElapsedTime               *int64   `json:"Elapsed Time,omitempty"`
	Distance                  *float64 `json:"Distance,omitempty"`
	Filename                  *string  `json:"Filename,omitempty"`
	MovingTime                *int64   `json:"Moving Time,omitempty"`
	MaxSpeed                  *float64 `json:"Max Speed,omitempty"`
	AverageSpeed              *float64 `json:"Average Speed,omitempty"`
	ElevationGain             *int64   `json:"Elevation Gain,omitempty"`
	ElevationLoss             *float64 `json:"Elevation Loss,omitempty"`
	AverageHeartRate          *string  `json:"Average Heart Rate,omitempty"`
	Type                      *string  `json:"Type,omitempty"`
	StartTime                 *string  `json:"Start Time,omitempty"`
	CarbonSaved               *int64   `json:"Carbon Saved,omitempty"`
	Media                     *string  `json:"Media,omitempty"`
}
```

Then it needs to follow filename to find the corresponding file in the activities folder. Format could be:

- [x] .fit.gz
- [x] .gpx
- [ ] .fit
- [ ] .gpx.gz
- [ ] .tcx.gz

Treat this as a TODO list and tackle on format at a time. Skip the other formats.

The goal is to parse all activities with a given format into geoparquet. The geoparquet file should combine the route with the metadata from `Activity`. Save the geoparquet file to `data/activities.parquet`.

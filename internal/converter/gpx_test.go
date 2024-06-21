package converter

import (
	"fmt"
	"testing"
	"time"

	"git.rpjosh.de/RPJosh/workout/internal/models"
	"github.com/google/go-cmp/cmp"
)

const (
	TimeFormat = "2006-01-02T15:04:05Z"
)

// TestParseNotify tests to parse an GPX file generated
// by "Notify for Amazafit" / "Notify for Miband"
func TestParseNotify(t *testing.T) {

	content := `
<?xml version="1.0" encoding="UTF-8"?>
<gpx version="1.1" creator="Notify for Amazfit" xmlns="http://www.topografix.com/GPX/1/1" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" 
xsi:schemaLocation="
http://www.topografix.com/GPX/1/1 
http://www.topografix.com/GPX/1/1/gpx.xsd 
http://www.garmin.com/xmlschemas/GpxExtensions/v3 
http://www.garmin.com/xmlschemas/GpxExtensionsv3.xsd 
http://www.garmin.com/xmlschemas/TrackPointExtension/v1 
http://www.garmin.com/xmlschemas/TrackPointExtensionv1.xsd"
xmlns:gpxtpx="http://www.garmin.com/xmlschemas/TrackPointExtension/v1" xmlns:gpxx="http://www.garmin.com/xmlschemas/GpxExtensions/v3">
<name>Gehen</name>
<trk><name>Gehen</name><number>1</number><trkseg>
<trkpt lat="48.67218666666667" lon="10.872203333333333">
    <ele>120</ele>
    <time>2024-04-05T12:55:31Z</time>
    <extensions>
     <gpxtpx:TrackPointExtension>
      <gpxtpx:hr>115</gpxtpx:hr>
     </gpxtpx:TrackPointExtension>
    </extensions>
   </trkpt>
<trkpt lat="48.67219166666667" lon="10.872224666666666">
    <ele>0</ele>
    <time>2024-04-05T12:55:32Z</time>
   </trkpt>
<trkpt lat="48.671743666666664" lon="10.877556666666667">
    <ele>301</ele>
    <time>2024-04-05T12:57:33Z</time>
    <extensions>
     <gpxtpx:TrackPointExtension>
      <gpxtpx:hr>150</gpxtpx:hr>
     </gpxtpx:TrackPointExtension>
    </extensions>
   </trkpt>
<trkpt lat="48.671728333333334" lon="10.877595">
    <ele>310</ele>
    <time>2024-04-05T12:57:34Z</time>
    <extensions>
     <gpxtpx:TrackPointExtension>
      <gpxtpx:hr>165</gpxtpx:hr>
     </gpxtpx:TrackPointExtension>
    </extensions>
   </trkpt>
   </trkseg></trk>
 </gpx>
	`

	expected := &models.GpxFile{
		Type: models.TYPE_HIKING,
		Points: []models.GpxPoint{
			{
				Lat:       48.67218666666667,
				Lon:       10.872203333333333,
				Timestamp: parseTime("2024-04-05T12:55:31Z"),
				Elevation: 120,
				HeartRate: 115,
			},
			{
				Lat:       48.67219166666667,
				Lon:       10.872224666666666,
				Timestamp: parseTime("2024-04-05T12:55:32Z"),
				Elevation: 0,
				HeartRate: 0,
			},
			{
				Lat:       48.671743666666664,
				Lon:       10.877556666666667,
				Timestamp: parseTime("2024-04-05T12:57:33Z"),
				Elevation: 301,
				HeartRate: 150,
			},
			{
				Lat:       48.671728333333334,
				Lon:       10.877595,
				Timestamp: parseTime("2024-04-05T12:57:34Z"),
				Elevation: 310,
				HeartRate: 165,
			},
		},
	}

	// Parse
	got, err := ParseGPX([]byte(content))
	if err != nil {
		t.Fatalf("Failed to parse GPX file: %s", err)
	}

	// Compare structs
	if diff := cmp.Diff(expected, got); diff != "" {
		t.Errorf("Mismatch of parsed Notify GPX file (-want +got):\n%s", diff)
	}
}

func parseTime(timeStr string) time.Time {
	rtc, err := time.Parse(TimeFormat, timeStr)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse time %q: %s", timeStr, err))
	}

	return rtc
}

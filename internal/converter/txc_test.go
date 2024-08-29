package converter

import (
	"testing"

	"git.rpjosh.de/RPJosh/workout/internal/models"
	"github.com/google/go-cmp/cmp"
)

// TestParseTcxFitbit tests the parsing of a TCX file generated
// by the "Fitbit" app
func TestParseTcxFitbit(t *testing.T) {

	content := `
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<TrainingCenterDatabase xmlns="http://www.garmin.com/xmlschemas/TrainingCenterDatabase/v2">
    <Activities>
        <Activity Sport="Running">
            <Id>2024-08-29T17:34:59.000+02:00</Id>
            <Lap StartTime="2024-08-29T17:34:59.000+02:00">
                <TotalTimeSeconds>1064.0</TotalTimeSeconds>
                <DistanceMeters>304.42</DistanceMeters>
                <Calories>78</Calories>
                <Intensity>Active</Intensity>
                <TriggerMethod>Manual</TriggerMethod>
				<Track>
                    <Trackpoint>
                        <Time>2024-08-29T17:34:59.000+02:00</Time>
                        <Position>
                            <LatitudeDegrees>48.676828265190125</LatitudeDegrees>
                            <LongitudeDegrees>10.848488330841064</LongitudeDegrees>
                        </Position>
                        <AltitudeMeters>417.361200020385</AltitudeMeters>
                        <DistanceMeters>0.0</DistanceMeters>
                        <HeartRateBpm>
                            <Value>82</Value>
                        </HeartRateBpm>
                    </Trackpoint>
                    <Trackpoint>
                        <Time>2024-08-29T17:35:04.000+02:00</Time>
                        <Position>
                            <LatitudeDegrees>48.676828265190125</LatitudeDegrees>
                            <LongitudeDegrees>10.848488330841064</LongitudeDegrees>
                        </Position>
                        <AltitudeMeters>417.33949764805584</AltitudeMeters>
                        <DistanceMeters>0.0</DistanceMeters>
                        <HeartRateBpm>
                            <Value>82</Value>
                        </HeartRateBpm>
                    </Trackpoint>
                    <Trackpoint>
                        <Time>2024-08-29T17:35:09.000+02:00</Time>
                        <Position>
                            <LatitudeDegrees>48.676828265190125</LatitudeDegrees>
                            <LongitudeDegrees>10.848488330841064</LongitudeDegrees>
                        </Position>
                        <AltitudeMeters>429.56779527572667</AltitudeMeters>
                        <DistanceMeters>0.0</DistanceMeters>
                        <HeartRateBpm>
                            <Value>81</Value>
                        </HeartRateBpm>
                    </Trackpoint>
                    <Trackpoint>
                        <Time>2024-08-29T17:35:14.000+02:00</Time>
                        <Position>
                            <LatitudeDegrees>48.676828265190125</LatitudeDegrees>
                            <LongitudeDegrees>10.848488330841064</LongitudeDegrees>
                        </Position>
                        <AltitudeMeters>417.2960929033975</AltitudeMeters>
                        <DistanceMeters>0.0</DistanceMeters>
                        <HeartRateBpm>
                            <Value>92</Value>
                        </HeartRateBpm>
                    </Trackpoint>
                </Track>
            </Lap>
            <Creator xsi:type="Device_t" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
                <UnitId>0</UnitId>
                <ProductID>0</ProductID>
            </Creator>
        </Activity>
    </Activities>
</TrainingCenterDatabase>
	`

	expected := &models.GpxFile{
		Type: models.TYPE_RUNNING,
		Points: []models.GpxPoint{
			{
				Lat:       48.676828265190125,
				Lon:       10.848488330841064,
				Timestamp: parseTime("2024-08-29T15:34:59Z"),
				Elevation: 417,
				HeartRate: 82,
			},
			{
				Lat:       48.676828265190125,
				Lon:       10.848488330841064,
				Timestamp: parseTime("2024-08-29T15:35:04Z"),
				Elevation: 417,
				HeartRate: 82,
			},
			{
				Lat:       48.676828265190125,
				Lon:       10.848488330841064,
				Timestamp: parseTime("2024-08-29T15:35:09Z"),
				Elevation: 430,
				HeartRate: 81,
			},
			{
				Lat:       48.676828265190125,
				Lon:       10.848488330841064,
				Timestamp: parseTime("2024-08-29T15:35:14Z"),
				Elevation: 417,
				HeartRate: 92,
			},
		},
	}

	// Parse
	got, err := ParseTcx([]byte(content))
	if err != nil {
		t.Fatalf("Failed to parse TXC file: %s", err)
	}

	// Compare structs
	if diff := cmp.Diff(expected, got); diff != "" {
		t.Errorf("Mismatch of parsed Fitbit TCX file (-want +got):\n%s", diff)
	}
}

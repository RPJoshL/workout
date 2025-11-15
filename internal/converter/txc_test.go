package converter

import (
	"testing"
	"time"

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
                        <DistanceMeters>30.0</DistanceMeters>
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
                        <DistanceMeters>90.0</DistanceMeters>
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
                        <DistanceMeters>300.0</DistanceMeters>
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
				Distance:  0,
			},
			{
				Lat:       48.676828265190125,
				Lon:       10.848488330841064,
				Timestamp: parseTime("2024-08-29T15:35:04Z"),
				Elevation: 417,
				HeartRate: 82,
				Distance:  30,
			},
			{
				Lat:       48.676828265190125,
				Lon:       10.848488330841064,
				Timestamp: parseTime("2024-08-29T15:35:09Z"),
				Elevation: 430,
				HeartRate: 81,
				Distance:  90,
			},
			{
				Lat:       48.676828265190125,
				Lon:       10.848488330841064,
				Timestamp: parseTime("2024-08-29T15:35:14Z"),
				Elevation: 417,
				HeartRate: 92,
				Distance:  300,
			},
		},
		DeviceData: models.DeviceData{
			UseDeviceData: true,
			PauseDuration: int(defaultPauseThreshold.Seconds()),
		},
	}

	// Parse
	got, err := ParseTcx([]byte(content), 0)
	if err != nil {
		t.Fatalf("Failed to parse TXC file: %s", err)
	}

	// Compare structs
	if diff := cmp.Diff(expected, got); diff != "" {
		t.Errorf("Mismatch of parsed Fitbit TCX file (-want +got):\n%s", diff)
	}
}

func TestRemovePauses(t *testing.T) {
	baseTime := time.Now()

	input := &models.GpxFile{
		Points: []models.GpxPoint{
			{Lat: 48.1, Lon: 48.2, Timestamp: baseTime, HeartRate: 120},
			{Lat: 48.2, Lon: 48.3, Timestamp: baseTime.Add(10 * time.Second), HeartRate: 120},
			{Lat: 48.3, Lon: 48.3, Timestamp: baseTime.Add(40 * time.Second), HeartRate: 122},
			// Start of pause
			{Lat: 48.3, Lon: 48.5, Timestamp: baseTime.Add(80 * time.Second), HeartRate: 122},
			{Lat: 48.3, Lon: 48.3, Timestamp: baseTime.Add(120 * time.Second), HeartRate: 122},
			{Lat: 48.3, Lon: 48.6, Timestamp: baseTime.Add(160 * time.Second), HeartRate: 122},
			{Lat: 48.3, Lon: 48.3, Timestamp: baseTime.Add(200 * time.Second), HeartRate: 122},
			// End of pause
			{Lat: 48.4, Lon: 48.3, Timestamp: baseTime.Add(240 * time.Second), HeartRate: 167},
			{Lat: 48.5, Lon: 48.3, Timestamp: baseTime.Add(280 * time.Second), HeartRate: 167},
		},
	}

	expected := []models.GpxPoint{
		{Lat: 48.1, Lon: 48.2, Timestamp: baseTime, HeartRate: 120},
		{Lat: 48.2, Lon: 48.3, Timestamp: baseTime.Add(10 * time.Second), HeartRate: 120},
		{Lat: 48.3, Lon: 48.3, Timestamp: baseTime.Add(40 * time.Second), HeartRate: 122},
		{Lat: 48.4, Lon: 48.3, Timestamp: baseTime.Add(240 * time.Second), HeartRate: 167},
		{Lat: 48.5, Lon: 48.3, Timestamp: baseTime.Add(280 * time.Second), HeartRate: 167},
	}

	got := removePauses(input)

	// Compare structs
	if diff := cmp.Diff(expected, got.Points); diff != "" {
		t.Errorf("Mismatch of remove pauses from GPX file (-want +got):\n%s", diff)
	}
}

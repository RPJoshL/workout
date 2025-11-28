package converter

import (
	"testing"
	"time"

	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/pkg/assert"
	"github.com/google/go-cmp/cmp"
	"github.com/guregu/null/v5"
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
			PauseDuration: int(noGPSPauseThreshold.Seconds()),
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

func TestToTXC(t *testing.T) {
	input := models.Workout{
		Name:        "Joggen",
		Id:          123,
		TypeId:      models.TYPE_RUNNING,
		Start:       addTime(0),
		Duration:    120,
		Distance:    800,
		HeartRateAv: null.IntFrom(130),
		Calories:    120,
		WorkoutDetails: []models.WorkoutDetails{
			// Point with all data
			{
				Duration:  6,
				HeartRate: null.IntFrom(120),
				Speed:     120,
				Distance:  500,
				Latitude:  48.122,
				Longitude: 11.567,
				Elevation: 432,
				Time:      addTime(6),
			},
			// MIssing position and speed
			{
				Duration:  12,
				HeartRate: null.IntFrom(150),
				Distance:  800,
				Elevation: 400,
				Time:      addTime(12),
			},
		},
	}

	expected := `
	<TrainingCenterDatabase 
		xmlns="http://www.garmin.com/xmlschemas/TrainingCenterDatabase/v2"
		xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
		xsi:schemaLocation="https://www8.garmin.com/xmlschemas/TrainingCenterDatabasev2.xsd https://www8.garmin.com/xmlschemas/ActivityExtensionv2.xsd"
	>
		<Activities>
			<Activity Sport="Joggen">
				<Id>123</Id>
				<Lap StartTime="2025-04-10T06:30:00Z">
					<TotalTimeSeconds>120</TotalTimeSeconds>
					<DistanceMeters>800</DistanceMeters>
					<Calories>120</Calories>
					<Intensity>Active</Intensity>
					<TriggerMethod>Manual</TriggerMethod>
					<AverageHeartRateBpm>130</AverageHeartRateBpm>
					<Track>
						<Trackpoint>
							<Time>2025-04-10T06:30:06Z</Time>
							<Position>
								<LatitudeDegrees>48.122</LatitudeDegrees>
								<LongitudeDegrees>11.567</LongitudeDegrees>
							</Position>
							<AltitudeMeters>432</AltitudeMeters>
							<DistanceMeters>500</DistanceMeters>
							<HeartRateBpm>
								<Value>120</Value>
							</HeartRateBpm>
							<Extensions>
								<TPX xmlns="http://www.garmin.com/xmlschemas/ActivityExtension/v2">
									<Speed>120</Speed>
								</TPX>
							</Extensions>
						</Trackpoint>
						<Trackpoint>
							<Time>2025-04-10T06:30:12Z</Time>
							<AltitudeMeters>400</AltitudeMeters>
							<DistanceMeters>800</DistanceMeters>
							<HeartRateBpm>
								<Value>150</Value>
							</HeartRateBpm>
							<Extensions />
						</Trackpoint>
					</Track>
				</Lap>
			</Activity>
		</Activities>
	</TrainingCenterDatabase>
	`

	got, err := ToTCX(&input)
	assert.NoError(t, err)

	assert.XMLEq(t, expected, string(got))
}

// adds the provided amount of seconds to the base time
func addTime(seconds int) time.Time {
	base := time.Date(2025, time.April, 10, 6, 30, 0, 0, time.UTC)

	return base.Add(time.Duration(seconds) * time.Second)
}

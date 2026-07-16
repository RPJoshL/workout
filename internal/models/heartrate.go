package models

// HeartrateZone contains details for a specific heart
// rate zone
type HeartrateZone struct {

	// Lower bound
	Min int

	// Upper bound
	Max int

	// CSS color definition (for filling)
	Color string

	// A better foreground color for text
	FgColor string
}

func GetHeartrateZones() []HeartrateZone {
	return []HeartrateZone{
		{
			Min:     0,
			Max:     97,
			Color:   "#ff8a80",
			FgColor: "#ff8a80",
		},
		{
			Min:     97,
			Max:     116,
			Color:   "#2862ff",
			FgColor: "#4879fd",
		},
		{
			Min:     116,
			Max:     135,
			Color:   "#00cee9",
			FgColor: "#00cee9",
		},
		{
			Min:     135,
			Max:     154,
			Color:   "#65dd19",
			FgColor: "#65dd19",
		},
		{
			Min:     154,
			Max:     174,
			Color:   "#ff6d01",
			FgColor: "#ff6d01",
		},
		{
			Min:     174,
			Max:     220,
			Color:   "#aa00ff",
			FgColor: "#b638f5",
		},
	}
}

// GetZoneByHeartrate returns the zone
// in which the provided heart rate is
func GetZoneByHeartrate(rate int) HeartrateZone {
	zones := GetHeartrateZones()
	for _, z := range zones {
		if rate >= z.Min && rate < z.Max {
			return z
		}
	}

	return HeartrateZone{
		Color: "#aaa",
	}
}

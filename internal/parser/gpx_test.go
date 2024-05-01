package parser

func AddMetersToBase(horizontal, vertical float64) (float64, float64) {
	baseLat := 48.00
	baseLon := 12.00

	baseLat, baseLon = AddMetersToPosition(baseLat, baseLon, horizontal, true)
	baseLat, baseLon = AddMetersToPosition(baseLat, baseLon, vertical, false)
	return baseLat, baseLon
}

func AddMetersToBaseLat(horizontal, vertical int) float32 {
	lat, _ := AddMetersToBase(float64(horizontal), float64(vertical))
	return float32(lat)
}
func AddMetersToBaseLon(horizontal, vertical int) float32 {
	_, lon := AddMetersToBase(float64(horizontal), float64(vertical))
	return float32(lon)
}

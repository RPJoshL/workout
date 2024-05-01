package converter

import "git.rpjosh.de/RPJosh/workout/internal/models"

// ParseWorkoutFile parses a workout file in a supported format of the application.
// It's always defaulting to a GPX file!
func ParseWorkoutFile(filename string, content []byte) (rtc *models.GpxFile, err error) {

	// GPX is the only supported format currently
	return ParseGPX(content)

}

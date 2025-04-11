package converter

import (
	"strings"

	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
)

// ParseWorkoutFile parses a workout file in a supported format of the application.
// It's always defaulting to a GPX file!
func ParseWorkoutFile(filename string, content []byte) (rtc *models.GpxFile, err errors.Error) {
	filenameU := strings.ToUpper(filename)

	if strings.HasSuffix(filenameU, ".TCX") {
		return ParseTcx(content)
	} else {
		// Use a GPX file by default
		return ParseGPX(content)
	}
}

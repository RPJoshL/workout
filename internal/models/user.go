package models

import (
	"time"

	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/workout/internal/translator"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type KeyUserType int

const (
	KeyUser KeyUserType = iota
	KeyLanguage
)

// WebUser is a wrapper around [User] with additional client details
// from the browser
type WebUser struct {
	*User

	TimeZone *time.Location

	// Whether the user is "priveleged" and is allowed
	// to do security relevant actions
	Priveleged bool

	// Whether the underlaying user has to be updated within the database
	NeedsUpdate bool

	// API key that was used for authentication against the API
	ApiKey ApiKey

	// Language of the browser request
	Language translator.Language
}

type User struct {
	// Unique ID of the user
	Id int `json:"id" dbColumn:"Column:id,AutoIncrement,PrimaryKey"`
	// Display name showed to other users
	Name string `json:"name" dbColumn:"Column:name"`
	// Unique E-Mail address
	Mail string `json:"mail" dbColumn:"Column:mail"`
	// Hashed password with the argon2 algorithm
	Password string `json:"password" dbColumn:"Column:password"`
	// Body weight in kg
	Weight int `json:"weight" dbColumn:"Column:weight"`
	// Body height in cm
	Height int `json:"height" dbColumn:"Column:height"`
	// Year the user was born in
	BirthYear int `json:"birthYear" dbColumn:"Column:birth_year"`
	// VO2max value in mL/kg/min
	Vo2Max int `json:"vo2Max" dbColumn:"Column:vo2_max"`
	// Male (0) or Female (1)
	Gender int `json:"gender" dbColumn:"Column:gender"`
	// Weather the user enabled the dark theme instead of the light one
	DarkTheme int `json:"darkTheme" dbColumn:"Column:dark_theme"`
	// Timezone the user specified in the last request
	Timezone    string `json:"timezone" dbColumn:"Column:timezone,DefaultValue"`
	DbMetadata_ any    `json:"-" dbMetadata:"Schema:workout,Table:user"`
}

// User
const (
	User_Id        string = "Id|workout.user.id"
	User_Name      string = "Name|workout.user.name"
	User_Mail      string = "Mail|workout.user.mail"
	User_Password  string = "Password|workout.user.password"
	User_Weight    string = "Weight|workout.user.weight"
	User_Height    string = "Height|workout.user.height"
	User_BirthYear string = "BirthYear|workout.user.birth_year"
	User_Vo2Max    string = "Vo2Max|workout.user.vo2_max"
	User_Gender    string = "Gender|workout.user.gender"
	User_DarkTheme string = "DarkTheme|workout.user.dark_theme"
	User_Timezone  string = "Timezone|workout.user.timezone"
)

// Properties that are retrieved from [WebUser]
var WebUserProperties = []string{User_Timezone}

// NewWebUser initializes a new WebUser with the provided details.
//
// This function returns wheather the user has to be updated. The new details
// are already correctly updated in the user struct
func NewWebUser(user *User, timeZone string) (*WebUser, bool) {
	rtc := &WebUser{User: user}
	updateUser := rtc.SetClientTimeZone(timeZone)

	return rtc, updateUser
}

// SetClientTimeZone parses and sets a provided time zone by the client.
// It returns whether the underlaying user has to be updated within the database.
func (u *WebUser) SetClientTimeZone(timeZone string) bool {
	updateUser := false

	// Parse timezone
	if timeZone == "" {
		// Fallback to previously saved timezone in DB
		if tz, err := time.LoadLocation(u.User.Timezone); err != nil {
			u.TimeZone = time.UTC
			logger.Warning("Failed to parse timezone of user %q - %q: %s", u.User.Name, u.User.Timezone, err)
		} else {
			u.TimeZone = tz
		}
	} else {
		if tz, err := time.LoadLocation(timeZone); err != nil {
			u.TimeZone = time.UTC
			logger.Warning("Failed to parse timezone of user %q - %q: %s", u.User.Name, timeZone, err)
		} else {
			u.TimeZone = tz
			// Update timezone if it differs from DB value
			if timeZone != u.Timezone {
				updateUser = true
				u.User.Timezone = timeZone
			}
		}
	}

	// Apply update flag
	if updateUser {
		u.NeedsUpdate = true
	}

	return updateUser
}

// ApplyTimezone applies the users timezone offset to
// the servers timezone and returns the modified time
func (u *WebUser) ApplyTimezone(serverTime time.Time) time.Time {
	return serverTime.In(u.TimeZone)
}

// Sprintf formats the provided template string (like [fmt.Sprintf])
// in the users language
func (u *WebUser) Sprintf(text string, arguments ...any) string {
	languageTag := language.English
	switch u.Language {
	case translator.German:
		languageTag = language.German
	}

	return message.NewPrinter(languageTag).Sprintf(text, arguments...)
}

// ToServerTimezone transforms a date in a user timezone to
// the servers timezone and returns the modified time
func (u *WebUser) ToServerTimezone(clientTime time.Time) time.Time {
	clientTimeCorrect := time.Date(clientTime.Year(), clientTime.Month(), clientTime.Day(), clientTime.Hour(), clientTime.Minute(), clientTime.Second(), clientTime.Nanosecond(), u.TimeZone)
	return clientTimeCorrect
}

// GetTimeZoneOffset returns the offset in seconds from the UTC time zone
func (u *WebUser) GetTimeZoneOffset() int {
	t := time.Now().In(u.TimeZone)

	// Get offset to UTC
	_, offset := t.Zone()
	return offset
}

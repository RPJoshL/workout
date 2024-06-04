package models

import (
	"time"

	"git.rpjosh.de/RPJosh/go-logger"
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
	updateUser := false

	// Parse timezone
	if timeZone == "" {
		// Fallback to previously saved timezone in DB
		if tz, err := time.LoadLocation(user.Timezone); err != nil {
			rtc.TimeZone = time.UTC
			logger.Warning("Failed to parse timezone of user %q - %q: %s", user.Name, user.Timezone, err)
		} else {
			rtc.TimeZone = tz
		}
	} else {
		if tz, err := time.LoadLocation(timeZone); err != nil {
			rtc.TimeZone = time.UTC
			logger.Warning("Failed to parse timezone of user %q - %q: %s", user.Name, timeZone, err)
		} else {
			rtc.TimeZone = tz
			// Update timezone if it differs from DB value
			if timeZone != rtc.Timezone {
				updateUser = true
				user.Timezone = timeZone
			}
		}
	}

	return rtc, updateUser
}

// ApplyTimezone applies the users timezone offset to
// the servers timezone and returns the modified time
func (u *WebUser) ApplyTimezone(serverTime time.Time) time.Time {
	return serverTime.In(u.TimeZone)
}

// ToServerTimezone transforms a date in a users timezone to
// the servers timezone and returns the modified time
func (u *WebUser) ToServerTimezone(clientTime time.Time) time.Time {
	clientTimeCorrect := time.Date(clientTime.Year(), clientTime.Month(), clientTime.Day(), clientTime.Hour(), clientTime.Minute(), clientTime.Second(), clientTime.Nanosecond(), u.TimeZone)
	return clientTimeCorrect
}

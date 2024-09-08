package models

import (
	"github.com/guregu/null/v5"
	"time"
)

type ApiKey struct {
	// Unique ID of this token
	Id int `json:"id" dbColumn:"Column:id,AutoIncrement"`
	// Random (and unique) hashed value of the token
	Key string `json:"key" dbColumn:"Column:key,PrimaryKey"`
	// ID of the user to which the token belongs to
	UserId int `json:"userId" dbColumn:"Column:user_id,ForeignKey:workout.user.id"`
	// An obfuscated version of the tokens RAW value (unhashed)
	Obfuscated string `json:"obfuscated" dbColumn:"Column:obfuscated"`
	// Date and time this token was created
	CreationDate time.Time `json:"creationDate" dbColumn:"Column:creation_date,DefaultValue"`
	// Until which date the token is valid
	ValidUntil time.Time `json:"validUntil" dbColumn:"Column:valid_until"`
	// User set alias name of the token to identify this token later
	Alias null.String `json:"alias" dbColumn:"Column:alias,DefaultValue"`
	// Whether the user enabled the dark theme instead of the light one for this token
	DarkTheme   int `json:"darkTheme" dbColumn:"Column:dark_theme,DefaultValue"`
	DbMetadata_ any `json:"-" dbMetadata:"Schema:workout,Table:api_key"`
}

// ApiKey
const (
	ApiKey_Id           string = "Id|workout.api_key.id"
	ApiKey_Key          string = "Key|workout.api_key.key"
	ApiKey_UserId       string = "UserId|workout.api_key.user_id"
	ApiKey_Obfuscated   string = "Obfuscated|workout.api_key.obfuscated"
	ApiKey_CreationDate string = "CreationDate|workout.api_key.creation_date"
	ApiKey_ValidUntil   string = "ValidUntil|workout.api_key.valid_until"
	ApiKey_Alias        string = "Alias|workout.api_key.alias"
	ApiKey_DarkTheme    string = "DarkTheme|workout.api_key.dark_theme"
)

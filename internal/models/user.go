package models

type KeyUserType int

const (
	KeyUser KeyUserType = iota
	KeyLanguage
)

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
	// Weather the user enabled the dark theme instead of the light one
	DarkTheme   int `json:"darkTheme" dbColumn:"Column:dark_theme"`
	DbMetadata_ any `json:"-" dbMetadata:"Schema:workout,Table:user"`
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
	User_DarkTheme string = "DarkTheme|workout.user.dark_theme"
)

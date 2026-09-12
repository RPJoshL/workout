package models

type ExternalAPIType string

const (
	ExternalAPIStrave      ExternalAPIType = "strava"
	ExternalAPIPumpfoilOrg ExternalAPIType = "pumpfoilorg"
)

var ExternalAPIFromString = map[string]ExternalAPIType{
	"strava":      ExternalAPIStrave,
	"pumpfoilorg": ExternalAPIPumpfoilOrg,
}

type ExternalApi struct {
	Id int `json:"id" dbColumn:"Column:id,AutoIncrement,PrimaryKey"`
	// User for which this credential is applied
	UserId int `json:"userId" dbColumn:"Column:user_id,ForeignKey:workout.user.id"`
	// Unique identification of the server type (e.g. strava)
	Type string `json:"type" dbColumn:"Column:type"`
	// Optional (base) server address to connect to
	ServerAddress string `json:"serverAddress" dbColumn:"Column:server_address"`
	// Type specific credentials used to connect to the remote API server
	Credentials string `json:"credentials" dbColumn:"Column:credentials"`
	DbMetadata_ any    `json:"-" dbMetadata:"Schema:workout,Table:external_api"`
}

// ExternalApi
const (
	ExternalApi_Id            string = "Id|workout.external_api.id"
	ExternalApi_UserId        string = "UserId|workout.external_api.user_id"
	ExternalApi_Type          string = "Type|workout.external_api.type"
	ExternalApi_ServerAddress string = "ServerAddress|workout.external_api.server_address"
	ExternalApi_Credentials   string = "Credentials|workout.external_api.credentials"
)

package models

type Tag struct {
	// Unique ID of this tag
	Id int `json:"id" dbColumn:"Column:id,PrimaryKey"`
	// Short description name of the tag
	Name string `json:"name" dbColumn:"Column:name"`
	// Color code (#f20102) of the tag for the dark mode
	TagDark string `json:"tagDark" dbColumn:"Column:tag_dark"`
	// Color code (#f20102) of the tag for the white mode
	TagWhite    string `json:"tagWhite" dbColumn:"Column:tag_white"`
	DbMetadata_ any    `json:"-" dbMetadata:"Schema:workout,Table:tag"`
}

// Tag
const (
	Tag_Id       string = "Id|workout.tag.id"
	Tag_Name     string = "Name|workout.tag.name"
	Tag_TagDark  string = "TagDark|workout.tag.tag_dark"
	Tag_TagWhite string = "TagWhite|workout.tag.tag_white"
)

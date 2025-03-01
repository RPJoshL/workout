package automation

import (
	"strings"

	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
)

var (
	ErrInvalidColor = errors.NewError("#settings.automation.invalidColor", 400)
)

// TagRules extends [models.Tag] with additional information
// to the per user defined rules
type TagRules struct {
	models.Tag

	// How many rules exists with this tag
	RulesCount int `json:""`
}

// GetAllTags returns a list of all tags within the whole db
func (a *Api) GetAllTags() (rtc []TagRules, err errors.Error) {
	sel := a.R().Db.Struct.QuerySlice(&rtc)
	sel.CustomColumn("", "RulesCount", "NVL(ruleCount.cnt, 0)")
	sel.CustomJoin(
		`LEFT JOIN (
			SELECT COUNT(*) AS cnt, rt.tag_id
			FROM rule_tagging rt
			WHERE rt.user_id = ?
			GROUP BY rt.tag_id
		) ruleCount ON ruleCount.tag_id = tag.id`,
		a.R().User.Id,
	)

	if e := sel.Run(); e != nil {
		err = e.GetResponse().Log("Failed to select tags", e.GetError(), a)
	}

	return
}

// CreateTag creates new tag based on the provided values
func (a *Api) CreateTag(tag models.Tag) errors.Error {
	tag.Id = 0

	// Validate
	if err := validateTag(tag); err != nil {
		return err
	}

	if _, err := a.R().Db.Struct.Insert(&tag).Run(); err != nil {
		return errors.InternalError().Log("Failed to create tag", err, a)
	}

	return nil
}

func (a *Api) UpdateTag(tag models.Tag) errors.Error {
	// Validate
	if err := validateTag(tag); err != nil {
		return err
	}

	if err := a.R().Db.Struct.Update(&tag).Run(); err != nil {
		return errors.InternalError().Log("Failed to update tag", err, a)
	}

	return nil
}

// validateTag validates the provided tag values
func validateTag(tag models.Tag) errors.Error {
	// Colors
	colors := []string{tag.TagDark, tag.TagWhite}
	for _, c := range colors {
		// Vadlic colors are '#fff' or '#ffffff'
		if !(len(c) == 4 || len(c) == 7) {
			return ErrInvalidColor.Sprintf(c)
		}

		// Has to begin with a #
		if !strings.HasPrefix(c, "#") {
			return ErrInvalidColor.Sprintf(c)
		}
	}

	return nil
}

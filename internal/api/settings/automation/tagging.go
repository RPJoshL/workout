package automation

import (
	"git.rpjosh.de/RPJosh/go-ddl-parser"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/pkg/database"
	"git.rpjosh.de/RPJosh/workout/pkg/database/dbstruct"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
)

// GetAllRules returns a list of all automation rules for the current user
func (a *Api) GetAllRules() (rtc []models.RuleTagging, e errors.Error) {
	sel := a.R().Db.Struct.QuerySlice(&rtc)
	sel.Where().Column(models.RuleTagging_UserId, "=", a.R().User.Id).Add()
	sel.OrderBy("", models.RuleTagging_Name, "ASC")

	if err := sel.Run(); err != nil {
		return rtc, errors.InternalError().Log("Failed to select all automation rules", err, a)
	}

	return
}

// getLastWorkoutLocation returns the starting point of the last workout
func (a *Api) getLastWorkoutLocation() (location *ddl.Location, err errors.Error) {
	var lastId int

	if err := a.R().Db.QueryForValue(
		&lastId,
		"SELECT MAX(ID) FROM workout w WHERE user_id = ?",
		a.R().User.Id,
	); err != nil {
		if err.Type() == database.NoRows {
			//nolint:all nil describes no workout location available here. It's not a hard error!
			return nil, nil
		} else {
			return nil, errors.InternalError().Log("Failed to select last workout id", err, a)
		}
	}

	workout := models.Workout{}
	sel := a.R().Db.Struct.Query(&workout).Where().Column(models.Workout_Id, "=", lastId).Add()
	if err := sel.Run(); err != nil {
		return nil, errors.InternalError().Log("Failed to query last workout with ID %d", err, a, lastId)
	}

	return &workout.CityLocation, nil
}

func (a *Api) validateRule(rule models.RuleTagging) errors.Error {
	// Validate name
	if rule.Name == "" {
		return errors.BadRequest(a.R().Tr.Get("settings.automation.nameRequired"))
	}
	rulesWithName := []models.RuleTagging{}
	rulesSel := a.R().Db.Struct.QuerySlice(&rulesWithName).Where().Column(models.RuleTagging_UserId, "=", a.R().User.Id).Add()
	rulesSel.Where().Column(models.RuleTagging_Name, "=", rule.Name).Add()
	if err := rulesSel.Run(); err != nil {
		return errors.InternalError().Log("Failed to select rules with name %q", err, a, rule.Name)
	}
	if len(rulesWithName) != 0 && rulesWithName[0].Id != rule.Id {
		return errors.BadRequest(a.R().Tr.Get("settings.automation.ruleExists"))
	}

	// Validate tag
	tagSel := a.R().Db.Struct.Query(&models.Tag{}).Where().Column(models.Tag_Id, "=", rule.TagId).Add()
	if cnt, err := tagSel.Count(); err != nil {
		return errors.InternalError().Log("Failed to select tag with id %d", err, a, rule.TagId)
	} else if cnt != 1 {
		return errors.BadRequest(a.R().Tr.Get("settings.automation.tagNotExists"))
	}

	return nil
}

func (a *Api) CreateRule(rule models.RuleTagging) errors.Error {
	rule.Id = 0
	rule.UserId = a.R().User.Id

	if err := a.validateRule(rule); err != nil {
		return err
	}

	ins := a.R().Db.Struct.Insert(&rule).Selector(dbstruct.ColumnSelector{ForeignKeyReference: true})
	if _, err := ins.Run(); err != nil {
		return errors.InternalError().Log("Failed to insert tag", err, a)
	}

	return nil
}

func (a *Api) UpdateRule(rule models.RuleTagging) errors.Error {
	rule.UserId = a.R().User.Id

	var existing models.RuleTagging
	sel := a.R().Db.Struct.Query(&existing).Where().Column(models.RuleTagging_Id, "=", rule.Id).Add()
	sel.Where().Column(models.RuleTagging_UserId, "=", rule.UserId).Add()
	if cnt, err := sel.Count(); err != nil {
		return errors.InternalError().Log("Failed to select existing tag #%d", err, a, rule.Id)
	} else if cnt == 0 {
		return errors.NotFound()
	}

	upd := a.R().Db.Struct.Update(&rule).Selector(dbstruct.ColumnSelector{ForeignKeyReference: true})
	if err := upd.Run(); err != nil {
		return errors.InternalError().Log("Failed to update tag #%d", err, a, rule.Id)
	}

	return nil
}

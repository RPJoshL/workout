package automation

import (
	"net/http"
	"strconv"

	"git.rpjosh.de/RPJosh/workout/internal/api/router"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
	"git.rpjosh.de/RPJosh/workout/pkg/response"
	"github.com/a-h/templ"
)

type Api struct {
	router.ApiRequest
}

func (a *Api) GetRouter() *router.Router {
	routes := router.Routes{
		router.NewRoute(
			"AutomationPage",
			"GET",
			"/automation/",
			a.AutomationPage,
			router.Options{},
		),
		router.NewRoute(
			"CreateTag",
			"GET",
			"/automation/tag/{id}",
			a.TagModal,
			router.Options{},
		),
		router.NewRoute(
			"CreateUpdateTagFormEndpoint",
			"POST",
			"/automation/tag",
			a.CreateOrUpdateTag,
			router.Options{},
		),
		router.NewRoute(
			"DeleteTag",
			"DELETE",
			"/automation/tag/{id}",
			a.DeleteTag,
			router.Options{},
		),
		router.NewRoute(
			"CreateEditRule",
			"GET",
			"/automation/rule/{id}",
			a.TaggingModal,
			router.Options{},
		),
		router.NewRoute(
			"CreateUpdateRuleFormEndpoint",
			"POST",
			"/automation/rule",
			a.CreateOrUpdateRule,
			router.Options{},
		),
		router.NewRoute(
			"DeleteRule",
			"DELETE",
			"/automation/rule/{id}",
			a.DeleteRule,
			router.Options{},
		),
	}

	return &router.Router{
		Dependency: a,
		Routes:     routes,
	}
}

func (a *Api) AutomationPage(w http.ResponseWriter, r *http.Request) {
	if page, err := a.getAutomationPage(); err != nil {
		err.GetErrorStruct().Write(w, r)
	} else {
		a.R().Tmpl.Render(page, "generic.appName", "generic.appName")
	}
}

func (a *Api) getAutomationPage() (templ.Component, errors.Error) {
	tags, err := a.GetAllTags()
	if err != nil {
		return nil, err
	}

	rules, err := a.GetAllRules()
	if err != nil {
		return nil, err
	}

	return a.automationPage(tags, rules), nil
}

func (a *Api) TagModal(w http.ResponseWriter, r *http.Request) {
	if page, err := a.getAutomationPage(); err != nil {
		err.GetErrorStruct().Write(w, r)
	} else {
		tag := models.Tag{}
		id := r.PathValue("id")
		if id != "" && id != "new" {
			idInt, err := strconv.Atoi(id)
			if err != nil {
				errors.BadRequest(a.R().Tr.Getf("generic.numericError", "id", id)).Write(w, r)
				return
			}

			sel := a.R().Db.Struct.Query(&tag).Where().Column(models.Tag_Id, "=", idInt).Add()
			if err := sel.Run(); err != nil {
				err.GetResponse().Log("Failed to query tag with id %d", err, a, idInt).Write(w, r)
				return
			}
		}

		a.R().Tmpl.RenderModal(
			a.tagModal(tag), "settings.automation.newTag",
			page, "/settings/automation/", "generic.appName", "generic.appName", "",
		)
	}
}

func (a *Api) CreateOrUpdateTag(w http.ResponseWriter, r *http.Request) {
	// Parse values
	var tag models.Tag
	if err := a.R().Parser.Parse(&tag, router.RequestParserOptions{
		Mode:           router.ParseModeForm,
		InterpreteJson: true,
	}); err != nil {
		err.GetErrorStruct().Write(w, r)
	}

	// Create or update
	var err errors.Error
	if tag.Id == 0 {
		err = a.CreateTag(tag)
	} else {
		err = a.UpdateTag(tag)
	}

	if err != nil {
		err.GetErrorStruct().Write(w, r)
		return
	}

	// Inner content of the table div is swapped
	tags, err := a.GetAllTags()
	if err != nil {
		err.GetErrorStruct().Write(w, r)
	} else {
		a.R().Tmpl.RenderDirect(a.tagTable(tags))
	}
}

func (a *Api) DeleteTag(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		errors.BadRequest("#generic.numericError").Sprintf("id", r.PathValue("id")).Write(w, r)
		return
	}

	if res, e := a.R().Db.Db.Exec(`DELETE FROM tag WHERE id = ?`, id); e != nil {
		errors.InternalError().Log("Failed to delete tag with id %d", e, a, id).Write(w, r)
	} else {
		deleted, _ := res.RowsAffected()
		if deleted == 0 {
			errors.NotFound().Write(w, r)
		} else {
			response.WriteText("Deleted", 200, w)
		}
	}
}

func (a *Api) TaggingModal(w http.ResponseWriter, r *http.Request) {
	if page, err := a.getAutomationPage(); err != nil {
		err.GetErrorStruct().Write(w, r)
	} else {
		rule := models.RuleTagging{}
		id := r.PathValue("id")
		if id != "" && id != "new" {
			idInt, err := strconv.Atoi(id)
			if err != nil {
				errors.BadRequest(a.R().Tr.Getf("generic.numericError", "id", id)).Write(w, r)
				return
			}

			sel := a.R().Db.Struct.Query(&rule).Where().Column(models.RuleTagging_Id, "=", idInt).Add()
			sel.Where().Column(models.RuleTagging_UserId, "=", a.R().User.Id)
			if err := sel.Run(); err != nil {
				err.GetResponse().Log("Failed to query tagging rule with id %d", err, a, idInt).Write(w, r)
				return
			}
		}

		tags, err := a.GetAllTags()
		if err != nil {
			err.GetErrorStruct().Write(w, r)
			return
		}

		lastWorkout, err := a.getLastWorkoutLocation()
		if err != nil {
			err.GetErrorStruct().Write(w, r)
			return
		}

		a.R().Tmpl.RenderModal(
			a.taggingModal(rule, tags, lastWorkout), "settings.automation.newTaggingRule",
			page, "/settings/automation/", "generic.appName", "generic.appName", "",
		)
	}
}

func (a *Api) CreateOrUpdateRule(w http.ResponseWriter, r *http.Request) {
	// Parse values
	var rule models.RuleTagging
	if err := a.R().Parser.Parse(&rule, router.RequestParserOptions{
		Mode:           router.ParseModeForm,
		InterpreteJson: true,
		Recursive:      true,
	}); err != nil {
		err.GetErrorStruct().Write(w, r)
		return
	}

	var err errors.Error
	if rule.Id == 0 {
		err = a.CreateRule(rule)
	} else {
		err = a.UpdateRule(rule)
	}

	if err != nil {
		err.GetErrorStruct().Write(w, r)
		return
	}

	// Inner content of the table div is swapped
	rules, err := a.GetAllRules()
	if err != nil {
		err.GetErrorStruct().Write(w, r)
		return
	}
	tags, err := a.GetAllTags()
	if err != nil {
		err.GetErrorStruct().Write(w, r)
	} else {
		a.R().Tmpl.RenderDirect(a.taggingTable(rules, tags))
	}
}

func (a *Api) DeleteRule(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		errors.BadRequest("#generic.numericError").Sprintf("id", r.PathValue("id")).Write(w, r)
		return
	}

	if res, e := a.R().Db.Db.Exec(`DELETE FROM rule_tagging WHERE id = ? AND user_id = ?`, id, a.R().User.Id); e != nil {
		errors.InternalError().Log("Failed to delete tagging rule with id %d", e, a, id).Write(w, r)
	} else {
		deleted, _ := res.RowsAffected()
		if deleted == 0 {
			errors.NotFound().Write(w, r)
		} else {
			response.WriteText("Deleted", 200, w)
		}
	}
}

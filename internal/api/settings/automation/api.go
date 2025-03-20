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

func (api *Api) GetRouter() *router.Router {
	routes := router.Routes{
		router.NewRoute(
			"AutomationPage",
			"GET",
			"/automation/",
			api.AutomationPage,
			router.Options{},
		),
		router.NewRoute(
			"CreateTag",
			"GET",
			"/automation/tag/{id}",
			api.TagModal,
			router.Options{},
		),
		router.NewRoute(
			"CreateUpdateTagFormEndpoint",
			"POST",
			"/automation/tag",
			api.CreateOrUpdateTag,
			router.Options{},
		),
		router.NewRoute(
			"DeleteTag",
			"DELETE",
			"/automation/tag/{id}",
			api.DeleteTag,
			router.Options{},
		),
		router.NewRoute(
			"CreateEditRule",
			"GET",
			"/automation/rule/{id}",
			api.TaggingModal,
			router.Options{},
		),
		router.NewRoute(
			"CreateUpdateRuleFormEndpoint",
			"POST",
			"/automation/rule",
			api.CreateOrUpdateRule,
			router.Options{},
		),
		router.NewRoute(
			"DeleteRule",
			"DELETE",
			"/automation/rule/{id}",
			api.DeleteRule,
			router.Options{},
		),
	}

	return &router.Router{
		Dependency: api,
		Routes:     routes,
	}
}

func (api *Api) AutomationPage(w http.ResponseWriter, r *http.Request) {
	if page, err := api.getAutomationPage(); err != nil {
		err.GetErrorStruct().Write(w, r)
	} else {
		api.R().Tmpl.Render(page, "generic.appName", "generic.appName")
	}
}

func (api *Api) getAutomationPage() (templ.Component, errors.Error) {
	tags, err := api.GetAllTags()
	if err != nil {
		return nil, err
	}

	rules, err := api.GetAllRules()
	if err != nil {
		return nil, err
	}

	return api.automationPage(tags, rules), nil
}

func (api *Api) TagModal(w http.ResponseWriter, r *http.Request) {
	if page, err := api.getAutomationPage(); err != nil {
		err.GetErrorStruct().Write(w, r)
	} else {
		tag := models.Tag{}
		id := r.PathValue("id")
		if id != "" && id != "new" {
			idInt, err := strconv.Atoi(id)
			if err != nil {
				errors.BadRequest(api.R().Tr.Getf("generic.numericError", "id", id)).Write(w, r)
				return
			}

			sel := api.R().Db.Struct.Query(&tag).Where().Column(models.Tag_Id, "=", idInt).Add()
			if err := sel.Run(); err != nil {
				err.GetResponse().Log("Failed to query tag with id %d", err, api, idInt).Write(w, r)
				return
			}
		}

		api.R().Tmpl.RenderModal(
			api.tagModal(tag), "settings.automation.newTag",
			page, "/settings/automation/", "generic.appName", "generic.appName", "",
		)
	}
}

func (api *Api) CreateOrUpdateTag(w http.ResponseWriter, r *http.Request) {
	// Parse values
	var tag models.Tag
	if err := api.R().Parser.Parse(&tag, router.RequestParserOptions{
		Mode:           router.ParseModeForm,
		InterpreteJson: true,
	}); err != nil {
		err.GetErrorStruct().Write(w, r)
	}

	// Create or update
	var err errors.Error
	if tag.Id == 0 {
		err = api.CreateTag(tag)
	} else {
		err = api.UpdateTag(tag)
	}

	if err != nil {
		err.GetErrorStruct().Write(w, r)
		return
	}

	// Inner content of the table div is swapped
	tags, err := api.GetAllTags()
	if err != nil {
		err.GetErrorStruct().Write(w, r)
	} else {
		api.R().Tmpl.RenderDirect(api.tagTable(tags))
	}
}

func (api *Api) DeleteTag(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		errors.BadRequest("#generic.numericError").Sprintf("id", r.PathValue("id")).Write(w, r)
		return
	}

	if res, e := api.R().Db.Db.Exec(`DELETE FROM tag WHERE id = ?`, id); e != nil {
		errors.InternalError().Log("Failed to delete tag with id %d", e, api, id).Write(w, r)
	} else {
		deleted, _ := res.RowsAffected()
		if deleted == 0 {
			errors.NotFound().Write(w, r)
		} else {
			response.WriteText("Deleted", 200, w)
		}
	}
}

func (api *Api) TaggingModal(w http.ResponseWriter, r *http.Request) {
	if page, err := api.getAutomationPage(); err != nil {
		err.GetErrorStruct().Write(w, r)
	} else {
		rule := models.RuleTagging{}
		id := r.PathValue("id")
		if id != "" && id != "new" {
			idInt, err := strconv.Atoi(id)
			if err != nil {
				errors.BadRequest(api.R().Tr.Getf("generic.numericError", "id", id)).Write(w, r)
				return
			}

			sel := api.R().Db.Struct.Query(&rule).Where().Column(models.RuleTagging_Id, "=", idInt).Add()
			sel.Where().Column(models.RuleTagging_UserId, "=", api.R().User.Id)
			if err := sel.Run(); err != nil {
				err.GetResponse().Log("Failed to query tagging rule with id %d", err, api, idInt).Write(w, r)
				return
			}
		}

		tags, err := api.GetAllTags()
		if err != nil {
			err.GetErrorStruct().Write(w, r)
			return
		}

		lastWorkout, err := api.getLastWorkoutLocation()
		if err != nil {
			err.GetErrorStruct().Write(w, r)
			return
		}

		api.R().Tmpl.RenderModal(
			api.taggingModal(rule, tags, lastWorkout), "settings.automation.newTaggingRule",
			page, "/settings/automation/", "generic.appName", "generic.appName", "",
		)
	}
}

func (api *Api) CreateOrUpdateRule(w http.ResponseWriter, r *http.Request) {
	// Parse values
	var rule models.RuleTagging
	if err := api.R().Parser.Parse(&rule, router.RequestParserOptions{
		Mode:           router.ParseModeForm,
		InterpreteJson: true,
		Recursive:      true,
	}); err != nil {
		err.GetErrorStruct().Write(w, r)
		return
	}

	var err errors.Error
	if rule.Id == 0 {
		err = api.CreateRule(rule)
	} else {
		err = api.UpdateRule(rule)
	}

	if err != nil {
		err.GetErrorStruct().Write(w, r)
		return
	}

	// Inner content of the table div is swapped
	rules, err := api.GetAllRules()
	if err != nil {
		err.GetErrorStruct().Write(w, r)
		return
	}
	tags, err := api.GetAllTags()
	if err != nil {
		err.GetErrorStruct().Write(w, r)
	} else {
		api.R().Tmpl.RenderDirect(api.taggingTable(rules, tags))
	}
}

func (api *Api) DeleteRule(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		errors.BadRequest("#generic.numericError").Sprintf("id", r.PathValue("id")).Write(w, r)
		return
	}

	if res, e := api.R().Db.Db.Exec(`DELETE FROM rule_tagging WHERE id = ? AND user_id = ?`, id, api.R().User.Id); e != nil {
		errors.InternalError().Log("Failed to delete tagging rule with id %d", e, api, id).Write(w, r)
	} else {
		deleted, _ := res.RowsAffected()
		if deleted == 0 {
			errors.NotFound().Write(w, r)
		} else {
			response.WriteText("Deleted", 200, w)
		}
	}
}

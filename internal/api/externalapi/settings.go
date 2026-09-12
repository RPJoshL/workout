package externalapi

import (
	extapi "git.rpjosh.de/RPJosh/workout/internal/externalapi"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
)

func (a *Api) getApiByName(typ string) (extapi.API, models.ExternalAPIType, bool) {
	apiType, ok := models.ExternalAPIFromString[typ]
	if !ok {
		return nil, "", false
	}

	api, ok := a.apiLookup[apiType]
	if !ok {
		return nil, "", false
	}

	return api, apiType, true
}

func (a *Api) saveExternalAPI(config *models.ExternalApi) errors.Error {
	config.UserId = a.R().User.Id

	userAPIs, err := a.GetExternalAPIs()
	if err != nil {
		return err
	}

	if api, ok := userAPIs[models.ExternalAPIType(config.Type)]; ok {
		config.Id = api.Id
		if err := a.R().Db.Struct.Update(config).Run(); err != nil {
			return err.GetResponse().Log("Failed to update external API config", err.GetError(), a)
		}

		return nil
	}

	if _, err := a.R().Db.Struct.Insert(config).Run(); err != nil {
		return err.GetResponse().Log("Failed to insert external API config", err.GetError(), a)
	}

	return nil
}

func (a *Api) deleteExternalAPI(apiType models.ExternalAPIType) errors.Error {
	_, e := a.R().Db.Db.Exec(
		`DELETE FROM external_api WHERE user_id = ? AND type = ?`,
		a.R().User.Id, string(apiType),
	)
	if e != nil {
		return errors.InternalError().Log("Failed to delete external API config", e, a)
	}

	return nil
}

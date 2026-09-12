package externalapi

import (
	"slices"

	"git.rpjosh.de/RPJosh/workout/internal/externalapi"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
)

// GetExternalAPIs returns all configured external API for the current user
func (a *Api) GetExternalAPIs() (map[models.ExternalAPIType]*models.ExternalApi, errors.Error) {
	var arr []models.ExternalApi

	sel := a.R().Db.Struct.QuerySlice(&arr)
	sel.Where().Column(models.ExternalApi_UserId, "=", a.R().User.Id).Add()

	if err := sel.Run(); err != nil {
		return nil, err.GetResponse().Log("Failed to query external API config", err.GetError(), a)
	}

	rtc := make(map[models.ExternalAPIType]*models.ExternalApi, len(arr))
	for i := range arr {
		rtc[models.ExternalAPIType(arr[i].Type)] = &arr[i]
	}

	return rtc, nil
}

// ResolveExternalAPIs returns all configured external API services
func (a *Api) ResolveExternalAPIs() (map[models.ExternalAPIType]externalapi.API, errors.Error) {
	byConfig, err := a.GetExternalAPIs()
	if err != nil {
		return nil, err
	}

	rtc := make(map[models.ExternalAPIType]externalapi.API, len(byConfig))
	for typ := range byConfig {
		service, ok := a.apiLookup[typ]
		if !ok {
			continue
		}

		rtc[typ] = service
	}

	return rtc, nil
}

// ResolveExternalAPIsForType returns all configured external API services that are applicable
// for the provided workout type
func (a *Api) ResolveExternalAPIsForType(typ int) (map[models.ExternalAPIType]externalapi.API, errors.Error) {
	configuredAPIs, err := a.ResolveExternalAPIs()
	if err != nil {
		return nil, err
	}

	rtc := make(map[models.ExternalAPIType]externalapi.API, len(configuredAPIs))
	for apiType := range configuredAPIs {
		applicalableTypes := configuredAPIs[apiType].Config().WorkoutTypes

		if len(applicalableTypes) == 0 || slices.Contains(applicalableTypes, typ) {
			rtc[apiType] = configuredAPIs[apiType]
		}
	}

	return rtc, nil
}

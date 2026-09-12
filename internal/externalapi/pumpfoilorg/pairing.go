package pumpfoilorg

import (
	"fmt"
	"net/http"

	"git.rpjosh.de/RPJosh/workout/internal/api/components/form"
	"git.rpjosh.de/RPJosh/workout/internal/api/router"
	"git.rpjosh.de/RPJosh/workout/internal/externalapi"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
)

type PairingData struct {
	ServerURL  string `form:"serverURL"`
	Label      string `form:"label"`
	Code       string `form:"code"`
	ClaimToken string `form:"claimToken"`
	Waiting    bool   `form:"waiting"`
	FoilID     int    `form:"foilId"`
}

type authInit struct{}

func (a *authInit) Fields() []form.Field {
	return []form.Field{
		{
			Type:  form.Text,
			Name:  "serverURL",
			Label: "settings.externalAPI.serverURL",
			Value: defaultURL,
		},
		{
			Type:  form.Text,
			Name:  "label",
			Label: "settings.externalAPI.deviceLabel",
			Hint:  "settings.externalAPI.deviceHint",
		},
		{
			Type:  form.Number,
			Name:  "foilId",
			Label: "settings.externalAPI.foilLabel",
			Hint:  "settings.externalAPI.foilHint",
		},
	}
}

func (a *authInit) Execute(r *http.Request, ui externalapi.AuthenticationUI) (*externalapi.AuthenticationStepResult, error) {
	var data PairingData
	if err := ui.R.Parser.Parse(&data, router.RequestParserOptions{
		Mode: router.ParseModeForm,
	}); err != nil {
		return nil, err
	}

	if data.ServerURL == "" {
		data.ServerURL = defaultURL
	}
	if data.Label == "" {
		data.Label = "RPout"
	}

	client := NewPumpfoilOrgClient(data.ServerURL, "")
	initRes, err := client.PairInit(data.Label)
	if err != nil {
		return nil, err
	}

	data.Code = initRes.Code
	data.ClaimToken = initRes.ClaimToken

	return &externalapi.AuthenticationStepResult{
		NextComponent: PairingStep(data, ui),
	}, nil
}

type authPoll struct{}

func (a *authPoll) Fields() []form.Field {
	return []form.Field{}
}

func (a *authPoll) Execute(r *http.Request, ui externalapi.AuthenticationUI) (*externalapi.AuthenticationStepResult, error) {
	var data PairingData
	if err := ui.R.Parser.Parse(&data, router.RequestParserOptions{
		Mode: router.ParseModeForm,
	}); err != nil {
		return nil, err
	}

	if data.ServerURL == "" || data.ClaimToken == "" {
		return nil, fmt.Errorf("missing pairing data")
	}

	client := NewPumpfoilOrgClient(data.ServerURL, "")
	pollRes, err := client.PairPoll(data.ClaimToken)
	if err != nil {
		return nil, err
	}

	if pollRes.DeviceToken == nil || *pollRes.DeviceToken == "" {
		data.Waiting = true

		return &externalapi.AuthenticationStepResult{
			NextComponent: PairingStep(data, ui),
		}, nil
	}

	creds, err := encodePumpfoilOrgCredentials(*pollRes.DeviceToken, data.FoilID)
	if err != nil {
		return nil, errors.InternalError().Log("Failed to encode credentials", err, a)
	}

	return &externalapi.AuthenticationStepResult{
		Config: &models.ExternalApi{
			Type:          string(models.ExternalAPIPumpfoilOrg),
			ServerAddress: data.ServerURL,
			Credentials:   creds,
		},
	}, nil
}

package token

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"time"

	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/pkg/database"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
	"git.rpjosh.de/RPJosh/workout/pkg/utils"
)

var (
	ErrNotPriveleged = errors.NewError("Forbidden (user must be priveleged)", 403)
	ErrNoTokenAuth   = errors.BadRequest("Not authenticated with API key")
)

// CreateToken creates a new token in the database with the provided data
func (a *Api) CreateToken(data models.ApiKey, offset int) (models.ApiKey, errors.Error) {
	origKeyValue := ""

	// User has to be priveleged
	if !a.R().User.Priveleged {
		return data, ErrNotPriveleged
	}

	// Generate a random token and hash it
	for {
		randomBytes, _ := utils.GenerateRandomBytes(32)
		hashedValue := a.hashToken(randomBytes)

		// Token mustn't exist in datbase already
		sel := a.R().Db.Struct.Query(&models.ApiKey{}).Where().Column(models.ApiKey_Key, "=", hashedValue).Add()
		if count, err := sel.Count(); err != nil {
			return data, errors.InternalError().Log("Failed to select existing API Key", err, a)
		} else if count == 0 {
			origKeyValue = hex.EncodeToString(randomBytes)
			data.Key = a.hashToken(randomBytes)
			data.Obfuscated = fmt.Sprintf("%s...%s", origKeyValue[0:3], origKeyValue[len(origKeyValue)-3:])
			break
		}
	}

	// Fill default values
	data.UserId = a.R().User.Id
	data.CreationDate = time.Now()
	if offset > 0 {
		data.ValidUntil = time.Now().Add(time.Minute * time.Duration(offset))
	} else if data.ValidUntil.Before(time.Now()) {
		data.ValidUntil = time.Now().AddDate(1, 0, 0)
	}

	// Create the token
	if id, err := a.R().Db.Struct.Insert(&data).Run(); err != nil {
		return data, errors.InternalError().Log("Failed to create token", err, a)
	} else {
		if err := a.R().Db.Struct.Query(&data).Where().Column(models.ApiKey_Id, "=", id).Add().Run(); err != nil {
			return data, errors.InternalError().Log("Failed to select token", err, a)
		}

		// Set original value for the client
		data.Key = origKeyValue
		return data, nil
	}
}

// IsTokenValid checks whether the provided (RAW) token value was found
// within the database and is still valid
func (a *Api) IsTokenValid(token string) (rtc models.ApiKey, err errors.Error) {
	tokenStr, _ := hex.DecodeString(token)
	sel := a.R().Db.Struct.Query(&rtc).Where().Column(models.ApiKey_Key, "=", a.hashToken(tokenStr)).Add()
	sel.Where().Column(models.ApiKey_ValidUntil, ">", time.Now()).Add()

	if err := sel.Run(); err != nil {
		if err.Type() == database.NoRows {
			return rtc, errors.NewError("Invalid API-Key", 401)
		} else {
			return rtc, errors.InternalError().Log("Failed to select API key", err, a)
		}
	}

	return
}

// hashToken creates a hash value of the provided bytes that was
// used to store the token within the database
func (a *Api) hashToken(value []byte) string {
	hasher := sha512.New()
	if _, err := hasher.Write(value); err != nil {
		a.Logger().Error("Failed to write value to hasher: %s", err)
		return ""
	}

	return hex.EncodeToString(hasher.Sum(nil))
}

// showApikey returns the provided API
func (a *Api) showApikey(id int) (rtc models.ApiKey, err errors.Error) {

	// Show the API key that was used for authentication
	if id == -1 || a.R().User.ApiKey.Id == id {
		// Not authenticated with API key
		if a.R().User.ApiKey.Id == 0 {
			return rtc, ErrNoTokenAuth
		}

		return a.R().User.ApiKey, nil
	}

	// User has to be priveleged
	if !a.R().User.Priveleged {
		return rtc, ErrNotPriveleged
	}

	sel := a.R().Db.Struct.Query(&rtc)
	sel.Where().Column(models.ApiKey_Id, "=", id).Add()
	sel.Where().Column(models.ApiKey_UserId, "=", a.R().User.Id).Add()

	if e := sel.Run(); e != nil {
		if e.Type() == database.NoRows {
			return rtc, errors.NotFound()
		} else {
			return rtc, errors.InternalError().Log("Failed to select API key", e, a)
		}
	}

	return
}

// deleteApiKey deletes the provided API key
func (a *Api) deleteApiKey(id int) errors.Error {

	// Select API key to make sure it exists
	usedKey, err := a.showApikey(id)
	if err != nil {
		return err
	}

	// Delete it
	_, e := a.R().Db.Db.Exec(`DELETE FROM api_key WHERE id = ?`, usedKey.Id)
	if e != nil {
		return errors.InternalError().Log("Failed to delete API key", e, a)
	}

	return nil
}

// GetAllTokens returns a list of all tokens for the currently authenticated
// user
func (a *Api) GetAllTokens() (rtc []models.ApiKey, err errors.Error) {
	sel := a.R().Db.Struct.QuerySlice(&rtc)
	sel.Where().Column(models.ApiKey_UserId, "=", a.R().User.Id).Add()
	sel.Where().Column(models.ApiKey_ValidUntil, ">", time.Now()).Add()
	sel.OrderBy("", models.ApiKey_CreationDate, "DESC")
	if e := sel.Run(); e != nil {
		return rtc, errors.InternalError().Log("Failed to query tokens", e, a)
	}

	return
}

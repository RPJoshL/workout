package user

import (
	"testing"

	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/internal/tests"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
)

func TestUserLogin(t *testing.T) {
	api := &Api{}
	tests.InjectRequestData(api)

	// Create a simple test user
	password := "Aloah1234"
	passwordHashed, err := api.hashUserPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %s", err)
	}

	mail := "hallo@test.de"
	user := models.User{
		Name:     "Test",
		Password: passwordHashed,
		Mail:     mail,
	}
	if id, err := api.R().Db.Struct.Insert(&user).Run(); err != nil {
		t.Fatalf("Failed to insert user: %s", err)
	} else {
		user.Id = int(id)
	}

	// Incorrect password
	_, errGot := api.IsLoginCorrect(user.Mail, "Random!")
	if errors.IsNot(errGot, ErrPasswordIncorrect) {
		t.Errorf("Incorrect error received. Expected: %s, Got: %s", ErrPasswordIncorrect, errGot)
	}

	// Correct credentials
	idGot, errGot := api.IsLoginCorrect(user.Mail, password)
	if errGot != nil {
		t.Errorf("Unexpected error for correct login: %s", errGot)
	} else if idGot != user.Id {
		t.Errorf("Received incorrect user user. Expectd %d. Got %d", user.Id, idGot)
	}
}

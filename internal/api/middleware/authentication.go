package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"git.rpjosh.de/RPJosh/workout/internal/api/jwto"
	"git.rpjosh.de/RPJosh/workout/internal/api/utils"
	"git.rpjosh.de/RPJosh/workout/internal/dbutils"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/pkg/database"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
	"git.rpjosh.de/RPJosh/workout/pkg/response"
	"git.rpjosh.de/RPJosh/workout/pkg/webserver"
	"github.com/RPJoshL/go-logger"
)

// Name of the authentication cookie
const (
	CookieNameAuth           = "WorkoutCookie"
	CookieNameAndroidWebview = "AndroidClientVersion"
	CookieNameAndroidTheme   = "AndroidClientTheme"
)

// IsLoginCorrect checks if the provided username and password are
// correct and returns the matching user ID
type IsLoginCorrect func(mail, password string) (int, errors.Error)

// IsApiKeyCorrect checks whether the provided API key is valid
type IsApiKeyCorrect func(token string) (models.ApiKey, errors.Error)

var GlobalIsLoginCorrect IsLoginCorrect
var GlobalIsApiKeyCorrect IsApiKeyCorrect

// AuthenticationMiddleware is a middleware for validating JWT Tokens.
// Therefore, an "Authorization" header with the "Bearer" schema or a cookie
// with the token is required.
// If no valid token was provided, 401 will be returned immediately
func AuthenticationMiddleware(next http.Handler, key []byte, db *dbutils.Db) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		webUser := &models.WebUser{User: &models.User{}}

		// If we implement an API, we should not return a redirect!
		isApi := strings.HasPrefix(r.URL.Path, "/api")
		redirectURL := utils.BuildUrl("/user/login", "redirectTo", r.URL.Path)

		// Authentication by username and password
		userId, e := authByUsernamePassword(r)
		if e != nil {
			e.GetErrorStruct().Write(w, r)
			return
		}
		webUser.Id = userId
		webUser.Privileged = userId != 0

		if err := AuthenticateForWebview(r, webUser); err != nil {
			err.GetErrorStruct().Write(w, r)
			return
		}

		// Authentication by API key
		if apiHeader := r.Header.Get("X-Api-Key"); webUser.Id == 0 && apiHeader != "" {
			key, err := GlobalIsApiKeyCorrect(apiHeader)
			if err != nil {
				err.GetErrorStruct().Write(w, r)
				return
			}

			webUser.ApiKey = key
			webUser.Id = key.UserId
		}

		// Authentication by JWT token
		if webUser.Id == 0 {
			token, err := GetJwtToken(r)
			if err != nil {
				if isApi {
					errors.Write(w, r, err)
				} else {
					response.RedirectTo(redirectURL, 302, w, r)
				}
				return
			}

			claims, authorized, err := jwto.ValidateToken(token, key)
			if !authorized {
				logger.Debug("Not authorized: %s", err)
				if isApi {
					response.WriteText("Unauthorized", 401, w)
				} else {
					response.RedirectTo("/user/login", 302, w, r)
				}
				return
			}

			webUser.Id = claims.UserId

			// User is privileged if token was created within last 10 minutes
			if claims.IssuedAt != nil {
				webUser.Privileged = claims.IssuedAt.After(time.Now().Add(-10 * time.Minute))
			}
		}

		// Select full user from database
		qer := db.Struct.Query(webUser.User)
		qer.Where().Column(models.User_Id, "=", webUser.Id).Add()

		if err := qer.Run(); err != nil {
			// User is already deleted
			if err.Type() == database.NoRows {
				logger.Debug("User does not exist anymore: %d", webUser.Id)
				// Set empty cookie to delete the existing one
				c := &http.Cookie{
					Name:    CookieNameAuth,
					Value:   "",
					Path:    "/",
					Expires: time.Unix(0, 0),

					HttpOnly: true,
				}
				http.SetCookie(w, c)

				if isApi {
					response.WriteText("Unauthorized", 401, w)
				} else {
					response.RedirectTo("/user/login", 302, w, r)
				}
				return
			}

			logger.Warning("Failed to select user from database: %s", err)
			response.WriteError(err.GetResponse(), w, r)
			return
		}

		// Apply additional properties for a web user
		webUser.SetClientTimeZone(r.Header.Get("Time-Zone"))
		webUser.SetTheme(getValueFromCookie(r, CookieNameAndroidTheme))

		// Set user object accessible for all endpoints
		req := r.WithContext(context.WithValue(r.Context(), models.KeyUser, webUser))
		req = req.WithContext(context.WithValue(req.Context(), webserver.KeyUsername, webUser.Name))

		next.ServeHTTP(w, req)
	})
}

func getValueFromCookie(r *http.Request, name string) string {
	cookie, err := r.Cookie(name)
	if err != nil {
		return ""
	}

	return cookie.Value
}

// AuthenticateForWebview checks whether the request is coming from the Android Webview and
// sets the corresponding properties in the user struct.
//
// The webview should provide the API key in a cookie as a raw string (not as a JWT token)
func AuthenticateForWebview(r *http.Request, user *models.WebUser) errors.Error {
	user.AndroidWebview = strings.EqualFold("true", r.Header.Get("Client-Webview"))

	// Fallback to cookie detection as headers are lost on navigation
	cookieVersion, errCookie := r.Cookie(CookieNameAndroidWebview)
	if errCookie == nil && !user.AndroidWebview {
		user.AndroidWebview = true
		user.AppVersion = cookieVersion.Value
	}

	if !user.AndroidWebview {
		return nil
	}

	if user.AppVersion == "" {
		user.AppVersion = r.Header.Get("Client-Version")
	}

	cookie, errCookie := r.Cookie(CookieNameAuth)
	if errCookie != nil {
		//nolint:nilerr // No cookie provided, so we can not authenticate the user
		return nil
	}

	key, err := GlobalIsApiKeyCorrect(cookie.Value)
	if err != nil {
		return err
	}

	user.ApiKey = key
	user.Id = key.UserId

	return nil
}

// AuthByUsernamePassword handles the authentication by username and passowrd
// and returns the authenticated user if this authentication mode was used
func authByUsernamePassword(r *http.Request) (userId int, err errors.Error) {
	username := r.Header.Get("Username")
	password := r.Header.Get("Password")

	// Username and password required
	if username == "" || password == "" {
		return 0, nil
	}

	return GlobalIsLoginCorrect(username, password)
}

// GetJwtToken returns an JWT token from the request.
// The token is read either from a cookie or from the authorization
// header
func GetJwtToken(r *http.Request) (string, error) {
	authHeader := strings.Split(r.Header.Get("Authorization"), "Bearer ")
	cookie, errCookie := r.Cookie(CookieNameAuth)

	// Check if any JWT token was provided in the reqest
	if len(authHeader) != 2 && errCookie != nil {
		if len(authHeader) == 1 {
			return "", errors.NewError("No authorization token or cookie given", 403)
		} else {
			logger.Debug("Received malformed JWT token: %s", authHeader)
			return "", errors.NewError("Malformed token", 401)
		}
	}

	var token string
	if len(authHeader) == 2 {
		token = authHeader[1]
	} else {
		token = cookie.Value
	}
	return token, nil
}

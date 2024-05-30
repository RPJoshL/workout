package middleware

import (
	"context"
	"net/http"
	"strings"

	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/workout/internal/api/jwto"
	"git.rpjosh.de/RPJosh/workout/internal/api/utils"
	"git.rpjosh.de/RPJosh/workout/internal/database"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
	"git.rpjosh.de/RPJosh/workout/pkg/response"
	"git.rpjosh.de/RPJosh/workout/pkg/webserver"
)

// Name of the authentication cookie
const CookieName = "WorkoutCookie"

// AuthenticationMiddleware is a middleware for validating JWT Tokens.
// Therefore, an "Authorization" header with the "Bearer" schema or a cookie
// with the token is required.
// If no valid token was provided, 401 will be returned immediately
func AuthenticationMiddleware(next http.Handler, key []byte, db *database.DatabaseUtils) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// If we implement an API, we should not return a redirect!
		isApi := false
		redirectURL := utils.BuildUrl("/user/login", "redirectTo", r.URL.Path)

		token, err := GetJwtToken(r)
		if err != nil {
			if isApi {
				errors.Write(w, r, err)
			} else {
				response.RedirectTo(redirectURL, w, r)
			}
			return
		}

		claims, authorized, err := jwto.ValidateToken(token, key)
		if !authorized {
			logger.Debug("Not authorized: %s", err)
			if isApi {
				response.WriteText("Unauthorized", 401, w)
			} else {
				response.RedirectTo("/user/login", w, r)
			}
			return
		} else {

			// Select full user from database
			user := &models.User{}
			qer := db.Struct.Query(user)
			qer.Where().Column(models.User_Id, "=", claims.UserId).Add()
			if err := qer.Run(); err != nil {
				logger.Warning("Failed to select user from database: %s", err)
				response.WriteError(err.GetResponse(), w, r)
				return
			}

			// Set user object accessable for all endpoints
			req := r.WithContext(context.WithValue(r.Context(), models.KeyUser, user))
			req = req.WithContext(context.WithValue(req.Context(), webserver.KeyUsername, user.Name))

			next.ServeHTTP(w, req)
		}
	})
}

// GetJwtToken returns an JWT token from the request.
// The token is read either from a cookie or from the authorization
// header
func GetJwtToken(r *http.Request) (string, error) {
	authHeader := strings.Split(r.Header.Get("Authorization"), "Bearer ")
	cookie, errCookie := r.Cookie(CookieName)

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

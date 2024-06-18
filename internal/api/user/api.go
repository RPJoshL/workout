package user

import (
	"net/http"
	"strings"
	"time"

	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/workout/internal/api/jwto"
	"git.rpjosh.de/RPJosh/workout/internal/api/middleware"
	"git.rpjosh.de/RPJosh/workout/internal/api/router"
	"git.rpjosh.de/RPJosh/workout/internal/api/utils"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
	"git.rpjosh.de/RPJosh/workout/pkg/response"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v4"
)

type Api struct {
	router.ApiRequest

	conf *models.AppConfig
}

func GetRoutes(conf *models.AppConfig) *router.Router {
	api := &Api{
		conf: conf,
	}

	routes := router.Routes{
		router.NewRoute(
			"LoginPage",
			"GET",
			"/login",
			api.GetLoginPage,
			router.Options{
				UseNoAuth: true,
			},
		),
		router.NewRoute(
			"Login",
			"POST",
			"/login",
			api.Login,
			router.Options{
				UseNoAuth: true,
			},
		),
		router.NewRoute(
			"Logout",
			"POST",
			"/logout",
			api.Logout,
			router.Options{},
		),
		router.NewRoute(
			"DarkTheme",
			"POST",
			"/theme/{newTheme}",
			api.ChangeTheme,
			router.Options{},
		),
	}

	return &router.Router{
		Dependency: api,
		Routes:     routes,
	}
}

func (api *Api) GetLoginPage(w http.ResponseWriter, r *http.Request) {

	// The user has the option to specifiy the URL after login to redirect to.
	// This is automatically set if the user wants to access a site but wasn't
	// authorized yet
	redirectTo := r.FormValue("redirectTo")
	if redirectTo == "" {
		redirectTo = "/"
	}

	// If the user is authenticated, we don't display the login page!
	if api.isUserAuthorized(r) {
		response.RedirectTo(redirectTo, 302, w, r)
	} else {
		api.R().Tmpl.RenderWithoutLayout(api.LoginPage(redirectTo), "login.title", "login.description")
	}
}

// isUserAuthorized checks weather the user is already authorized.
// Because we use the option "UseNoAuth" for some login pages, we cannot
// check directly if the user is authorized
func (api *Api) isUserAuthorized(r *http.Request) bool {
	token, err := middleware.GetJwtToken(r)
	if err != nil {
		return false
	}

	// Try to parse it
	_, authorized, _ := jwto.ValidateToken(token, api.conf.JWTKey)
	return authorized
}

func (api *Api) Login(w http.ResponseWriter, r *http.Request) {

	// Extract parameters
	mail := r.FormValue("email")
	password := r.FormValue("password")

	// Check password
	userId, err := api.IsLoginCorrect(mail, password)
	if err != nil {
		err.GetErrorStruct().Write(w, r)
		return
	}

	// Create token
	expires := time.Now().Add(time.Duration(time.Hour * 6))
	if utils.IsTrue(r.FormValue("keepLoggedIn")) {
		expires = time.Now().AddDate(0, 0, 30)
	}
	jwtToken, erro := jwto.CreateToken(api.conf.JWTKey, &jwto.Claims{
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expires),
		},
	})
	if erro != nil {
		logger.Warning("Failed to create token: %s", erro)
		response.WriteError(errors.InternalError(), w, r)
	}

	// Set cookie
	cookie := http.Cookie{
		Name:     middleware.CookieName,
		Value:    jwtToken,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,
		Secure:   !api.conf.DevMode,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &cookie)

	response.WriteText("Cookie set", 200, w)
}

func (api *Api) Logout(w http.ResponseWriter, r *http.Request) {
	// Set empty cookie to delete the existing one
	c := &http.Cookie{
		Name:    middleware.CookieName,
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),

		HttpOnly: true,
	}
	http.SetCookie(w, c)

	w.Header().Set("Hx-Refresh", "true")
	response.WriteText("Cookie deleted", 200, w)
}

func (api *Api) ChangeTheme(w http.ResponseWriter, r *http.Request) {

	// Get the new theme
	newThemeVal := strings.ToLower(chi.URLParam(r, "newTheme"))
	newTheme := 0
	switch newThemeVal {
	case "1", "dark", "dunkel":
		newTheme = 1
	case "0", "light", "hell":
		newTheme = 0
	case "switch":
		if api.R().User.DarkTheme == 0 {
			newTheme = 1
		}
	default:
		errors.BadRequest("Invalid theme value provided").Write(w, r)
		return
	}

	// Check if we need to update the theme
	if api.R().User.DarkTheme != newTheme {
		newUser := *api.R().User
		newUser.DarkTheme = newTheme

		if err := api.UpdateProperty(*newUser.User, models.User_DarkTheme); err != nil {
			err.GetErrorStruct().Write(w, r)
		} else {
			w.Header().Add("Hx-Refresh", "true")
			response.WriteText("Theme updated", 200, w)
		}
	} else {
		response.Write(204, w)
	}
}

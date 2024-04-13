package api

import (
	"fmt"
	"io/fs"
	"net/http"
	"regexp"
	"strings"

	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/go-webserver/errors"
	"git.rpjosh.de/RPJosh/go-webserver/frontend"
	"git.rpjosh.de/RPJosh/go-webserver/webserver"
	root "git.rpjosh.de/RPJosh/workout"
	"git.rpjosh.de/RPJosh/workout/internal/api/router"
	"git.rpjosh.de/RPJosh/workout/internal/api/templates"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/internal/translator"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/svg"
)

type errorConfig struct {
	errors.DefaultConfig

	conf *models.AppConfig
}

type dep struct {
	conf *models.AppConfig
}

// Routes setups the WebServer with all routes of the App
func Routes(server *webserver.WebServer[*models.AppConfig]) http.Handler {
	router := chi.NewRouter()
	d := dep{conf: server.Dependency}
	router.Use(middleware.RealIP, server.RecoverPanic, server.LogRequest, server.SecureHeaders, d.redirectMiddleware, d.cacheMiddleware)

	// Setup global error handler
	errors.Config = errorConfig{conf: server.Dependency}

	// Add minifier to save bandwidth
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("image/svg+xml", svg.Minify)
	m.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)
	//router.Use(m.Middleware)

	// Serve static files
	if staticFolder, err := fs.Sub(root.Static, "static"); err != nil {
		logger.Error("Cannot access the embedded directory 'static': %s", err)
	} else {
		frontend.FileServer(router, "/static", http.FS(staticFolder))
	}

	// Setup API / static files endpoints
	api := Api{Config: server.Dependency}
	api.SetupServer(router)

	return router
}

// cacheMiddleware adds a midleware that adds the content type and cache controle
// header to the response
func (d *dep) cacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Don't cache anything for dev mode
		if d.conf.DevMode {
			next.ServeHTTP(w, r)
			return
		}

		// Cache images for 20 days. These should not change in general
		if strings.HasPrefix(r.URL.Path, "/static/img") || strings.HasPrefix(r.URL.Path, "/docs/assets/img") || strings.HasPrefix(r.URL.Path, "/docs/assets/images") {
			w.Header().Set("Cache-Control", "max-age=1728000")
		}

		// Cache documentation javascript files for 20 days
		if strings.HasPrefix(r.URL.Path, "/docs/assets/js") || strings.HasPrefix(r.URL.Path, "/docs/assets/javascript") {
			w.Header().Set("Cache-Control", "max-age=1728000")
			w.Header().Set("Content-Type", "text/javascript")
		}

		// Static CSS files CAN change. But because these are only global defaults,
		// it is not important that the user is using an up-to-date version.
		// Cache time of 5 day
		if strings.HasPrefix(r.URL.Path, "/static/css") || strings.HasPrefix(r.URL.Path, "/docs/assets/stylesheets") {
			w.Header().Set("Cache-Control", "max-age=432000")
			w.Header().Set("Content-Type", "text/css")
		}

		// Custom javascript functions are important to work. Cache them only
		// for a session period of 4 hours
		if strings.HasPrefix(r.URL.Path, "/static/js") {
			w.Header().Set("Cache-Control", "max-age=14400")
			w.Header().Set("Content-Type", "text/javascript")
		}
		// Third party packages are not changing
		if strings.HasPrefix(r.URL.Path, "/static/js/3d/") {
			w.Header().Set("Cache-Control", "max-age=2592000")
		}

		next.ServeHTTP(w, r)
	})
}

// redirectMiddleware automatically redirects the user to a language specific site
// based on the set cookie
func (d *dep) redirectMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Exclude static folders from redirect
		if strings.HasPrefix(r.URL.Path, "/static") || strings.HasPrefix(r.URL.Path, "/dev") || strings.HasPrefix(r.URL.Path, "/docs") {
			next.ServeHTTP(w, r)
			return
		}

		// Read cookie
		c, _ := r.Cookie("Language")
		if c != nil {
			if lang, err := translator.GetLanguageByString(c.Value); err != nil {
				logger.Debug("%s", err)
				next.ServeHTTP(w, r)
			} else if !strings.HasPrefix(r.URL.Path, "/"+lang.String()) {
				// Redirect to matching language
				newPath := "/" + lang.String() + templates.ReplaceLanguageFromPath(r.URL.Path)
				w.Header().Set("Location", newPath)
				w.WriteHeader(302)
			} else {
				next.ServeHTTP(w, r)
			}
		} else if acceptLang := r.Header.Get("Accept-Language"); acceptLang != "" {
			// There is no cookie set → the user doesn't have a preference.
			// We will redirect based on the "Accept-Language" header
			if strings.HasPrefix(acceptLang, "de") {
				// Redirect when url doesn't have a "/de" prefix
				if !strings.HasPrefix(r.URL.Path, "/de") {
					newPath := "/de" + templates.ReplaceLanguageFromPath(r.URL.Path)
					w.Header().Set("Location", newPath)
					w.WriteHeader(302)
				}
			} else if strings.HasPrefix(acceptLang, "en") {
				// Redirect when url doesn't have a "/en" prefix
				if !strings.HasPrefix(r.URL.Path, "/en") {
					newPath := "/en" + templates.ReplaceLanguageFromPath(r.URL.Path)
					w.Header().Set("Location", newPath)
					w.WriteHeader(302)
				}
			}

			next.ServeHTTP(w, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func (e errorConfig) HandlePanic(err any, trace string, w http.ResponseWriter, r *http.Request) {

	// Get the "real" error
	var concreteErr error
	if eb, ok := err.(error); ok {
		concreteErr = eb
	}
	if val, ok := err.(string); ok {
		concreteErr = fmt.Errorf("%s", val)
	}

	// Render error if error was provided
	if concreteErr != nil {
		// Log error cause
		logger.Error("Error: %s", concreteErr)

		// Write debug trace
		logger.Debug(trace)

		// Initialize templ for error rendering
		req := router.NewApiRequest(r, w, router.Route{Name: "ErrorHandler"})
		tmpl := templates.NewTemplates(&req.R().Tr, e.conf, w, r, req.R().Comp)

		// Render error page
		tmpl.CheckError(concreteErr)
	} else {
		// Fallback to default
		e.DefaultConfig.HandlePanic(err, trace, w, r)
	}

}

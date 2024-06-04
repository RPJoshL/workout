package api

import (
	"fmt"
	"io/fs"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"git.rpjosh.de/RPJosh/go-logger"
	root "git.rpjosh.de/RPJosh/workout"
	"git.rpjosh.de/RPJosh/workout/internal/api/components"
	rmiddleware "git.rpjosh.de/RPJosh/workout/internal/api/middleware"
	"git.rpjosh.de/RPJosh/workout/internal/api/router"
	"git.rpjosh.de/RPJosh/workout/internal/api/templates"
	errPage "git.rpjosh.de/RPJosh/workout/internal/api/templates/err"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/internal/translator"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
	"git.rpjosh.de/RPJosh/workout/pkg/response"
	"git.rpjosh.de/RPJosh/workout/pkg/webserver"
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
	router.Use(server.RequestId, rmiddleware.LanguageMiddleware, middleware.RealIP, server.RecoverPanic, server.SecureHeaders, d.cacheMiddleware)

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
		webserver.FileServer(router, "/static", http.FS(staticFolder))
	}

	// Setup API / static files endpoints
	api := Api{Config: server.Dependency}
	api.SetupServer(router)

	return router
}

// Write transforms translations key into their full reference
func (e errorConfig) Write(err errors.ErrorResponse, writer http.ResponseWriter, r *http.Request) {
	message := err.Message

	if strings.HasPrefix(message, "#") {

		// Get language from context we set previously
		langStr := "en"
		lang := r.Context().Value(models.KeyLanguage)
		if langS, ok := lang.(string); ok {
			langStr = langS
		}

		t := translator.NewTranslator()
		t.Language, _ = translator.GetLanguageByString(langStr)
		// Apply sprintf correctly (with translator)
		message = err.ApplySprintf(t).Message
	}

	response.WriteText(message, err.Status, writer)
}

func (e errorConfig) HandlePanic(err any, trace string, w http.ResponseWriter, r *http.Request) {

	// Get translator
	langStr := "en"
	lang := r.Context().Value(models.KeyLanguage)
	if langS, ok := lang.(string); ok {
		langStr = langS
	}
	t := *router.GlobalTranslator
	t.Language, _ = translator.GetLanguageByString(langStr)

	// We need a templ instance to render error page
	tmpl := templates.NewTemplates(&t, e.conf, w, r, components.NewComponents(&t), nil)
	ePage := errPage.Err{T: &t, Render: tmpl.Render, Link: tmpl.Link}

	// Try to parse it to an error response (the error occured in awareness of the developer :)
	if errResponse, ok := err.(errors.ErrorResponse); ok {
		message := errResponse.Message
		if strings.HasPrefix(message, "#") {
			message = t.Get(message[1:])
		}

		ePage.Error(errResponse.Status, message, w)
		//errResponse.Write(w, r)
		return
	}

	// Log error and write header
	logger.Error("Error: %s", fmt.Errorf("%s", err))

	ePage.Error(500, fmt.Sprintf("%s", err), w)
	//w.WriteHeader(500)
	//w.Header().Set("Connection", "close")

	// Write debug trace
	logger.Debug(trace)
}

func (c errorConfig) GetLoggerFromDependendency(dep any) *logger.Logger {
	depRequest, ok := dep.(router.ApiRequestler)
	if !ok {
		logger.Warning("Dependency for [errors.Log()] is not [router.ApiRequestler]. Got %q", reflect.TypeOf(dep))
		return logger.GetGlobalLogger()
	}

	return depRequest.R().Logger
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

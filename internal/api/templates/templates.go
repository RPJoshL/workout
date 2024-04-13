package templates

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"net/http"
	"regexp"
	"runtime"
	"strings"

	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/go-webserver/errors"
	"git.rpjosh.de/RPJosh/workout/internal/api/components"
	errPage "git.rpjosh.de/RPJosh/workout/internal/api/templates/err"
	"git.rpjosh.de/RPJosh/workout/internal/database"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/internal/translator"
	"github.com/a-h/templ"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/svg"
)

// Templates contains generic components that are used across this site
// and generic function
type Templates struct {
	translator *translator.Translator
	config     *models.AppConfig
	w          http.ResponseWriter
	r          *http.Request

	comp *components.Components
}

func NewTemplates(tr *translator.Translator, config *models.AppConfig, w http.ResponseWriter, r *http.Request, comp *components.Components) *Templates {
	return &Templates{
		translator: tr,
		config:     config,
		w:          w,
		r:          r,
		comp:       comp,
	}
}

// Render renders the given component into the main layout of the site.
// You have to provide the ID of the translation key for the title and the description
// of the current page. This is used for SEO optimizations.
// All CSS files that are parents or inside of the folder of the calling file are added as class files.
func (t *Templates) Render(component templ.Component, title, description string) {
	t.r.Header.Set("Content-Type", "text/html")

	// Minify the response to save bandwidth
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("image/svg+xml", svg.Minify)
	m.Add("text/html", &html.Minifier{
		KeepDefaultAttrVals: true,
	})
	m.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)
	mw := m.Writer("text/html", t.w)

	// Get unique CSS identifier based on path
	className := "random"
	_, file, _, ok := runtime.Caller(1)
	if ok {
		className = ""

		// Get all containing folders to add these as class names for the div (hashed)
		packageName := strings.Join(strings.Split(file, "/internal/")[1:], "/")

		lastSlash := 0
		for i := 0; i < strings.Count(packageName, "/"); i++ {
			// Get the index of the next "/"
			nextSlash := strings.Index(packageName[lastSlash:], "/") + lastSlash
			lastSlash = nextSlash + 1

			// Hash the file name and add it as a class name
			hashContent := packageName[0:nextSlash]
			h := sha1.New()
			h.Write([]byte(hashContent))
			hash := hex.EncodeToString(h.Sum(nil))[0:16]
			className += " col-" + hash

			logger.Debug("Hashing %q: %s", hashContent, hash)
		}
	}

	// Render the component into the main layout
	t.Layout(title, description).Render(templ.WithChildren(t.r.Context(), t.wrapWithSpan(className, component)), mw)

	if err := mw.Close(); err != nil {
		logger.Warning("Failed to close minify writer: %s", err)
	}
}

func (t *Templates) wrapWithSpan(className string, component templ.Component) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		if _, err := io.WriteString(w, `<span style="display: inline-block; width: 100%; height: 100%; box-sizing: border-box;" class="`+className+`">`); err != nil {
			return err
		}
		if err := component.Render(ctx, w); err != nil {
			return err
		}

		_, err := io.WriteString(w, `</span>`)

		return err
	})
}

// ReplaceLanguageFromPath removes any present language prefix from the url
func ReplaceLanguageFromPath(path string) string {
	languages := []string{"/en", "/de"}

	for _, l := range languages {
		if strings.HasPrefix(path, l) {
			path = strings.TrimPrefix(path, l)
			break
		}
	}

	// Add leading '/'
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return path
}

// link returns a relative link to the given target like '/download/'
func (t *Templates) Link(target string) templ.SafeURL {

	// The link doesn't contain any language and the target is also 'en'
	// => don't change the URL
	if t.translator.Language == translator.English && t.r.URL.Path == ReplaceLanguageFromPath(t.r.URL.Path) {
		return templ.SafeURL(target)
	}

	return templ.SafeURL("/" + t.translator.Language.String() + target)
}

func (t *Templates) CheckError(err error) bool {
	// Nothing to do
	if err == nil {
		return false
	}

	// Get error page
	errPage := errPage.Err{T: t.translator, Render: t.Render, Link: t.Link}

	// Try to cast it to a database error
	if dbError, ok := err.(database.DatabaseError); ok {
		errPage.Error(dbError.GetResponse().Status, dbError.GetResponse().Message)
	} else if rpError, ok := err.(errors.ErrorResponse); ok {
		errPage.Error(rpError.Status, rpError.Message)
	} else {
		errPage.Error(500, t.translator.Get("error.internal"))
	}

	// Print warning
	logger.Warning("Request for \"%s %s\" failed: %s", t.r.Method, t.r.URL, err)

	return true
}

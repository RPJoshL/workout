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
	user       *models.User

	comp *components.Components
}

func NewTemplates(tr *translator.Translator, config *models.AppConfig, w http.ResponseWriter, r *http.Request, comp *components.Components, user *models.User) *Templates {
	return &Templates{
		translator: tr,
		config:     config,
		w:          w,
		r:          r,
		comp:       comp,
		user:       user,
	}
}

// Render renders the given component into the main layout of the site.
// You have to provide the ID of the translation key for the title and the description
// of the current page. This is used for SEO optimizations.
// All CSS files that are parents or inside of the folder of the calling file are added as class files.
func (t *Templates) Render(component templ.Component, title, description string) {
	t.r.Header.Set("Content-Type", "text/html")

	// Get css files
	mw, className := t.getCss()

	// Don't return the main layout if content is swapped
	swapHeader := t.r.Header.Get("Hx-target")
	if swapHeader == "content" {
		// Update browser history to the requested path
		t.w.Header().Set("HX-Push-Url", t.r.URL.Path)
		t.Content().Render(templ.WithChildren(t.r.Context(), t.wrapWithSpan(className, component)), mw)
	} else {
		// Render the component into the main layout
		t.Layout(title, description, true, t.modal()).Render(templ.WithChildren(t.r.Context(), t.wrapWithSpan(className, component)), mw)
	}

	if err := mw.Close(); err != nil {
		logger.Warning("Failed to close minify writer: %s", err)
	}
}

// RenderWithoutLayout renders the given component into the main site WITHOUT the layout
// (header and footer)
func (t *Templates) RenderWithoutLayout(component templ.Component, title, description string) {
	t.r.Header.Set("Content-Type", "text/html")

	// Get css files
	mw, className := t.getCss()

	// Render the component into the main layout
	t.Layout(title, description, false, t.modal()).Render(templ.WithChildren(t.r.Context(), t.wrapWithSpan(className, component)), mw)

	if err := mw.Close(); err != nil {
		logger.Warning("Failed to close minify writer: %s", err)
	}
}

// RenderModal renders a single modal.
// You have to provide the default component and path that should be rendered below
// the modal if the user uses the absolute path of the modal
func (t *Templates) RenderModal(modal templ.Component, modalTitle string, def templ.Component, defPath, title, description string) {
	t.r.Header.Set("Content-Type", "text/html")

	// Get css files
	mw, className := t.getCss()

	// Don't return the main layout if content is swapped
	swapHeader := t.r.Header.Get("Hx-target")
	logger.Debug("Received swap target %q", swapHeader)
	if swapHeader == "modal-content" {
		// Update browser history to the requested path
		t.w.Header().Set("HX-Push-Url", t.r.URL.Path)
		t.modalVisible().Render(templ.WithChildren(t.r.Context(), t.wrapWithSpan(className, modal)), mw)
	} else {
		m := t.wrapWithChilds(t.modalWithData("true", defPath, t.translator.Get(modalTitle)), modal)
		t.Layout(title, description, true, m).Render(templ.WithChildren(t.r.Context(), t.wrapWithSpan(className, def)), mw)
	}

	if err := mw.Close(); err != nil {
		logger.Warning("Failed to close minify writer: %s", err)
	}
}

// getCss returns all CSS filenames to apply for a template of the invoking method.
// It does also return a writer that minifies html and css ressources
// when writing the template to
func (t *Templates) getCss() (writer io.WriteCloser, className string) {
	// Minify the response to save bandwidth
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("image/svg+xml", svg.Minify)
	m.Add("text/html", &html.Minifier{
		KeepDefaultAttrVals: true,
	})
	m.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)
	writer = m.Writer("text/html", t.w)

	// Get unique CSS identifier based on path
	className = "random"
	_, file, _, ok := runtime.Caller(2)
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

	return
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

func (t *Templates) wrapWithChilds(root templ.Component, child templ.Component) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		return root.Render(templ.WithChildren(ctx, child), w)
	})
}

// link returns a relative link to the given target like '/download/'
func (t *Templates) Link(target string) templ.SafeURL {
	return templ.SafeURL("/" + strings.TrimPrefix(target, "/"))
}

// CheckError checks if the error is not nil and
// renders an error page to the user.
// This function has to be called BEFORE any render function
// wrote to the http.response.
//
// Use [DisplayError] for errors that ocurred during an API request
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
	logger.Info("Request for \"%s %s\" failed: %s", t.r.Method, t.r.URL, err)

	return true
}

func (t *Templates) DisplayError(err error) bool {
	// Nothing to do
	if err == nil {
		return false
	}

	// @TODO display a popup or something like that!

	// Print warning
	logger.Info("API Request for \"%s %s\" failed: %s", t.r.Method, t.r.URL, err)

	return true
}

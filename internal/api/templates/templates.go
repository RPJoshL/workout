package templates

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"net/http"
	"regexp"
	"runtime"
	"strings"

	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/workout/internal/api/components"
	errpage "git.rpjosh.de/RPJosh/workout/internal/api/templates/err"
	"git.rpjosh.de/RPJosh/workout/internal/api/templates/header"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/internal/translator"
	"git.rpjosh.de/RPJosh/workout/pkg/database"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
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
	user       *models.WebUser

	comp   *components.Components
	header *header.Header
}

func NewTemplates(tr *translator.Translator, config *models.AppConfig, w http.ResponseWriter, r *http.Request, comp *components.Components, user *models.WebUser) *Templates {
	return &Templates{
		translator: tr,
		config:     config,
		w:          w,
		r:          r,
		comp:       comp,
		user:       user,
		header:     header.NewHeader(user, tr, comp, r),
	}
}

// Render renders the given component into the main layout of the site.
// You have to provide the ID of the translation key for the title and the description
// of the current page. This is used for SEO optimizations.
// All CSS files that are parents or inside the folder of the calling file, are added as class files.
func (t *Templates) Render(component templ.Component, title, description string) {
	t.r.Header.Set("Content-Type", "text/html")

	// Get CSS files
	mw, className := t.getCss()

	// Don't return the main layout if content is swapped
	swapHeader := t.r.Header.Get("Hx-Target")
	var err error
	if swapHeader == "content" {
		// Update browser history to the requested path
		t.w.Header().Set("Hx-Push-Url", t.r.URL.Path)
		err = t.Content().Render(templ.WithChildren(t.r.Context(), t.wrapWithSpan(className, component)), mw)
	} else {
		// Render the component into the main layout
		err = t.Layout(title, description, true, t.modal()).Render(templ.WithChildren(t.r.Context(), t.wrapWithSpan(className+" content-wrapper-main", component)), mw)
	}

	if err != nil {
		logger.Debug("Failed to render content / layout: %s", err)
	}
	if err := mw.Close(); err != nil {
		logger.Warning("Failed to close minify writer: %s", err)
	}
}

// RenderWithoutLayout renders the given component into the main site WITHOUT the layout
// (header and footer)
func (t *Templates) RenderWithoutLayout(component templ.Component, title, description string) {
	t.r.Header.Set("Content-Type", "text/html")

	// Get CSS files
	mw, className := t.getCss()

	// Render the component into the main layout
	if err := t.Layout(title, description, false, t.modal()).Render(templ.WithChildren(t.r.Context(), t.wrapWithSpan(className, component)), mw); err != nil {
		logger.Debug("Failed to render layout: %s", err)
	}

	if err := mw.Close(); err != nil {
		logger.Warning("Failed to close minify writer: %s", err)
	}
}

// RenderDirect returns the HTML element directly without any wrapper
func (t *Templates) RenderDirect(component templ.Component) {
	t.r.Header.Set("Content-Type", "text/html")

	// Get writer
	mw, _ := t.getCss()

	// Render the component into the main layout
	if err := component.Render(t.r.Context(), mw); err != nil {
		logger.Debug("Failed to render component: %s", err)
	}

	if err := mw.Close(); err != nil {
		logger.Warning("Failed to close minify writer: %s", err)
	}
}

// RenderModal renders a single modal.
// You have to provide the default component and path that should be rendered below
// the modal if the user uses the absolute path of the modal.
//
// Note: both components HAS TO BE in the same package from where you are calling this to apply classes
// correctly.
// For rendering BASE components with a different path, specify a correct "rootLayoutClass", which should contain
// a file / import path generated with [utils.GetCallerFile()]. This is optional
func (t *Templates) RenderModal(modal templ.Component, modalTitle string, def templ.Component, defPath, title, description, rootLayoutClass string) {
	t.r.Header.Set("Content-Type", "text/html")

	// Get CSS files
	mw, className := t.getCss()

	// Don't return the main layout if content is swapped
	swapHeader := t.r.Header.Get("Hx-Target")
	logger.Trace("Received swap target %q", swapHeader)
	if swapHeader == "modal-content" {
		// Update browser history to the requested path
		t.w.Header().Set("Hx-Push-Url", t.r.URL.Path)
		_ = t.modalVisible(className).Render(templ.WithChildren(t.r.Context(), t.wrapWithSpan(className, modal)), mw)
	} else {
		m := t.wrapWithChilds(t.modalWithData("true", defPath, t.translator.Get(modalTitle), className), modal)

		// If the main layout is rendered in another package, adjust class name of layout
		layoutClass := className
		if rootLayoutClass != "" {
			layoutClass = getCssClassNames(rootLayoutClass)
		}

		_ = t.Layout(title, description, true, m).Render(templ.WithChildren(t.r.Context(), t.wrapWithSpan(layoutClass, def)), mw)
	}

	if err := mw.Close(); err != nil {
		logger.Warning("Failed to close minify writer: %s", err)
	}
}

// getCss returns all CSS filenames to apply for a template of the invoking method.
// It does also return a writer that minifies HTML and CSS ressources
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
		className = getCssClassNames(file)
	}

	// Add theme class
	if t.user != nil {
		if t.user.DarkTheme == 1 {
			className += " theme-cust-dark"
		} else {
			className += " theme-cust-light"
		}
	}

	return
}

func getCssClassNames(file string) string {
	// Get all containing folders to add these as class names for the div (hashed)
	packageName := strings.Join(strings.Split(file, "/internal/")[1:], "/")

	lastSlash := 0
	var className strings.Builder
	for range strings.Count(packageName, "/") {
		// Get the index of the next "/"
		nextSlash := strings.Index(packageName[lastSlash:], "/") + lastSlash
		lastSlash = nextSlash + 1

		// Hash the file name and add it as a class name
		hashContent := packageName[0:nextSlash]
		h := sha1.New()
		h.Write([]byte(hashContent))
		hash := hex.EncodeToString(h.Sum(nil))[0:16]
		className.WriteString(" col-" + hash)

		logger.Trace("Hashing %q: %s", hashContent, hash)
	}

	return className.String()
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

func (t *Templates) wrapWithChilds(root, child templ.Component) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		return root.Render(templ.WithChildren(ctx, child), w)
	})
}

// Link returns a relative link to the given target like '/download/'
func (t *Templates) Link(target string) templ.SafeURL {
	return templ.SafeURL("/" + strings.TrimPrefix(target, "/"))
}

// CheckError checks if the error is not nil and
// renders an error page to the user.
// This function has to be called BEFORE any render function
// wrote to the http.response.
//
// Use [DisplayError] for errors that occurred during an API request
func (t *Templates) CheckError(err error) bool {
	// Nothing to do
	if err == nil {
		return false
	}

	// Get error page
	errPage := errpage.Err{T: t.translator, Render: t.Render, Link: t.Link}

	// Try to cast it to a database error
	if dbError, ok := errors.GetAs[database.Error](err); ok {
		errPage.Error(dbError.GetResponse().Status, dbError.GetResponse().Message, t.w)
	} else if rpError, ok := errors.GetAs[errors.ErrorResponse](err); ok {
		errPage.Error(rpError.Status, rpError.Message, t.w)
	} else {
		errPage.Error(500, t.translator.Get("error.internal"), t.w)
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

// GetHtml returns the RAW html code for the provided component
func (t *Templates) GetHtml(comp templ.Component) string {
	buf := new(bytes.Buffer)
	if err := comp.Render(context.Background(), buf); err != nil {
		logger.Warning("Failed to write HTML content into dummy buffer")
		return ""
	}

	return buf.String()
}

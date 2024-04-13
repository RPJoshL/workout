package err

import (
	"git.rpjosh.de/RPJosh/workout/internal/translator"
	"github.com/a-h/templ"
)

type Err struct {
	T *translator.Translator

	// Render function of "tmpl" to avoid an import cyclce
	Render func(component templ.Component, title string, description string)

	// Link returns a relative link to the given target like '/download/'.
	// See tmpl.Link for more infos. This cannot be used directly to avoid an
	// import cycle
	Link func(target string) templ.SafeURL
}

func (e *Err) Error(code int, description string) {
	e.Render(e.error(code, description), "error.title", "error.description")
}

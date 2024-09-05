package header

import (
	"net/http"
	"strings"

	"git.rpjosh.de/RPJosh/workout/internal/api/components"
	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/internal/translator"
	"github.com/a-h/templ"
)

// Header renders the top side header of the site in
// the various scenes (Main, Settings)
type Header struct {
	user       *models.WebUser
	translator *translator.Translator
	comp       *components.Components
	request    *http.Request
}

func NewHeader(user *models.WebUser, trans *translator.Translator, comp *components.Components, request *http.Request) *Header {
	return &Header{
		user:       user,
		translator: trans,
		comp:       comp,
		request:    request,
	}
}

// GetHeader returns the matching header for the requested URL.
// For custom styling you can provide a "context" that indicates which tab
// is displayed. You can use that for custom styling
func (h *Header) GetHeader(context string) templ.Component {
	if strings.HasPrefix(h.request.URL.Path, "/settings") {
		return h.settingsHeader(context)
	} else {
		return h.mainHeader(context)
	}
}

package markdown

import (
	icons "git.rpjosh.de/RPJosh/workout/internal/api/components/icon"
	"git.rpjosh.de/RPJosh/workout/pkg/utils"
)

// Markdown is a live editor and previewer for markdown
// styled text powered by EasyMDE
type Markdown struct {
	Icons *icons.Icons
}

// Options contains generic options to customize
// the easymde instance
type Options struct {

	// Weather this markdown field is read only
	ReadOnly bool

	// Name is used inside a form as a key for the markdown
	// editors content
	Name string

	// Internal ID of the form
	formId string
	// Internal ID of the preview toggle button
	previewToggleId string
}

func (o *Options) getFormId() string {
	if o.formId == "" {
		o.formId, _ = utils.GenerateRandomString(12)
		o.formId = "o" + o.formId
	}

	return o.formId
}

func (o *Options) getPreviewToggleId() string {
	if o.previewToggleId == "" {
		o.previewToggleId, _ = utils.GenerateRandomString(12)
		o.previewToggleId = "o" + o.previewToggleId
	}

	return o.previewToggleId
}

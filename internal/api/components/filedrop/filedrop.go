package filedrop

import (
	icons "git.rpjosh.de/RPJosh/workout/internal/api/components/icon"
	"git.rpjosh.de/RPJosh/workout/pkg/utils"
)

// FileDrop creates file drop element in javascript
type FileDrop struct {
	Icons *icons.Icons
}

type Options struct {

	// Name of the input element for forms
	Name string

	// Text to display if no file is uploaded yet
	NoFileText string

	// Text to display if a file was already uploaded.
	// The text "{{file}}" is replaced with the filename (if present)
	FileText string

	// Internal ID of the root drop field
	id string
}

func (o *Options) getId() string {
	if o.id == "" {
		o.id, _ = utils.GenerateRandomString(12)
		o.id = "o" + o.id
	}

	return o.id
}

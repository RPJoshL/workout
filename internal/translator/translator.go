package translator

import (
	"embed"
	"fmt"
	"reflect"
	"strings"

	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/workout/i18n"
	"git.rpjosh.de/RPJosh/workout/pkg/errors"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"gopkg.in/yaml.v3"
)

// Language represents a language that is supported from this application
type Language int

const (
	English Language = 0
	German           = iota
)

// String returns a two digit code stating the language
func (l Language) String() string {
	switch l {
	case English:
		return "en"
	case German:
		return "de"
	}

	return "??"
}

// Language returns the language for the text pacakge
func (l Language) Language() language.Tag {
	switch l {
	case English:
		return language.English
	case German:
		return language.German
	}

	return language.English
}

// GetLanguageByString returns the langauge identified
// by the provided string
func GetLanguageByString(val string) (Language, error) {
	switch strings.ToLower(val) {
	case "en", "english":
		return English, nil
	case "de", "deutsch", "german":
		return German, nil
	default:
		return English, errors.BadRequest("Language not found")
	}
}

// Translator
type Translator struct {

	// Prefered language to use for translations
	Language Language

	// Internal representation of all translations
	de *map[string]string
	en *map[string]string
}

// NewTranslator creates a new instance of a translator by reading the
// and parsing all available translation files
func NewTranslator() *Translator {
	rtc := &Translator{}

	// Parse the yaml files and store them in memory for faster access time
	rtc.de = parseFile(i18n.German, "de/index.yaml")
	rtc.en = parseFile(i18n.English, "en/index.yaml")

	return rtc
}

// Get returns the translation value for the provided key
func (t *Translator) Get(key string) string {
	if key == "" {
		return ""
	}

	// Get german translation
	if t.Language == German {
		if val, exists := (*t.de)[key]; exists {
			return val
		}
	}

	// Print a debug warning if a key for the provided value could not be found
	if t.Language != English {
		logger.Debug("Didn't found a translation for key %q and language %q", key, t.Language)
	}

	if val, exists := (*t.en)[key]; exists {
		return val
	} else {
		logger.Warning("Found no value for translation key %q", key)
		return key
	}
}

// Getf returns the translation value that expends to an expression
// for "fmt.Sprintf". All placeholders are replaced with the provided values
func (t *Translator) Getf(key string, values ...any) string {
	val := t.Get(key)

	// Format string
	return fmt.Sprintf(val, values...)
}

// Sprintf formats the given expression like "fmt.Sprintf" but with localized
// formatting enabled
func (t *Translator) Sprintf(str string, values ...any) string {
	// Get localized printer
	p := message.NewPrinter(t.Language.Language())

	return p.Sprintf(str, values...)
}

// parseFile parses the given file from the embedded file system
// and returns it contents as a map 'key1.key2' → value
func parseFile(fs embed.FS, path string) *map[string]string {
	file, err := fs.ReadFile(path)
	if err != nil {
		logger.Error("Failed to read file from ebedded file system: %s", err)
		return &map[string]string{}
	}

	// Unmarshal the config
	var parsed map[string]interface{}
	if err := yaml.Unmarshal(file, &parsed); err != nil {
		logger.Error("Failed to parse yaml file: %s", err)
		return &map[string]string{}
	}

	// Flat the yaml file that we can access properties via 'key1.key2'
	rtc := &map[string]string{}
	flatenYaml("", parsed, rtc)

	return rtc
}

// flatenYaml flattes the given, unmarsheld yaml file to a map that can be accessed
// via 'key1.key2'
func flatenYaml(prevKey string, in map[string]interface{}, out *map[string]string) {
	for k, v := range in {

		// Additional hierarchie level
		if m, ok := v.(map[string]any); ok {
			flatenYaml(prevKey+k+".", m, out)
		} else if str, ok := v.(string); ok {
			// String value → add it to the map
			(*out)[prevKey+k] = str
		} else {
			logger.Warning("Received invalid value while flatten yaml: %s", reflect.TypeOf(v))
		}
	}
}

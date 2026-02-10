package translator

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"

	"git.rpjosh.de/RPJosh/go-logger"
	"github.com/a-h/templ"
)

// WithComponents returns a translation with custom components inside.
// The translation can contain a string like '<c1>That's my text</c1>' that's embedded
// into the given child components and the value '#text'.
// Like: '<c1> <span class="myClass"> #text# </span></c1>
func (t *Translator) WithComponents(key string) templ.Component {
	return t.WithComponentsFull(key, true)
}

// WithComponentsFull is a more configurable way for the function "WithComponents()"
// for translation keys that don't contain any custom tags
func (t *Translator) WithComponentsFull(key string, printWarningIfNotUsed bool) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, writer io.Writer) (err error) {
		ctx = templ.InitializeContext(ctx)

		// Get childrens and write them to string buffer
		childBuff := new(bytes.Buffer)
		if err := templ.GetChildren(ctx).Render(ctx, childBuff); err != nil {
			return err
		}
		childString := childBuff.String()

		// Loop through translation string
		translation := t.Get(key)
		counter := 1
		lastStep := 0
		tagRegex := regexp.MustCompile(`<c(\d+)>`)
		for {
			// Find and extract next number
			rtc := tagRegex.FindStringSubmatch(translation[lastStep:])
			next := math.MaxInt32
			if len(rtc) >= 2 {
				nextInt, err := strconv.Atoi(rtc[1])
				if err != nil {
					logger.Warning("Failed to convert %q to an int", rtc[1])
				} else {
					next = nextInt
				}
			}

			// Get open and closing tag
			openTag := fmt.Sprintf("<c%d>", next)
			closeTag := fmt.Sprintf("</c%d>", next)

			// Find 'cX' within translation string
			if strings.Contains(translation, openTag) && strings.Contains(translation, closeTag) {
				openIndex := strings.LastIndex(translation, openTag)
				closeIndex := strings.LastIndex(translation, closeTag)

				// Write string until open tag
				if openIndex > lastStep {
					if _, err := io.WriteString(writer, escapeString(translation[lastStep:openIndex])); err != nil {
						return err
					}
				}

				// Extract translation between tag
				trans := translation[openIndex+len(openTag) : closeIndex]

				// Get indexex in childs
				openIndexChild := strings.LastIndex(childString, openTag)
				closeIndexChild := strings.LastIndex(childString, closeTag)
				if openIndexChild == -1 || closeIndexChild == -1 || closeIndexChild <= openIndexChild {
					logger.Warning("Found insufficant childs in provided component for translation %q (%d, %d)", key, openIndexChild, closeIndexChild)
					break
				}
				childString := childString[openIndexChild+len(openTag) : closeIndexChild]
				// Replace '#text' with translation context
				if !strings.Contains(childString, "#text#") {
					logger.Warning("No '#text' found in childs for translation %q (%d)", key, counter)
					break
				}
				if _, err := io.WriteString(writer, strings.ReplaceAll(childString, "#text#", trans)); err != nil {
					return err
				}

				// Increment last step
				lastStep = closeIndex + len(closeTag)
			} else {
				// Write the rest of the string
				if _, err := io.WriteString(writer, escapeString(translation[lastStep:])); err != nil {
					return err
				}

				break
			}

			counter++
		}

		if counter == 1 && printWarningIfNotUsed {
			logger.Warning("Called translation function withComponent(%q) but didn't found tags inside", key)
		}

		return
	})
}

// escapeString escapes all html special characters within the given string.
// Only the '<br>' tag will be kept present
func escapeString(str string) string {
	rtc := templ.EscapeString(str)

	// Replace escaped <br> with correct <br>
	rtc = strings.ReplaceAll(rtc, "&lt;br&gt;", "<br>")
	// Replace inline formattings (italic and bright)
	rtc = strings.ReplaceAll(rtc, "&lt;i&gt;", "<i>")
	rtc = strings.ReplaceAll(rtc, "&lt;/i&gt;", "</i>")
	rtc = strings.ReplaceAll(rtc, "&lt;b&gt;", "<b>")
	rtc = strings.ReplaceAll(rtc, "&lt;/b&gt;", "</b>")

	return rtc
}

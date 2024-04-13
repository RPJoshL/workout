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
func (t *Translator) WithComponentsFull(key string, printWarningIfNotUsed bool) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, writer io.Writer) (err error) {

		// Initialize Buffer and context
		buff, isBuffer := writer.(*bytes.Buffer)
		if !isBuffer {
			buff = templ.GetBuffer()
			defer templ.ReleaseBuffer(buff)
		}
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
			open := fmt.Sprintf("<c%d>", next)
			close := fmt.Sprintf("</c%d>", next)

			// Find 'cX' within translation string
			if strings.Contains(translation, open) && strings.Contains(translation, close) {
				openIndex := strings.LastIndex(translation, open)
				closeIndex := strings.LastIndex(translation, close)

				// Write string until open tag
				if openIndex > lastStep {
					buff.WriteString(escapeString(translation[lastStep:openIndex]))
				}

				// Extract translation between tag
				trans := translation[openIndex+len(open) : closeIndex]

				// Get indexex in childs
				openIndexChild := strings.LastIndex(childString, open)
				closeIndexChild := strings.LastIndex(childString, close)
				if openIndexChild == -1 || closeIndexChild == -1 || closeIndexChild <= openIndexChild {
					logger.Warning("Found insufficant childs in provided component for translation %q (%d, %d)", key, openIndexChild, closeIndexChild)
					break
				}
				childString := childString[openIndexChild+len(open) : closeIndexChild]
				// Replace '#text' with translation context
				if !strings.Contains(childString, "#text#") {
					logger.Warning("No '#text' found in childs for translation %q (%d)", key, counter)
					break
				}
				buff.WriteString(strings.Replace(childString, "#text#", trans, -1))

				// Increment last step
				lastStep = closeIndex + len(close)
			} else {
				// Write the rest of the string
				buff.WriteString(escapeString(translation[lastStep:]))
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

// escapeString escapes all html special characters withing the given string.
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

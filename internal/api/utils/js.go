package utils

import (
	"context"
	"fmt"
	"io"
	"runtime"
	"strings"

	"github.com/RPJoshL/go-logger"
	"github.com/a-h/templ"
)

// Module calls a JS function from a module that is located inside the same Go module
// on initial render
func Module(functionName string, vals ...any) templ.Component {
	// Get the package name of the invoking function
	callingPackge := "error-1234"
	if _, file, _, ok := runtime.Caller(1); ok {
		// Get the last part of the file as package name
		lastSlash := strings.LastIndex(file, "/")
		callingPackge = file[:lastSlash]
		lastSlash = strings.LastIndex(callingPackge, "/")
		callingPackge = callingPackge[lastSlash+1:]
	} else {
		logger.Warning("Failed to get name of invoking file")
	}

	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		header := fmt.Sprintf(`
<script type="module">
	import { %s } from '/static/js/modules/%s.js'
	%s
</script>
		`, functionName, callingPackge, templ.SafeScriptInline(functionName, vals...))

		if _, err := io.WriteString(w, header); err != nil {
			return err
		}

		return nil
	})
}

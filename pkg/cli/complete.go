package cli

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"git.rpjosh.de/RPJosh/go-logger"
)

// isCompletionSupported validates that the root struct does have a method named
// "EnableAutoComplete" and calls it.
// If no such method is found false will be returned
func isCompletionSupported(root *cliField[any]) bool {
	// Get the field
	method := root.reflectValue.Addr().MethodByName("EnableAutoComplete")
	if !method.IsValid() {
		return false
	}

	// The Auto Complete function should have no parameter
	if method.Type().NumIn() != 0 {
		logger.Warning("EnableAutoComplete() function should have no params!")
		return false
	}

	// Call it
	method.Call([]reflect.Value{})
	root.isCompletion = true

	// Set the global log level to error so that the run isn't interrupted with logging to stdout
	log := logger.GetGlobalLogger()
	log.Level = logger.LevelError
	logger.SetGlobalLogger(log)

	return true
}

// getCompletionFunction returns a reflect.function that can be used to obtain the
// completion results by the tag value.
// If no function was found, the return value will be "zero"
func getCompletionFunction(structField reflect.StructField, structure any) (rtc reflect.Value) {
	// Get the values from the tag
	tag := structField.Tag.Get("completion")
	if tag == "" {
		// Not a relevant field for autocomplete
		return
	}

	// Try to find a method with that name
	method := reflect.ValueOf(structure).MethodByName(tag)
	if !method.IsValid() {
		logger.Warning("Method %q not found for completion tag", tag)
		return
	}

	// Check if arguments and return values are ok
	if method.Type().NumIn() != 2 && method.Type().NumOut() != 1 {
		logger.Warning("Expected exactly two argument and one return value for autocomplete function")
		return
	}

	return method
}

// getCompletionOptionCheckFunction returns a reflect.function that can be used to check
// if a specific option should be provided for the user.
// If not function was found, the return value will be "zero"
func getCompletionOptionCheckFunction(val reflect.Value) (rtc reflect.Value) {
	// Try to find a method with that name
	method := val.MethodByName("CanOptionBeUsedForComplete")
	if !method.IsValid() {
		return
	}

	// Check if arguments and return values are ok
	if method.Type().NumIn() != 1 && method.Type().NumOut() != 1 {
		logger.Warning("Expected exactly one argument and one return value for CanOptionBeUsedForComplete() function")
		return
	}

	return method
}

// printCurrentOptions prints all options that the user does currently have
func printCurrentOptions(entry, _ *cliField[any], usedParams []string, currentInput string) {
	opts := make([]string, 0)

	// Get the help of the root
	help := ""
	if entry.help.IsValid() && entry.help.Type().NumOut() == 1 {
		rtc := entry.help.Call([]reflect.Value{})
		if str, ok := rtc[0].Interface().(string); ok {
			help = str
		}
	}

	// Replace not allowed characters for use in completion
	help = strings.ReplaceAll(help, "\\|", "~~~~****~~~~")
	help = strings.ReplaceAll(help, ":", "_")

outer:
	for i := range entry.chields {
		child := entry.chields[i]

		// Validate that the option wasn't used already
		for _, used := range usedParams {
			if used == entry.longKey+"."+child.longKey {
				continue outer
			}
		}

		// Validate if the fild should be used
		if entry.completionOptionCheck.IsValid() {
			rtc := entry.completionOptionCheck.Call([]reflect.Value{reflect.ValueOf(child.longKey)})
			if shouldContain, ok := rtc[0].Interface().(bool); ok && !shouldContain {
				continue outer
			}
		}

		// Otherwise, add it to the valid options
		if child.longKey != "" {
			// Try to extract the explanation out of the help
			if help != "" {
				regex, err := regexp.Compile(`(?ms)^ *?` + child.longKey + ` .*?\|(?P<Description>.*?(\||\n|$)(^[^\|]*?\n)*)`)
				if err != nil {
					logger.Warning("Failed to compile regex for key %q: %s", child.longKey, err)
				} else {
					matches := regex.FindStringSubmatch(help)
					index := regex.SubexpIndex("Description")
					if index >= len(matches) {
						// Nothing found
						logger.Warning("No matches found four key%q and help %q", child.longKey, help)
						continue
					}
					description := matches[index]

					// Remove empty lines for every option
					descriptions := strings.Split(description, "\n")
					prettified := ""
					for _, desc := range descriptions {
						prettified += " " + strings.Trim(desc, " ")
					}

					// Replace all escaped "|" again
					prettified = strings.ReplaceAll(prettified, "~~~~****~~~~", "|")
					opts = append(opts, fmt.Sprintf("%s\t%s", child.longKey, strings.Trim(prettified, "\n. |")))
					continue
				}
			}

			opts = append(opts, child.longKey)
		}
	}

	printOptionsForAutocomplete(opts, currentInput, false)
}

func printOptionsForAutocomplete(options []string, currentInput string, quote bool) {
	// Determine the length of options that do not begin with a "-"
	rootOptionsCount := 0
	for _, opt := range options {
		if !strings.HasPrefix(opt, "-") {
			rootOptionsCount++
		}
	}

	// Sort the array
	// sort.Strings(options)

	// Print all options. When no "-" was given, also don't show additional options
	for _, opt := range options {
		if strings.HasPrefix(currentInput, "-") || !strings.HasPrefix(opt, "-") || rootOptionsCount == 0 {
			if quote {
				if strings.HasPrefix(currentInput, "\"") {
					fmt.Printf("%q\n", strings.ReplaceAll(opt, " ", ""))
				} else {
					fmt.Printf("%s\n", strings.ReplaceAll(opt, " ", ""))
				}
			} else {
				fmt.Printf("%s\n", opt)
			}
		}
	}
}

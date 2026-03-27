package cli

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/RPJoshL/go-logger"
)

// ParseParams Parses the given command line options into the given structs.
//
// Tag structure: longKey,shortKey,defaultValue,{description},{required?}
//
// The long key has to be unique.
// If no key was given (tag: ','), the hieararchie will be ignored.
//
// If the defaultValue should be "" you can specify "~~".
// If no value should be required (for e.g. "--version"), you can specify
// "~~~". Note that also the setter should have no parameter.
//
// The required field can have multiple meanings and values:
//
//   - \+ -> the field has to be present (+)
//   - - -> the field has not be present (optional)
//   - 1..9        -> the parameter will be matched by position INSTEAD OF the given key.
//     Therefore, these have to stand at the beginning before all key arguments,
//     and can't be a struct (no root)
//   - +var1,+var2 -> the field has to be present, if longKey1 OR longKey2 in the SAME level is present
//
// Only for struct fields that are not a struct by themselves, the value will be parsed as key + value.
//
// If you want to use a custom setter, specify a method "SetVarname(value string): string" as a method of the struct.
// This method will be automatically called, when the key and the value was found. When the string is not empty,
// an error with the given message will be thrown.
// The same thing does also count for Structs. Here a method "SetStructname(val valueRootStruct)" can be supplied, which will be
// called when no more / all values has been parsed. Through the first parameter shared resources can be used.
//
// You can also specify a Help() string Method to print a help if the user provided an invalid key or an unknown one.
// This function should return a string with "|" as delemiters for a description of the method.
//
// If you want to have auto complete support (for bash only at the moment) you have to provide a method named "EnableAutoComplete()" inside your
// root struct that is being called when the program was launched from the autocomplete script.
// There is also a function "CanOptionBeUsedForComplete(longKey string) bool" available, that is being used from the script to toggle whether the option should be
// provided to the user.
// You can also use an additional tag named "completion" that contains the name of the function that can be called to obtain the possible
// values. This function is getting the valueRootStruct and the already inputed value as a first argument and has to return the possible values as []string.
//
// Supported data types to convert automatically (also pointers of these types):
//   - string
//   - int
//   - float
//   - boolean
//   - string[]: As single arguments surrounded by [] ([ "1. Param" "2. Para" ]).
//     The autocomplete function receives '[]string' instead of a single 'string'
//   - int[]:    As a comma separated list within one argument (1,2,3,4)
//
// A return value <= 0 indicates an error
func ParseParams(args []string, structs any) int {
	// The cliFields are constructed in a tree structure. Load all of them
	rootField := cliField[any]{
		isRoot:       true,
		chields:      getFields(structs),
		reflectValue: reflect.ValueOf(structs).Elem(),
	}
	rootField.setupRootField()

	return parse(&rootField, args[1:], &rootField, 0)
}

// Loops through all the arguments and checks if the key is contained by one of
// the child fields. If it's another root field, the function will be called recursively
//
//nolint:gocyclo // @TODO Needs to be refactored
func parse(root *cliField[any], args []string, entry *cliField[any], level int) int {
	var usedParams []string

	pos := 0
	i := 0
	pseudo := 0 // args + 1
	for i < len(args)+1 {
		if i < len(args) {
			// Check for help arguments
			argLower := strings.ToLower(args[i])
			if argLower == "--help" || argLower == "-h" || argLower == "?" {
				// Never print help for completion
				if entry.isCompletion {
					os.Exit(0)
				}

				if root.help.IsValid() {
					root.printHelp("")
				} else if entry.help.IsValid() {
					entry.printHelp("")
				} else {
					fmt.Println("No help available")
				}

				os.Exit(0)
			}

			// Check for autocomplete function call
			if level == 0 && argLower == "__complete" {
				if !isCompletionSupported(root) {
					return -1
				} else {
					i++
					continue
				}
			}

			// Check for positional argument
			nextPositionalField := getNextByPosition(root, pos)
			if nextPositionalField != nil {
				// Try to get autocomplete values
				if entry.isCompletion && nextPositionalField.completionFunction.IsValid() && i+1 == len(args) {
					result := nextPositionalField.completionFunction.Call([]reflect.Value{entry.reflectValue.Addr().Elem().Addr(), reflect.ValueOf(args[i])})
					if results, ok := result[0].Interface().([]string); ok {
						printOptionsForAutocomplete(results, "-", true)
						os.Exit(0)
					} else {
						logger.Warning("Did not receive a string array as result from completion function")
					}
				} else if entry.isCompletion && i+1 == len(args) {
					// Don't try to set values
					os.Exit(0)
				}

				if err := nextPositionalField.setValue(args[i]); err != nil {
					root.printHelp(err.Error())
					return -1
				}

				usedParams = append(usedParams, root.longKey+"."+nextPositionalField.longKey)
				pos++
				i++
				continue
			}
		} else {
			pseudo = -1
		}

		// The user pressed a tab
		if entry.isCompletion && i == len(args)-1 {
			printCurrentOptions(root, entry, usedParams, args[i])
			os.Exit(0)
		}

		var field *cliField[any]
		var found bool

		if pseudo != -1 {
			field, found = getByKey(root, args[i+pseudo], "")
		}

		if !found /*&& i == 0*/ && i < len(args) && level == 0 {
			// @TODO Should we really jump back? At the moment not. This would result into problems for positional parameters (missing -> using the key of someting other)
			// This could be irritating for the user because he thinks - I've defined this key.....
			if level == 0 {
				root.printHelp(fmt.Sprintf("Unknown option '%s'", args[i]))
				return -1
			} else {
				return i
			}
		} else if !found {
			// No matching option → check if all required parameters were met
			for _, f := range root.chields {
				if (f.required || f.requiredPos != 0) && !contains(&usedParams, root.longKey+"."+f.longKey) {
					root.printHelp(fmt.Sprintf("Missing required parameter '%s'", f.longKey))
					return -1
				}

				// set the specified default value
				if f.defaultValue != nil && *f.defaultValue != "~~~" {
					// if the value is a pointer set the default value only when it is nil
					if (f.reflectValue.Kind() != reflect.Ptr || f.reflectValue.IsNil()) && !contains(&usedParams, root.longKey+"."+f.longKey) {
						if err := f.setValue(*f.defaultValue); err != nil {
							root.printHelp(err.Error())
							return -1
						}
					}
				}
			}

			// Validate all required with
			for _, f := range usedParams {
				rFields, ff := getByKey(root, f, root.longKey+".")
				if ff {
					for _, el := range rFields.requiredWith {
						if !contains(&usedParams, root.longKey+"."+el) {
							root.printHelp(fmt.Sprintf("Parameter '%s' does also require '%s'", f, el))
							return -1
						}
					}
				}
			}

			// Never call root setter when in autocomplete mode
			if entry.isCompletion {
				os.Exit(0)
			}

			// Call finish function for struct
			//nolint:staticcheck // False positive
			if root.setter.IsValid() && !(root.disabled || (field != nil && field.disabled)) {
				if root.setter.Type().NumIn() == 0 {
					root.setter.Call([]reflect.Value{})
				} else {
					root.setter.Call([]reflect.Value{entry.reflectValue.Addr().Elem().Addr()})
				}
			}

			return i
		} else {
			if field.isRoot {
				// Root key specified but no more options → error message
				if i >= len(args) {
					root.printHelp(fmt.Sprintf("The option '%s' requires an value", args[i-1]))
					return -1
				}

				usedParams = append(usedParams, root.longKey+"."+field.longKey)
				newLevel := level + 1
				o := parse(field, args[i+1:], entry, newLevel)

				// error occurred
				if o == -1 {
					return -1
				}
				i += 1 + o
			} else {
				// Logic for setting the struct starts

				if field.defaultValue != nil && *field.defaultValue == "~~~" {
					// Only call setter
					if err := field.callSetterWithoutValue(); err != nil {
						root.printHelp(err.Error())
						return -1
					}

					i -= 1
				} else {
					if i+1 >= len(args) {
						root.printHelp(fmt.Sprintf("The option '%s' requires an value", args[i]))
						return -1
					}

					// Assign variables
					var valToSet any = args[i+1]
					var valAutoComplete any = args[i]

					// Logic for string array as input
					if field.reflectValue.Type().Kind() == reflect.Slice && field.reflectValue.Type().Elem().Kind() == reflect.String {
						values := make([]string, 0)

						// When the user specifies a single '[', that means he wants to pass an array
						if args[i+1] == "[" {
							i += 2

							// Loop until we find a single closing ']'.
							// That's much easier than following the root loop structure
							closingFound := false
							for i < len(args) {
								// Closing tag found
								if args[i] == "]" {
									closingFound = true
									i--
									break
								}

								// Check for autocompletion
								if entry.isCompletion && field.completionFunction.IsValid() && i+1 == len(args) {
									closingFound = true
									values = append(values, args[i])
									valAutoComplete = values
									i--
									break
								} else if entry.isCompletion && i == len(args)-1 {
									// Don't return anything (no completion available)
									logger.Debug("No completion")
									os.Exit(0)
								} else {
									values = append(values, args[i])
									i++
								}
							}

							// No closing tag found
							if !closingFound {
								if entry.isCompletion {
									os.Exit(0)
								}

								root.printHelp("Found no closing bracket ']' for array input")
								return -1
							}
						} else if entry.isCompletion && field.completionFunction.IsValid() && i+2 == len(args) {
							// Only change autocomplete value
							valAutoComplete = []string{args[i+1]}
						} else {
							// The user provided no array -> use a single value for the parameter array
							values = append(values, args[i+1])
						}

						valToSet = values
					}

					// Try to get autocomplete values
					if entry.isCompletion && field.completionFunction.IsValid() && i+2 == len(args) {
						result := field.completionFunction.Call([]reflect.Value{entry.reflectValue.Addr().Elem().Addr(), reflect.ValueOf(valAutoComplete)})
						if results, ok := result[0].Interface().([]string); ok {
							printOptionsForAutocomplete(results, "-", true)
							os.Exit(0)
						} else {
							logger.Warning("Did not receive a string array as result from completion function")
						}
					} else if entry.isCompletion && i+1 == len(args)-1 {
						// Don't return anything (no completion available)
						os.Exit(0)
					}

					if err := field.setValue(valToSet); err != nil {
						root.printHelp(err.Error())
						return -1
					}
				}

				usedParams = append(usedParams, root.longKey+"."+field.longKey)
				pos++
				i += 2
			}
		}
	}

	return 1
}

// Converts the given value to the specified type
func convertValue(val string, t reflect.Type) (any, error) {
	switch t.Kind() {
	case reflect.String:
		return val, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val, err := strconv.Atoi(val)
		if err != nil {
			return val, err
		}

		switch t.Kind() {
		case reflect.Int64:
			return int64(val), err
		case reflect.Int32:
			return int32(val), err
		case reflect.Int16:
			return int16(val), err
		}

		return val, err
	case reflect.Float32, reflect.Float64:
		return strconv.ParseFloat(val, 64)
	case reflect.Bool:
		return strconv.ParseBool(val)
	case reflect.Pointer:
		val, err := convertValue(val, t.Elem())
		if err != nil {
			return nil, err
		}

		// &val returns *interface{}..... But why?
		d := reflect.ValueOf(val)
		dPtr := reflect.New(d.Type())
		dPtr.Elem().Set(d)
		return dPtr.Interface(), nil
	case reflect.Slice:
		concreteType := t.Elem()
		rtc := reflect.MakeSlice(t, 0, 0)

		// Loop through every element and convert the value
		if concreteType.Kind() == reflect.String {

		} else if val != "" {
			for i, el := range strings.Split(val, ",") {
				convValue, err := convertValue(el, concreteType)
				if err != nil {
					return nil, fmt.Errorf("invalid value at position %d: %w", i, err)
				}
				rtc = reflect.Append(rtc, reflect.ValueOf(convValue))
			}
		}

		return rtc.Interface(), nil
	default:
		return nil, fmt.Errorf("no supported data type given")
	}
}

func (field *cliField[T]) printHelp(message string) {
	fmt.Println()
	if field.help.IsValid() {
		if message != "" {
			logger.Error("%s", message)
		}

		// Execute the function
		rtc := field.help.Call([]reflect.Value{})
		if len(rtc) >= 1 {
			// The function did return a value. That means it returned a string value with delemiters
			if str, ok := rtc[0].Interface().(string); ok {
				// Trim all leading \n
				trimmed := strings.Trim(str, "\n")
				// Escape all "\|" with custom string
				escaped := strings.ReplaceAll(trimmed, "\\|", "~~~~****~~~~")
				// Now all "|" can be removed
				removed := strings.ReplaceAll(escaped, "|", "")
				// Unescae the previous escape sequence
				unescaped := strings.ReplaceAll(removed, "~~~~****~~~~", "|")

				// Print it
				fmt.Println(unescaped)
			} else {
				logger.Warning("Help() function did not return a string")
			}
		}
	} else {
		logger.Error("%s", message)
	}
	fmt.Println()
}

// Searches for the key in all the child fields
func getByKey(fields *cliField[any], key, rootPrefix string) (field *cliField[any], found bool) {
	for _, field := range fields.chields {
		if key == rootPrefix+field.longKey || key == rootPrefix+field.shortKey {
			return &field, true
		}
	}

	return nil, false
}

// Get the next bigger position field. If no one was found, nil will be returned
func getNextByPosition(fields *cliField[any], lastPosition int) (field *cliField[any]) {
	var minn *cliField[any]

	for i, field := range fields.chields {
		if field.requiredPos != 0 && field.requiredPos > lastPosition && (minn == nil || field.requiredPos < minn.requiredPos) {
			minn = &fields.chields[i]
		}
	}

	return minn
}

// Checks if the element is contained inside the array (Where is the comparable interface)
func contains[T any](array *[]T, element T) bool {
	for _, curr := range *array {
		if reflect.DeepEqual(curr, element) {
			return true
		}
	}

	return false
}

// Calls the setter of the field without a value
func (field *cliField[T]) callSetterWithoutValue() error {
	if field.disabled {
		return nil
	}

	if !field.setter.IsValid() {
		logger.Debug("No Setter found for %s. Because '~~~' was given, nothing will be done", field.longKey)
		return nil
	}

	response := field.setter.Call([]reflect.Value{})[0]
	r := response.Interface()
	if reflect.TypeOf(r).Kind() == reflect.String {
		if r != "" {
			return fmt.Errorf("%s", r)
		}
	}
	return nil
}

// Applies the given value to the struct / field
func (field *cliField[T]) setValue(value any) error {
	if field.disabled {
		return nil
	}

	if field.setter.IsValid() {
		in := make([]reflect.Value, field.setter.Type().NumIn())

		for i := range field.setter.Type().NumIn() {
			t := field.setter.Type().In(i)

			var err error
			object := value
			// The type is either a 'string' or '[]string'
			if strVal, ok := value.(string); ok {
				object, err = convertValue(strVal, t)
				if err != nil {
					return err
				}
			}

			in[i] = reflect.ValueOf(object)
		}

		response := field.setter.Call(in)[0]
		r := response.Interface()
		if reflect.TypeOf(r).Kind() == reflect.String {
			if r != "" {
				return fmt.Errorf("%s", r)
			}
		}
	} else {
		var err error
		valueToSet := value

		// The type is either a 'string' or '[]string'
		if strVal, ok := value.(string); ok {
			valueToSet, err = convertValue(strVal, field.reflectValue.Type())
			if err != nil {
				logger.Debug("Unable to convert the value '%s' for the key '%s'", value, field.longKey)
				return err
			}
		}

		if !field.reflectValue.CanSet() {
			return fmt.Errorf("cannot set field value")
		}

		field.reflectValue.Set(reflect.ValueOf(valueToSet))
	}

	return nil
}

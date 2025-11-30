package utils

import (
	"crypto/rand"
	"math/big"
	"os"
	"strconv"
	"strings"

	"git.rpjosh.de/RPJosh/go-logger"
)

// GetEnvString tries to get an environment variable from the system
// as a string value. If the env was not found the given default value
// will be returned
func GetEnvString(name, defaultValue string) string {
	val := defaultValue
	if strVal, isSet := os.LookupEnv(name); isSet {
		val = strVal
	}

	return val
}

// RequireEnvSecret returns the environment variable with the given name.
// If it could not be found, a fatal error will be logged and the program stops.
//
// If an environment variable with the suffix "_FILE" exists, the value is read
// from the file identified by the env value
func RequireEnvSecret(name string) string {
	if fileVal, isSet := os.LookupEnv(name + "_FILE"); isSet {
		// Red file
		content, err := os.ReadFile(fileVal)
		if err != nil {
			logger.Fatal("Failed to read secret file %q: %s", fileVal, err)
		}

		return string(content)
	}

	// Fall back to default behaviour
	return RequireEnvString(name)
}

// RequireEnvString returns the environment variable with the given name.
// If it could not be found, a fatal error will be logged and the program stops
func RequireEnvString(name string) string {
	if strVal, isSet := os.LookupEnv(name); isSet {
		return strVal
	} else {
		logger.Fatal("Required environment variable %q not set", name)
		return ""
	}
}

// GetEnvBool tries to get an environment variable from the system
// as a boolean value. If the env was not found the given default value
// will be returned
func GetEnvBool(name string, defaultValue bool) bool {
	val := defaultValue
	if strVal, isSet := os.LookupEnv(name); isSet {
		strVal = strings.ToLower(strVal)
		return strVal == "1" || strVal == "true" || strVal == "yes" || strVal == "ja"
	}

	return val
}

// GenerateRandomString returns a securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly
func GenerateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	if rnd, err := generateRandom(n, letters); err == nil {
		return string(rnd), nil
	} else {
		return "", err
	}
}

// GenerateRandomNumber returns a securely generated random number.
// It will return an error if the system's secure random
// number generator fails to function correctly
func GenerateRandomNumber(n int) (int64, error) {
	const letters = "0123456789"
	if rnd, err := generateRandom(n, letters); err == nil {
		return strconv.ParseInt(string(rnd), 10, 64)
	} else {
		return 0, err
	}
}

// generateRandom returns a securely generated random string with the provided character in it.
// It will return an error if the system's secure random
// number generator fails to function correctly
func generateRandom(n int, letters string) ([]byte, error) {
	ret := make([]byte, n)
	for i := range n {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return []byte{}, err
		}
		ret[i] = letters[num.Int64()]
	}

	return ret, nil
}

// GenerateRandomBytes generates a cryptographically secure random
// number of bytes
func GenerateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// Remove removes one element from the slice.
// The order won't be preserved for performance.
//
// Sample (remove [2]): 10, 20, 30, 40, 50 => 10, 20, 50, 40
func Remove[T any](s *[]T, i int) []T {
	(*s)[i] = (*s)[len(*s)-1]
	return (*s)[:len(*s)-1]
}

// RemovePreserveOrder is like [Remove] but preserves the order
// of elements.
// This method is by far (0.2 vs 8.9) not as efficient as
// [Remove] because a new copy of the slice is created
func RemovePreserveOrder[T any](s *[]T, i int) []T {
	return append((*s)[:i], (*s)[i+1:]...)
}

// ToInt transforms the provided value to an integer.
// Errors are logged internally
func ToInt(val string) int {
	rtc, err := strconv.Atoi(val)
	if err != nil {
		logger.Warning("Failed to convert %q to an integer: %s", val, err)
	}

	return rtc
}

// ToFloat transforms the provided value to a float value.
// Errors are logged internally
func ToFloat(val string) float64 {
	rtc, err := strconv.ParseFloat(val, 64)
	if err != nil {
		logger.Warning("Failed to convert %q to an float: %s", val, err)
	}

	return rtc
}

// WithoutError returns the direct value from a function that
// returns both a value and an error.
//
// If any error is provided, it will be logged with the warning level
func WithoutError[T any](val T, err error) T {
	if err != nil {
		logger.Warning("Ignored error: %s", err)
	}

	return val
}

package utils

import (
	"crypto/sha1"
	"encoding/hex"
	"runtime"
	"strings"

	"git.rpjosh.de/RPJosh/go-logger"
)

// GetCssHashOfFile returns the hash value of the provided file that is applied while compiling
// a stylesheet that is placed in a folder which contains your file.
// You have to provide a path like 'internal/api/router/hi.go'
func GetCssHashOfFile(file string) string {
	// Strip everything before "/internal"
	name := ""
	packageName := strings.Split(file, "/internal/")
	if len(packageName) > 1 {
		// Prefix was splitted -> remove the first part
		name = strings.Join(packageName[1:], "/")
	} else {
		// No prefix provided. Make sure that the name doesn't begin with a "/"
		name = strings.TrimPrefix(packageName[0], "/")
	}

	// Remove file
	name = name[0:strings.LastIndex(name, "/")]

	// Hash the name
	h := sha1.New()
	h.Write([]byte(name))
	hash := hex.EncodeToString(h.Sum(nil))[0:16]
	return "col-" + hash
}

// GetCssHash returns the hash value of the calling file
func GetCssHash() string {
	if _, file, _, ok := runtime.Caller(1); ok {
		return GetCssHashOfFile(file)
	} else {
		logger.Warning("Failed to get name of invoking file")
		return "error-1234"
	}
}

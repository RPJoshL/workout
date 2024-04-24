package utils

import (
	"crypto/sha1"
	"encoding/hex"
	"net/url"
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

// BuildUrl builds an URL with the provided query values.
// This function expects "key", "value" pairs as a parameter
func BuildUrl(baseURL string, params ...string) string {
	u, err := url.Parse(baseURL)
	if err != nil {
		logger.Error("Failed to parse URL %q", baseURL)
		u, _ = url.Parse("/")
	}

	query := u.Query()
	for i := 0; i < len(params); i += 2 {
		key := params[i]
		value := params[i+1]
		query.Add(key, value)
	}

	u.RawQuery = query.Encode()
	return u.String()
}

// IsTrue reports weather the provided value represents
// the boolean value "true"
func IsTrue(value string) bool {
	value = strings.ToLower(value)
	return value == "1" || value == "true" || value == "on"
}

// GetMbyte returns the provided amount of mbytes
// as bytes
func MToBytes(mByte int) int64 {
	return int64(mByte) * 1024 * 1024
}

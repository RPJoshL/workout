package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/workout/pkg/utils"
)

const cssFilePath = "./static/css/pages.css"
const sCSSFilePath = "./static/css/pages.scss"
const thirdPartyFilePath = "./static/css/third.css"
const sep = "---------------------"

// Program to collect
func main() {
	defer logger.CloseFile()

	// Configure logger
	logger.SetGlobalLogger(logger.GetLoggerFromEnv(&logger.Logger{
		ColoredOutput: true,
		Level:         logger.LevelInfo,
		PrintSource:   true,
		File:          &logger.FileLogger{},
	}))

	// Remove previous, compiled file
	removeFiles()

	// Create SCSS file to collect and write the CSS files into
	sCSSFile, err := os.Create(sCSSFilePath)
	if err != nil {
		logger.Fatal("Failed to create .scss file: %s", err)
	}
	defer sCSSFile.Close()

	// Find all .css files and write them into the sCSSFile
	var walk func(dir string)
	walk = func(dir string) {
		files, err := os.ReadDir(dir)
		if err != nil {
			logger.Fatal("Failed to read dir %q: %s", dir, err)
		}

		for _, info := range files {
			name := dir + "/" + info.Name()

			// Recursive walk
			if info.IsDir() {
				walk(name)
			} else if strings.HasSuffix(name, ".css") || strings.HasSuffix(name, ".scss") {
				// Hash the file name
				packageName := strings.Join(strings.Split(name, "./internal/")[1:], "/")
				hashContent := packageName[0:strings.LastIndex(packageName, "/")]
				h := sha1.New()
				h.Write([]byte(hashContent))
				hash := hex.EncodeToString(h.Sum(nil))[0:16]

				logger.Info("Added file %q: %s", name, hashContent)

				// Write file content of CSS to SCSS file
				cssContent, err := os.ReadFile(name)
				if err != nil {
					logger.Fatal("Failed to read content of file %q: %s", name, err)
				}
				sCSSFile.Write([]byte(fmt.Sprintf("\n// %s %s %s\n.col-%s {\n%s\n}",
					sep, name, sep, hash, cssContent,
				)))
			}
		}
	}
	walk("./internal")

	// Compile ".scss" to ".css"
	cmd := exec.Command("node-sass", sCSSFilePath, cssFilePath)
	if err := cmd.Run(); err != nil {
		logger.Error("Failed to run node-sass: %s", err)
	}
	cmd.Wait()

	// Remove ".scss" file
	if utils.GetEnvBool("REMOVE_SCSS_FILE", true) {
		if err := os.Remove(sCSSFilePath); err != nil {
			// Don't do anything if file does not exist
			if !errors.Is(err, os.ErrNotExist) {
				logger.Fatal("Failed to delete previous css file")
			}
		}
	}

	// Minify CSS file
	minifyFile(cssFilePath)

	// Append third party css files
	cmd = exec.Command("sh", "-c", fmt.Sprintf("cat %q >> %q", thirdPartyFilePath, cssFilePath))
	if err := cmd.Run(); err != nil {
		logger.Fatal("Failed to append third party css file: %s", err)
	}
	cmd.Wait()

	logger.Info("Compiled CSS file successfully")
}

// removeFiles removes all files that were created within this program
func removeFiles() {
	if err := os.Remove(cssFilePath); err != nil {
		// Don't do anything if file does not exist
		if !errors.Is(err, os.ErrNotExist) {
			logger.Fatal("Failed to delete previous css file")
		}
	}
	if err := os.Remove(sCSSFilePath); err != nil {
		// Don't do anything if file does not exist
		if !errors.Is(err, os.ErrNotExist) {
			logger.Fatal("Failed to delete previous css file")
		}
	}
}

// minifyFile minifies the provided file. The original file will be overwritten
// by the minified version
func minifyFile(path string) {
	cmd := exec.Command("minify", cssFilePath)

	// Create buffer to store the minified CSS file in (minify writes to stdout)
	buf := new(bytes.Buffer)
	cmd.Stdout = buf

	if err := cmd.Start(); err != nil {
		logger.Fatal("Failed to minify file: %q", err)
	}
	cmd.Wait()

	// Close file and overwrite original one
	origFile, err := os.Create(cssFilePath)
	if err != nil {
		logger.Fatal("Failed to open original file %q for minification: %s", path, err)
	}
	defer origFile.Close()

	if _, err := io.Copy(origFile, buf); err != nil {
		logger.Fatal("Failed to copy minified css file to original: %s", err)
	}
}

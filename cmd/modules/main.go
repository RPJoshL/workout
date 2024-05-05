package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/workout/pkg/utils"
)

const modulesPath = "./static/js/modules/"
const readPath = "./internal/api"

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

	// Javascript files to delete
	jsFiles := make([]string, 0)

	// Find all .js or .ts files and write them
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
			} else if strings.HasSuffix(name, ".js") || strings.HasSuffix(name, ".ts") {

				// Get the last part of the directory (expecting go module path)
				lastSlash := strings.LastIndex(dir, "/")
				goModule := dir[lastSlash:]
				logger.Info("Adding file %q to %s.js", name, goModule)

				// Append it to file
				cmd := exec.Command("sh", "-c", fmt.Sprintf("cat %q >> %q.js", name, modulesPath+goModule))
				if strings.HasSuffix(name, ".ts") {
					// Replace ".ts" with ".js"
					nameJs := strings.TrimSuffix(name, ".ts")
					nameJs += ".js"
					jsFiles = append(jsFiles, nameJs)

					command := fmt.Sprintf("tsc -t es2019 --allowSyntheticDefaultImports %q && cat %q | sed '/^import /d' >> '%s.js' && rm %q", name, nameJs, modulesPath+goModule, nameJs)
					cmd = exec.Command("sh", "-c", command)
				}

				// Buffer stdout and stderr
				var outbuf, errbuf strings.Builder // or bytes.Buffer
				cmd.Stdout = &outbuf
				cmd.Stderr = &errbuf
				if err := cmd.Run(); err != nil {
					logger.Error("%s", outbuf.String()+errbuf.String())
					logger.Fatal("Failed to run processing for file %q: %s", name, err)
				}
			}
		}
	}
	walk(readPath)

	// Remove any files (if existing)
	for _, f := range jsFiles {
		os.Remove(f)
	}

	// Print status
	logger.Info("Compiled modules successfully")

	if utils.GetEnvBool("DISABLE_MODULE_MINIFICATION", false) {
		os.Exit(0)
	}

	// Minify module files
	walk = func(dir string) {
		files, err := os.ReadDir(dir)
		if err != nil {
			logger.Fatal("Failed to read dir %q: %s", dir, err)
		}

		for _, info := range files {
			name := dir + info.Name()

			// Recursive walk
			if info.IsDir() {
				walk(name)
			} else if strings.HasSuffix(name, ".js") {
				cmd := exec.Command("sh", "-c", fmt.Sprintf(
					"cat %q | minify --js > '%s%s.tmp' && mv '%s%s.tmp' %q",
					name, dir, info.Name(), dir, info.Name(), name,
				))
				// Buffer stdout and stderr
				var outbuf, errbuf strings.Builder // or bytes.Buffer
				cmd.Stdout = &outbuf
				cmd.Stderr = &errbuf
				if err := cmd.Run(); err != nil {
					logger.Error("%s", outbuf.String()+errbuf.String())
					logger.Fatal("Failed to minify module %q: %s", name, err)
				}
				cmd.Wait()
			}
		}
	}
	walk(modulesPath)
}

// removeFiles removes all files that were created within this program
func removeFiles() {
	if err := os.RemoveAll(modulesPath); err != nil {
		// Don't do anything if file does not exist
		if !errors.Is(err, os.ErrNotExist) {
			logger.Fatal("Failed to delete previous css file")
		}
	}

	// Create simple ".gitkeep" again
	cmd := exec.Command("sh", "-c", fmt.Sprintf("mkdir %q && echo '' > %q.gitkeep", modulesPath, modulesPath))
	if err := cmd.Run(); err != nil {
		logger.Fatal("Failed to create gitkeep file in %q: %s", modulesPath, err)
	}
}

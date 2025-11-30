package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/workout/pkg/utils"
)

const (
	modulesPath = "./static/js/modules/"
	readPath    = "./internal/api"
)

// Regex for an import line in typescript
var importRegex = regexp.MustCompile(`(?m)import\s+\{?\s*([\w\s,]*)\s*\}?\s*from\s+["']([^"']+)["'];?`)

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
	var wg sync.WaitGroup
	var mtx sync.Mutex
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
				wg.Add(1)
				go func(name string) {
					lastSlash := strings.LastIndex(dir, "/")
					goModule := dir[lastSlash:]

					// Append it to file
					cmd := exec.Command("sh", "-c", fmt.Sprintf("cat %q >> %q.js", name, modulesPath+goModule))
					if before, ok := strings.CutSuffix(name, ".ts"); ok {
						// Replace ".ts" with ".js"
						nameJs := before
						nameJs += ".js"

						mtx.Lock()
						jsFiles = append(jsFiles, nameJs)
						mtx.Unlock()

						command := fmt.Sprintf(
							"tsc -t es2022 --baseUrl . --moduleResolution bundler --module esnext --allowSyntheticDefaultImports --skipLibCheck %q && cat %q | sed '/^import /d' >> '%s.js' && rm %q",
							name, nameJs, modulesPath+goModule, nameJs,
						)
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
					wg.Done()

					logger.Info("Added file %q to %s.js", name, goModule)
				}(name)
			}
		}
	}
	walk(readPath)
	wg.Wait()

	// Remove any files (if existing)
	for _, f := range jsFiles {
		if err := os.Remove(f); err != nil {
			logger.Warning("Failed to remove js file %q: %s", f, err)
		}
	}

	resolveImports()

	// Print status
	logger.Info("Compiled modules successfully")

	if utils.GetEnvBool("DISABLE_MODULE_MINIFICATION", false) {
		logger.Info("Minification disabled")

		// Change reload file
		if err := os.WriteFile("./nodemon.reload", []byte(time.Now().Format("15:04:05")), os.ModePerm); err != nil {
			logger.Error("Failed to write reload file './nodemon.reload': %s", err)
		}

		logger.CloseFile()
		return
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

	// Change reload file
	if err := os.WriteFile("./nodemon.reload", []byte(time.Now().Format("15:04:05")), os.ModePerm); err != nil {
		logger.Error("Failed to write reload file './nodemon.reload': %s", err)
	}
}

// resolveImports modifies the local import paths with the correctly bundled ones
func resolveImports() {
	err := filepath.Walk(modulesPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process js files
		if info.IsDir() || !strings.HasSuffix(path, ".js") {
			return nil
		}

		// Read complete file content
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %q: %w", path, err)
		}

		newContent := importRegex.ReplaceAllStringFunc(string(content), func(match string) string {
			matches := importRegex.FindStringSubmatch(match)
			if len(matches) < 3 {
				return match
			}

			imports, path := matches[1], matches[2]

			// Has to be at least inside two subfolders
			if strings.Count(path, "/") < 2 {
				return match
			}

			// Get resolved file path
			lastSlash := strings.LastIndex(path, "/")
			goModule := path[:lastSlash]
			lastSlash = strings.LastIndex(goModule, "/")
			goModule = goModule[lastSlash+1:]

			newPath := "/static/js/modules/" + goModule + ".js"
			logger.Trace("Replacing import path %q with %q", path, newPath)

			// Neue Import-Zeile bauen
			return fmt.Sprintf(`import { %s } from %q;`, imports, newPath)
		})
		err = os.WriteFile(path, []byte(newContent), 0o644)
		if err != nil {
			return fmt.Errorf("failed to write new file content to %q: %w", path, err)
		}

		return nil
	})

	if err != nil {
		logger.Fatal("Failed to resolve import paths: %s", err)
	}
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

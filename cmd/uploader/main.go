package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"git.rpjosh.de/RPJosh/go-logger"
)

// main starts the upload process
func main() {
	defer logger.CloseFile()

	// Parse configuration file
	filePath := "./config.yaml"
	flag.StringVar(&filePath, "config", filePath, "Path to the configuration file")
	config := GetConfig(filePath)

	// Search for new files every 15 seconds
	ticker := time.Tick(time.Duration(config.App.Interval) * time.Second)
	// Until user want to abort
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	// Initial run
	checkDirectory(config)

outer:
	for {
		select {
		case <-ticker:
			checkDirectory(config)
		case <-done:
			logger.Info("Received interrupt. Aborting..")
			break outer

		}
	}
}

// checkDirectory loops through all GPX files in the specified directory
// and uploads them as new workout files
func checkDirectory(conf Config) {
	baseDirectory := strings.TrimRight(conf.App.Directory, "/")

	entries, err := os.ReadDir(baseDirectory)
	if err != nil {
		logger.Error("Failed to read files in directory: %s", err)
		return
	}

	// Wait a few seconds after getting directroy info to only uploaded fully transfered files
	time.Sleep(1 * time.Second)

	// Number of files that were uploaded in this run
	uploadedFiles := 0

	for _, e := range entries {
		var err error
		var delete bool

		// Ignore directories
		if e.IsDir() {
			continue
		}

		// Only select files with GPX extension
		nameUpper := strings.ToUpper(e.Name())
		if strings.HasSuffix(nameUpper, ".GPX") {
			err = createWorkout(baseDirectory+"/"+e.Name(), conf)
			delete = true
		}

		// Move file to failed directory
		if err != nil {
			// Add timestamp to file name
			newName := fmt.Sprintf("%d_%s", time.Now().Unix(), e.Name())
			logger.Error("Failed to parse file %q: %s", e.Name(), err)
			logger.Info("Moving failed file to failed/%s", newName)

			// Create failed directory if not existing
			os.MkdirAll(baseDirectory+"/failed", os.ModePerm)

			// Move file to failed directory
			moveFile(baseDirectory+"/"+e.Name(), baseDirectory+"/failed/"+newName)
		}

		// Remove file
		if delete {
			os.Remove(baseDirectory + "/" + e.Name())
			uploadedFiles++
		}
	}

	// Call script if files were uploaded
	if uploadedFiles > 0 && conf.App.AfterUplaod != "" {
		cmd := exec.Command(conf.App.AfterUplaod)
		if err := cmd.Run(); err != nil {
			logger.Warning("Failed to run afterUpload script: %s", err)
		}
	}
}

func moveFile(source, dest string) {
	failedFile, err := os.Create(dest)
	if err != nil {
		logger.Error("Failed to move file to failed directory: %s", err)
	}
	defer failedFile.Close()
	sourceFile, err := os.Open(source)
	if err != nil {
		logger.Error("Failed to open source file: %s", err)
	}
	defer sourceFile.Close()
	if _, err := io.Copy(failedFile, sourceFile); err != nil {
		logger.Error("Failed to copy file into failed folder: %s", err)
	}
}

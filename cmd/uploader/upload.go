package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"git.rpjosh.de/RPJosh/go-logger"
)

// createWorkout creates a new workout by posting
// the file to the workout server.
//
// No additional workout details beside the GPX file are sent
func createWorkout(path string, conf Config) error {
	client := http.Client{Timeout: 15 * time.Second}

	// Build request body
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", filepath.Base(file.Name()))
	io.Copy(part, file)
	writer.Close()

	req, _ := http.NewRequest("POST", conf.App.Url+"/workout", body)
	req.Header.Add("Content-Type", writer.FormDataContentType())

	// Add authentication header
	req.Header.Add("Username", conf.User.Name)
	req.Header.Add("Password", conf.User.Password)
	req.Header.Add("Accept-Language", conf.App.Language)

	// Execute request
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	// Decody body first before checking status code to print in error message
	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	// Everything ok
	if res.StatusCode == 200 {
		logger.Debug("Created workout")
		return nil
	}

	// Workout already exists
	if res.StatusCode == 409 {
		logger.Debug("Workout already exists: %s", path)
		return nil
	}

	// Log response body
	logger.Debug("Response body:\n%s", resBody)

	return fmt.Errorf("creation of workout failed: %d", res.StatusCode)
}

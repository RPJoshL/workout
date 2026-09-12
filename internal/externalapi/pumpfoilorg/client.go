package pumpfoilorg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/RPJoshL/go-logger"
)

const defaultHTTPTimeout = 30 * time.Second

// client communicates with the pumpfoil.org API
type client struct {
	BaseURL     string
	DeviceToken string
	HTTPClient  *http.Client
}

func NewPumpfoilOrgClient(baseURL, deviceToken string) *client {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")

	return &client{
		BaseURL:     baseURL,
		DeviceToken: deviceToken,
		HTTPClient:  &http.Client{Timeout: defaultHTTPTimeout},
	}
}

type apiError struct {
	Status  int
	Message string
}

func (e *apiError) Error() string {
	return fmt.Sprintf("pumpfoil.org API error (HTTP %d): %s", e.Status, e.Message)
}

// getSessionPath returns the UI path to access the session
func (c *client) getSessionPath(id int) string {
	return c.BaseURL + "/sessions/" + strconv.Itoa(id)
}

func (c *client) do(method, path string, body any, auth bool) (rawData []byte, statusCode int, err error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("marshaling request body: %w", err)
		}

		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, bodyReader)
	if err != nil {
		return nil, 0, fmt.Errorf("creating http request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if auth && c.DeviceToken != "" {
		req.Header.Set("X-Device-Token", c.DeviceToken)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logger.Debug("Raw response for %s (status code: %d)\n%s", path, resp.StatusCode, string(data))
		return data, resp.StatusCode, &apiError{Status: resp.StatusCode, Message: strings.TrimSpace(string(data))}
	}

	return data, resp.StatusCode, nil
}

func (c *client) doAndParse(method, path string, body any, auth bool, res any) error {
	rawData, _, err := c.do(method, path, body, auth)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(rawData, res); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}

	return nil
}

type pairInitRequest struct {
	Platform string `json:"platform"`
	Label    string `json:"label"`
}

type pairInitResponse struct {
	Code       string    `json:"code"`
	ClaimToken string    `json:"claim_token"`
	ExpiresAt  time.Time `json:"expires_at"`
}

func (c *client) PairInit(label string) (rtc pairInitResponse, err error) {
	body := pairInitRequest{
		Platform: "RPout",
		Label:    label,
	}

	if err := c.doAndParse(http.MethodPost, "/api/devices/pair-init", body, false, &rtc); err != nil {
		return pairInitResponse{}, err
	}

	return rtc, nil
}

type pairPollResponse struct {
	DeviceToken *string `json:"device_token"`
}

func (c *client) PairPoll(claimToken string) (rtc pairPollResponse, err error) {
	path := "/api/devices/pair-poll?claim_token=" + claimToken

	if err := c.doAndParse(http.MethodGet, path, nil, false, &rtc); err != nil {
		return pairPollResponse{}, err
	}

	return rtc, nil
}

type sessionStartRequest struct {
	SessionUUID string `json:"session_uuid"`
	StartedAt   string `json:"started_at"`
	Sport       string `json:"sport"`
	GpsHz       int    `json:"gps_hz"`
	AccelHz     int    `json:"accel_hz"`
	AccelScale  int    `json:"accel_scale"`
	AppVersion  string `json:"app_version,omitempty"`
	FoilID      int    `json:"foil_id,omitempty"`
}

type sessionStartResponse struct {
	SessionID      int   `json:"session_id"`
	ReceivedChunks []int `json:"received_chunks"`
}

func (c *client) StartSession(req sessionStartRequest) (rtc sessionStartResponse, err error) {
	if err := c.doAndParse(http.MethodPost, "/api/ingest/session", req, true, &rtc); err != nil {
		return sessionStartResponse{}, err
	}

	return rtc, nil
}

type chunkRequest struct {
	Index    int    `json:"index"`
	Kind     string `json:"kind"`
	Encoding string `json:"encoding"`
	T0Ms     int    `json:"t0_ms"`
	Count    int    `json:"count"`
	Data     any    `json:"data"`
}

type chunkResponse struct {
	Ok    bool `json:"ok"`
	Index int  `json:"index"`
}

func (c *client) UploadChunk(sessionUUID string, req chunkRequest) (rtc chunkResponse, err error) {
	if err := c.doAndParse(http.MethodPost, "/api/ingest/session/"+sessionUUID+"/chunk", req, true, &rtc); err != nil {
		return chunkResponse{}, err
	}

	return rtc, nil
}

type SessionCompleteRequest struct {
	EndedAt     string `json:"ended_at"`
	TotalChunks int    `json:"total_chunks"`
}

type SessionCompleteResponse struct {
	SessionID int    `json:"session_id"`
	Status    string `json:"status"`
	Analysis  string `json:"analysis"`
}

func (c *client) CompleteSession(sessionUUID string, req SessionCompleteRequest) (*SessionCompleteResponse, error) {
	var res SessionCompleteResponse
	if err := c.doAndParse(http.MethodPost, "/api/ingest/session/"+sessionUUID+"/complete", req, true, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// PumpfoilOrgCredentials stores the device token for pumpfoil.org
type PumpfoilOrgCredentials struct {
	DeviceToken string `json:"device_token"`
	FoilID      int    `json:"foil_id"`
}

func parsePumpfoilOrgCredentials(credentials string) (*PumpfoilOrgCredentials, error) {
	var cred PumpfoilOrgCredentials
	if err := json.Unmarshal([]byte(credentials), &cred); err != nil {
		return nil, err
	}

	if cred.DeviceToken == "" {
		return nil, fmt.Errorf("missing device token")
	}
	return &cred, nil
}

func encodePumpfoilOrgCredentials(deviceToken string, foilID int) (string, error) {
	data, err := json.Marshal(PumpfoilOrgCredentials{DeviceToken: deviceToken, FoilID: foilID})
	if err != nil {
		return "", err
	}

	return string(data), nil
}

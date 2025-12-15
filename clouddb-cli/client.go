/*
 * Copyright 2025 Magnus Gille <mgille@gmail.com>
 */

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type CloudClient struct {
	BaseURL string
	Secret  string
	Client  *http.Client
}

func NewCloudClient(host, secret string) *CloudClient {
	return &CloudClient{
		BaseURL: fmt.Sprintf("http://%s/v1", host),
		Secret:  secret,
		Client:  &http.Client{},
	}
}

func (c *CloudClient) doRequest(method, endpoint string, body interface{}, result interface{}, params map[string]string) error {
	var bodyReader *bytes.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	} else {
		bodyReader = bytes.NewReader([]byte{})
	}

	u, err := url.Parse(c.BaseURL + endpoint)
	if err != nil {
		return fmt.Errorf("failed to parse url: %w", err)
	}

	if params != nil {
		q := u.Query()
		for k, v := range params {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}

	req, err := http.NewRequest(method, u.String(), bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.Secret != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Basic %s", c.Secret))
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	if result != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// GChart Methods

func (c *CloudClient) CreateGChart(chart GChartPostAPIv1) (string, error) {
	// The API returns the ID as a string in the body, not JSON
	// So we need to handle this specific case slightly differently or just read body

	jsonBody, err := json.Marshal(chart)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/gchart/", bytes.NewReader(jsonBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.Secret != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Basic %s", c.Secret))
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBytes))
	}

	return string(respBytes), nil
}

func (c *CloudClient) GetGChart(id string) (*GChartGetAPIv1, error) {
	var chart GChartGetAPIv1
	err := c.doRequest("GET", "/gchart/"+id, nil, &chart, nil)
	return &chart, err
}

func (c *CloudClient) SearchGChartHeaders(dateFrom string) (GChartAPIv1HeaderOnlyList, error) {
	var headers GChartAPIv1HeaderOnlyList
	params := map[string]string{}
	if dateFrom != "" {
		params["dateFrom"] = dateFrom
	}
	err := c.doRequest("GET", "/gchartheader", nil, &headers, params)
	return headers, err
}

// Version Methods

func (c *CloudClient) CreateVersion(ver VersionEntityPostAPIv1) (string, error) {
	// Similar to GChart, returns ID as string/int
	// doRequest expects JSON response for result pointers usually, but create returns simple string ID sometimes?
	// Checking code: response.WriteHeaderAndEntity(http.StatusCreated, strconv.FormatInt(key.IntID(), 10))
	// This writes the ID as the body.

	// So we use a custom request here as well for simplicity
	jsonBody, err := json.Marshal(ver)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/version", bytes.NewReader(jsonBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.Secret != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Basic %s", c.Secret))
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBytes))
	}

	return string(respBytes), nil
}

func (c *CloudClient) GetVersions(version string) (VersionEntityGetAPIv1List, error) {
	var versions VersionEntityGetAPIv1List
	params := map[string]string{}
	if version != "" {
		params["version"] = version
	}
	err := c.doRequest("GET", "/version", nil, &versions, params)
	return versions, err
}

func (c *CloudClient) GetLatestVersion() (*VersionEntityGetAPIv1, error) {
	var ver VersionEntityGetAPIv1
	err := c.doRequest("GET", "/version/latest", nil, &ver, nil)
	return &ver, err
}

// Telemetry Methods

func (c *CloudClient) UpsertTelemetry(tel TelemetryEntityPostAPIv1) (*TelemetryEntityGetAPIv1, error) {
	var result TelemetryEntityGetAPIv1
	// PUT /telemetry returns the object
	err := c.doRequest("PUT", "/telemetry", tel, &result, nil)
	return &result, err
}

func (c *CloudClient) GetTelemetry(createdAfter, updatedAfter, os, version string) (TelemetryEntityGetAPIv1List, error) {
	var list TelemetryEntityGetAPIv1List
	params := map[string]string{}
	if createdAfter != "" {
		params["createdAfter"] = createdAfter
	}
	if updatedAfter != "" {
		params["updatedAfter"] = updatedAfter
	}
	if os != "" {
		params["os"] = os
	}
	if version != "" {
		params["version"] = version
	}

	err := c.doRequest("GET", "/telemetry", nil, &list, params)
	return list, err
}

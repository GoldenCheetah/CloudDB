package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"unsafe"

	"context"

	"github.com/emicklei/go-restful"
	"google.golang.org/appengine/datastore"
)

type fakeKey struct {
	kind      string
	stringID  string
	intID     int64
	parent    *datastore.Key
	appID     string
	namespace string
}

func makeKey(kind, stringID string, intID int64, parent *datastore.Key) *datastore.Key {
	fk := fakeKey{
		kind:      kind,
		stringID:  stringID,
		intID:     intID,
		parent:    parent,
		appID:     "test-app",
		namespace: "",
	}
	return (*datastore.Key)(unsafe.Pointer(&fk))
}

func TestInsertGChart(t *testing.T) {
	os.Setenv("APPLICATION_ID", "test-app")
	os.Setenv("Basic_Auth", "secret")

	// Setup Mock Store
	mockStore := &MockStore{
		PutFunc: func(ctx context.Context, key *datastore.Key, src interface{}) (*datastore.Key, error) {
			return makeKey("gchartentity", "", 123, nil), nil
		},
		NewQueryFunc: func(kind string) Query {
			return &MockQuery{
				CountFunc: func(ctx context.Context) (int, error) {
					return 0, nil
				},
				GetAllFunc: func(ctx context.Context, dst interface{}) ([]*datastore.Key, error) {
					return nil, fmt.Errorf("mock error")
				},
			}
		},
		NewKeyFunc: func(ctx context.Context, kind, stringID string, intID int64, parent *datastore.Key) *datastore.Key {
			return makeKey(kind, stringID, intID, parent)
		},
		NewIncompleteKeyFunc: func(ctx context.Context, kind string, parent *datastore.Key) *datastore.Key {
			return makeKey(kind, "", 0, parent)
		},
		CacheGetFunc: func(ctx context.Context, key string, dst interface{}) error {
			return fmt.Errorf("mock error")
		},
	}
	DB = mockStore

	// Setup Request
	chart := GChartPostAPIv1{
		Header: CommonAPIHeaderV1{
			Name: "Test Chart",
		},
		ChartSport: "Bike",
	}
	body, _ := json.Marshal(chart)
	req, _ := http.NewRequest("POST", "/v1/gchart/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Basic secret")

	rw := httptest.NewRecorder()

	// Dispatch through DefaultContainer
	restful.DefaultContainer.ServeHTTP(rw, req)

	// Check Response
	if rw.Code != http.StatusCreated {
		t.Errorf("Expected 201 Created, got %d. Body: %s", rw.Code, rw.Body.String())
	}
}

func TestGetGChartById(t *testing.T) {
	os.Setenv("APPLICATION_ID", "test-app")
	os.Setenv("Basic_Auth", "secret")

	// Setup Mock Store
	mockStore := &MockStore{
		GetFunc: func(ctx context.Context, key *datastore.Key, dst interface{}) error {
			// Fill dst with dummy data
			if chart, ok := dst.(*GChartEntity); ok {
				chart.ChartSport = "Run"
				return nil
			}
			return datastore.ErrNoSuchEntity
		},
		NewKeyFunc: func(ctx context.Context, kind, stringID string, intID int64, parent *datastore.Key) *datastore.Key {
			return makeKey(kind, stringID, intID, parent)
		},
		NewQueryFunc: func(kind string) Query {
			return &MockQuery{
				GetAllFunc: func(ctx context.Context, dst interface{}) ([]*datastore.Key, error) {
					return nil, fmt.Errorf("mock error")
				},
			}
		},
		CacheGetFunc: func(ctx context.Context, key string, dst interface{}) error {
			return fmt.Errorf("mock error")
		},
	}
	DB = mockStore

	// Setup Request
	req, _ := http.NewRequest("GET", "/v1/gchart/123", nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Basic secret")

	rw := httptest.NewRecorder()

	// Dispatch through DefaultContainer
	restful.DefaultContainer.ServeHTTP(rw, req)

	// Check Response
	if rw.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d. Body: %s", rw.Code, rw.Body.String())
	}

	var respChart GChartGetAPIv1
	json.Unmarshal(rw.Body.Bytes(), &respChart)
	if respChart.ChartSport != "Run" {
		t.Errorf("Expected ChartSport 'Run', got '%s'", respChart.ChartSport)
	}
}

func TestUpsertTelemetry(t *testing.T) {
	os.Setenv("APPLICATION_ID", "test-app")
	os.Setenv("Basic_Auth", "secret")

	// Setup Mock Store
	mockStore := &MockStore{
		PutFunc: func(ctx context.Context, key *datastore.Key, src interface{}) (*datastore.Key, error) {
			return makeKey("telemetryentity", "test-key", 0, nil), nil
		},
		GetFunc: func(ctx context.Context, key *datastore.Key, dst interface{}) error {
			return datastore.ErrNoSuchEntity
		},
		NewKeyFunc: func(ctx context.Context, kind, stringID string, intID int64, parent *datastore.Key) *datastore.Key {
			return makeKey(kind, stringID, intID, parent)
		},
	}
	DB = mockStore

	// Setup Request
	telemetry := TelemetryEntityPostAPIv1{
		UserKey:    "test-key",
		OS:         "Linux",
		GCVersion:  "3.6",
		Increment:  25,
		LastChange: time.Now().Format(dateTimeLayout),
	}
	body, _ := json.Marshal(telemetry)
	req, _ := http.NewRequest("PUT", "/v1/telemetry", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic secret")

	rw := httptest.NewRecorder()

	// Dispatch through DefaultContainer
	restful.DefaultContainer.ServeHTTP(rw, req)

	// Check Response
	if rw.Code != http.StatusCreated {
		t.Errorf("Expected 201 Created, got %d. Body: %s", rw.Code, rw.Body.String())
	}
}

func TestGetVersion(t *testing.T) {
	os.Setenv("APPLICATION_ID", "test-app")
	os.Setenv("Basic_Auth", "secret")

	// Setup Mock Store
	mockStore := &MockStore{
		NewQueryFunc: func(kind string) Query {
			return &MockQuery{
				GetAllFunc: func(ctx context.Context, dst interface{}) ([]*datastore.Key, error) {
					// Populate dst with mock data
					versions := dst.(*[]VersionEntity)
					*versions = append(*versions, VersionEntity{
						Version:     3600,
						Type:        10,
						URL:         "http://example.com",
						Text:        "Release",
						VersionText: "3.6",
					})
					return []*datastore.Key{makeKey("versionentity", "", 1, nil)}, nil
				},
			}
		},
	}
	DB = mockStore

	// Setup Request
	req, _ := http.NewRequest("GET", "/v1/version?version=3500", nil)
	req.Header.Set("Authorization", "Basic secret")

	rw := httptest.NewRecorder()

	// Dispatch through DefaultContainer
	restful.DefaultContainer.ServeHTTP(rw, req)

	// Check Response
	if rw.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d. Body: %s", rw.Code, rw.Body.String())
	}

	var versions []VersionEntityGetAPIv1
	json.Unmarshal(rw.Body.Bytes(), &versions)
	if len(versions) != 1 {
		t.Errorf("Expected 1 version, got %d", len(versions))
	}
	if versions[0].Version != 3600 {
		t.Errorf("Expected Version 3600, got %d", versions[0].Version)
	}
}

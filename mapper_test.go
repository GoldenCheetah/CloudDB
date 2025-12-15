package main

import (
	"encoding/base64"
	"reflect"
	"testing"
	"time"
)

func TestCommonHeaderMapping(t *testing.T) {
	now := time.Now().Truncate(time.Second) // Truncate to second as format might lose precision
	nowStr := now.Format(dateTimeLayout)
	// Re-parse to match the precision loss in string conversion
	now, _ = time.Parse(dateTimeLayout, nowStr)

	api := &CommonAPIHeaderV1{
		Name:        "Test Name",
		Description: "Test Desc",
		Language:    "en",
		GcVersion:   "3.6",
		LastChanged: nowStr,
		CreatorId:   "user1",
		Curated:     true,
		Deleted:     false,
	}

	db := &CommonEntityHeader{}
	mapAPItoDBCommonHeader(api, db)

	if db.Name != api.Name {
		t.Errorf("Name mismatch: %s != %s", db.Name, api.Name)
	}
	if !db.LastChanged.Equal(now) {
		t.Errorf("LastChanged mismatch: %v != %v", db.LastChanged, now)
	}

	api2 := &CommonAPIHeaderV1{}
	mapDBtoAPICommonHeader(db, api2)

	if api2.Name != api.Name {
		t.Errorf("Reverse Name mismatch: %s != %s", api2.Name, api.Name)
	}
	if api2.LastChanged != api.LastChanged {
		t.Errorf("Reverse LastChanged mismatch: %s != %s", api2.LastChanged, api.LastChanged)
	}
}

func TestGChartMapping(t *testing.T) {
	imgData := []byte("fake image data")
	imgStr := base64.StdEncoding.EncodeToString(imgData)

	api := &GChartPostAPIv1{
		Header: CommonAPIHeaderV1{
			Name: "Chart1",
		},
		ChartSport:   "Bike",
		ChartType:    "Plot",
		ChartView:    "View1",
		ChartDef:     "Def1",
		Image:        imgStr,
		CreatorNick:  "Nick",
		CreatorEmail: "email@example.com",
	}

	db := &GChartEntity{}
	mapAPItoDBGChart(api, db)

	if db.ChartSport != api.ChartSport {
		t.Errorf("Sport mismatch")
	}
	if !reflect.DeepEqual(db.Image, imgData) {
		t.Errorf("Image mismatch")
	}

	// Test Image Size Limit
	largeImg := make([]byte, 1024001)
	largeImgStr := base64.StdEncoding.EncodeToString(largeImg)
	api.Image = largeImgStr
	mapAPItoDBGChart(api, db)
	if db.Image != nil {
		t.Errorf("Image should be nil if too large")
	}
}

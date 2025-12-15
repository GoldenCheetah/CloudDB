/*
 * Copyright 2025 Magnus Gille <mgille@gmail.com>
 */

package main

// Common Structures
type CommonAPIHeaderV1 struct {
	Id          int64  `json:"id"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	GcVersion   string `json:"gcversion"`
	LastChanged string `json:"lastChange"`
	CreatorId   string `json:"creatorId"`
	Language    string `json:"language"`
	Curated     bool   `json:"curated"`
	Deleted     bool   `json:"deleted"`
}

// GChart Structures
type GChartPostAPIv1 struct {
	Header       CommonAPIHeaderV1 `json:"header"`
	ChartSport   string            `json:"chartSport"`
	ChartType    string            `json:"chartType"`
	ChartView    string            `json:"chartView"`
	ChartDef     string            `json:"chartDef"`
	Image        string            `json:"image"`
	CreatorNick  string            `json:"creatorNick"`
	CreatorEmail string            `json:"creatorEmail"`
}

type GChartGetAPIv1 struct {
	Header       CommonAPIHeaderV1 `json:"header"`
	ChartSport   string            `json:"chartSport"`
	ChartType    string            `json:"chartType"`
	ChartView    string            `json:"chartView"`
	ChartDef     string            `json:"chartDef"`
	Image        string            `json:"image"`
	CreatorNick  string            `json:"creatorNick"`
	CreatorEmail string            `json:"creatorEmail"`
	DLCounter    int               `json:"downloadCount"`
}

type GChartGetAPIv1List []GChartGetAPIv1

type GChartAPIv1HeaderOnly struct {
	Header     CommonAPIHeaderV1 `json:"header"`
	ChartSport string            `json:"chartSport"`
	ChartType  string            `json:"chartType"`
	ChartView  string            `json:"chartView"`
}
type GChartAPIv1HeaderOnlyList []GChartAPIv1HeaderOnly

// Version Structures
type VersionEntityPostAPIv1 struct {
	Version     int    `json:"version"`
	Type        int    `json:"releaseType"`
	URL         string `json:"downloadURL"`
	VersionText string `json:"versionText"`
	Text        string `json:"text"`
}

type VersionEntityGetAPIv1 struct {
	Id          int64  `json:"id"`
	Version     int    `json:"version"`
	Type        int    `json:"releaseType"`
	URL         string `json:"downloadURL"`
	VersionText string `json:"versionText"`
	Text        string `json:"text"`
}

type VersionEntityGetAPIv1List []VersionEntityGetAPIv1

// Telemetry Structures
type TelemetryEntityPostAPIv1 struct {
	UserKey    string `json:"key"`
	LastChange string `json:"lastChange"`
	OS         string `json:"operatingSystem"`
	GCVersion  string `json:"version"`
	Increment  int64  `json:"increment"`
}

type TelemetryEntityGetAPIv1 struct {
	UserKey     string `json:"key"`
	Country     string `json:"country"`
	Region      string `json:"region"`
	City        string `json:"city"`
	CityLatLong string `json:"cityLatLong"`
	CreateDate  string `json:"createDate"`
	LastChange  string `json:"lastChange"`
	UseCount    int64  `json:"useCount"`
	OS          string `json:"operatingSystem"`
	GCVersion   string `json:"version"`
}

type TelemetryEntityGetAPIv1List []TelemetryEntityGetAPIv1

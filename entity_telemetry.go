/*
 * Copyright (c) 2016 Joern Rischmueller (joern.rm@gmail.com)
 *
 *  This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU Affero General Public License as
 *  published by the Free Software Foundation, either telemetry 3 of the
 *  License, or (at your option) any later telemetry.
 *
 *  This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU Affero General Public License for more details.
 *
 *  You should have received a copy of the GNU Affero General Public License
 *  along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package goldencheetah

import (
	"fmt"
	"net/http"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"

	"github.com/emicklei/go-restful"
)

// ---------------------------------------------------------------------------------------------------------------//
// Golden Cheetah (telemetry entity) which is stored in DB
// ---------------------------------------------------------------------------------------------------------------//
type TelemetryEntity struct {
	Country     string
	Region      string            `datastore:",noindex"`
	City        string            `datastore:",noindex"`
	CityLatLong string            `datastore:",noindex"`
	CreateDate  time.Time
	LastChange  time.Time
	UseCount    int64             `datastore:",noindex"`
	OS          string
	GCVersion   string
}

// ---------------------------------------------------------------------------------------------------------------//
// API View Definition
// ---------------------------------------------------------------------------------------------------------------//

// Full structure for POST/PUT
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

// ---------------------------------------------------------------------------------------------------------------//
// Data Storage View
// ---------------------------------------------------------------------------------------------------------------//

const telemetryDBEntityRootKey = "telemetryroot"
const telemetryDBEntity = "telemetryentity"

func mapAPItoDBTelemetry(api *TelemetryEntityPostAPIv1, db *TelemetryEntity) {
	if api.LastChange != "" {
		db.LastChange, _ = time.Parse(dateTimeLayout, api.LastChange)
	} else {
		db.LastChange = time.Now()
	}
	db.GCVersion = api.GCVersion
	db.OS = api.OS
}

func mapDBtoAPITelemetry(db *TelemetryEntity, api *TelemetryEntityGetAPIv1) {
	api.City = db.City
	api.Country = db.Country
	api.Region = db.Region
	api.CityLatLong = db.CityLatLong
	api.CreateDate = db.CreateDate.Format(dateTimeLayout)
	api.LastChange = db.LastChange.Format(dateTimeLayout)
	api.UseCount = db.UseCount
	api.GCVersion = db.GCVersion
	api.OS = db.OS
}

// supporting functions

// telemetryEntityKey returns the key used for all telemetryEntity entries.
func telemetryEntityRootKey(ctx context.Context) *datastore.Key {
	return datastore.NewKey(ctx, telemetryDBEntity, telemetryDBEntityRootKey, 0, nil)
}

// ---------------------------------------------------------------------------------------------------------------//
// request/response handler
// ---------------------------------------------------------------------------------------------------------------//

func upsertTelemetry(request *restful.Request, response *restful.Response) {
	ctx := appengine.NewContext(request.Request)

	telemetry := new(TelemetryEntityPostAPIv1)
	if err := request.ReadEntity(telemetry); err != nil {
		addPlainTextError(response, http.StatusInternalServerError, err.Error())
		return
	}
	// set Increment if not yet set (just in case)
	if telemetry.Increment == 0 {
		telemetry.Increment = 1
	}

	// No checks if the necessary fields are filed or not - since GoldenCheetah is
	// the only consumer of the APIs - any checks/response are to support this use-case

	// read if there is an entry existing for this IP Address
	key := datastore.NewKey(ctx, telemetryDBEntity, telemetry.UserKey, 0, telemetryEntityRootKey(ctx))

	currentTelemetry := new(TelemetryEntity)
	err := datastore.Get(ctx, key, currentTelemetry)
	if err == nil {
		// entry found, increment counter
		currentTelemetry.UseCount += telemetry.Increment
	} else {
		// entry not found, create a new one
		currentTelemetry.Country = request.HeaderParameter("X-AppEngine-Country")
		currentTelemetry.Region = request.HeaderParameter("X-AppEngine-Region")
		currentTelemetry.City = request.HeaderParameter("X-AppEngine-City")
		currentTelemetry.CityLatLong = request.HeaderParameter("X-AppEngine-CityLatLong")
		currentTelemetry.UseCount = 1
		currentTelemetry.CreateDate = time.Now()
	}
	// general mapping
	mapAPItoDBTelemetry(telemetry, currentTelemetry)

	if _, err := datastore.Put(ctx, key, currentTelemetry); err != nil {
		if appengine.IsOverQuota(err) {
			// return 503 and a text similar to what GAE is returning as well
			addPlainTextError(response, http.StatusServiceUnavailable, "503 - Over Quota")
		} else {
			addPlainTextError(response, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response.WriteHeaderAndEntity(http.StatusCreated, currentTelemetry)


}

func getTelemetry(request *restful.Request, response *restful.Response) {
	ctx := appengine.NewContext(request.Request)

	oldestDate := time.Date(2000, time.January, 1,0,0,0,0, time.UTC)
	var createdAfter time.Time
	var updatedAfter time.Time
	var err error
	if dateString := request.QueryParameter("createdAfter"); dateString != "" {
		createdAfter, err = time.Parse(time.RFC3339, dateString)
		if err != nil {
			addPlainTextError(response, http.StatusBadRequest, fmt.Sprint(err.Error(), " 'createdAfter' - Correct format is RFC3339"))
			return
		}
	} else {
		createdAfter = oldestDate
	}
	if dateString := request.QueryParameter("updatedAfter"); dateString != "" {
		updatedAfter, err = time.Parse(time.RFC3339, dateString)
		if err != nil {
			addPlainTextError(response, http.StatusBadRequest, fmt.Sprint(err.Error(), " 'updatedAfter' - Correct format is RFC3339"))
			return
		}
	} else {
		updatedAfter = oldestDate
	}
	os := request.QueryParameter("os")
	version := request.QueryParameter("version")

	// only one query parameter is processed on the request in case of multiple parameters,
	// follow the priority given by the sequence below (and ignore the other parameters)
	var q* datastore.Query
	if createdAfter != oldestDate {
		q = datastore.NewQuery(telemetryDBEntity).
			Filter("CreateDate >=", createdAfter)
	} else if updatedAfter != oldestDate {
		q = datastore.NewQuery(telemetryDBEntity).
			Filter("LastChange >=", updatedAfter)
	} else if os != "" {
		q = datastore.NewQuery(telemetryDBEntity).
			Filter("OS =", os)
	} else if version != "" {
		q = datastore.NewQuery(telemetryDBEntity).
			Filter("GCVersion =", version)
	} else {
		q = datastore.NewQuery(telemetryDBEntity)
	}

	var telemetryList TelemetryEntityGetAPIv1List

	var telemetryOnDBList []TelemetryEntity
	k, err := q.GetAll(ctx, &telemetryOnDBList)
	if err != nil && !isErrFieldMismatch(err) {
		if appengine.IsOverQuota(err) {
			// return 503 and a text similar to what GAE is returning as well
			addPlainTextError(response, http.StatusServiceUnavailable, "503 - Over Quota")
		} else {
			addPlainTextError(response, http.StatusInternalServerError, err.Error())
		}
		return
	}

	// DB Entity needs to be mapped back
	for i, telemetryDB := range telemetryOnDBList {
		var telemetryAPI TelemetryEntityGetAPIv1
		mapDBtoAPITelemetry(&telemetryDB, &telemetryAPI)
		telemetryAPI.UserKey = k[i].StringID()
		telemetryList = append(telemetryList, telemetryAPI)
	}

	response.WriteHeaderAndEntity(http.StatusOK, telemetryList)
}

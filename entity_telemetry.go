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
	Region      string
	City        string
	CityLatLong string
	CreateDate  time.Time
}

// ---------------------------------------------------------------------------------------------------------------//
// API View Definition
// ---------------------------------------------------------------------------------------------------------------//

// Full structure for POST/PUT
type TelemetryEntityPostAPIv1 struct {
	CreateDate string `json:"createDate"`
}

type TelemetryEntityGetAPIv1 struct {
	Country     string `json:"country"`
	Region      string `json:"region"`
	City        string `json:"city"`
	CityLatLong string `json:"cityLatLong"`
	CreateDate  string `json:"createDate"`
}

type TelemetryEntityGetAPIv1List []TelemetryEntityGetAPIv1

// ---------------------------------------------------------------------------------------------------------------//
// Data Storage View
// ---------------------------------------------------------------------------------------------------------------//

const telemetryDBEntityRootKey = "telemetryroot"
const telemetryDBEntity = "telemetryentity"

func mapAPItoDBTelemetry(api *TelemetryEntityPostAPIv1, db *TelemetryEntity) {
	if api.CreateDate != "" {
		db.CreateDate, _ = time.Parse(dateTimeLayout, api.CreateDate)
	} else {
		db.CreateDate = time.Now()
	}

}

func mapDBtoAPITelemetry(db *TelemetryEntity, api *TelemetryEntityGetAPIv1) {
	api.City = db.City
	api.Country = db.Country
	api.Region = db.Region
	api.CityLatLong = db.CityLatLong
	api.CreateDate = db.CreateDate.Format(dateTimeLayout)
}

// supporting functions

// telemetryEntityKey returns the key used for all telemetryEntity entries.
func telemetryEntityRootKey(ctx context.Context) *datastore.Key {
	return datastore.NewKey(ctx, telemetryDBEntity, telemetryDBEntityRootKey, 0, nil)
}

// ---------------------------------------------------------------------------------------------------------------//
// request/response handler
// ---------------------------------------------------------------------------------------------------------------//

func insertTelemetry(request *restful.Request, response *restful.Response) {
	ctx := appengine.NewContext(request.Request)

	telemetry := new(TelemetryEntityPostAPIv1)
	if err := request.ReadEntity(telemetry); err != nil {
		addPlainTextError(response, http.StatusInternalServerError, err.Error())
		return
	}

	// No checks if the necessary fields are filed or not - since GoldenCheetah is
	// the only consumer of the APIs - any checks/response are to support this use-case

	telemetryDB := new(TelemetryEntity)
	mapAPItoDBTelemetry(telemetry, telemetryDB)

	// now get the location information from the IP address
	telemetryDB.Country = request.HeaderParameter("X-AppEngine-Country")
	telemetryDB.Region = request.HeaderParameter("X-AppEngine-Region")
	telemetryDB.City = request.HeaderParameter("X-AppEngine-City")
	telemetryDB.CityLatLong = request.HeaderParameter("X-AppEngine-CityLatLong")

	// and now store it
	key := datastore.NewIncompleteKey(ctx, telemetryDBEntity, telemetryEntityRootKey(ctx))

	if _, err := datastore.Put(ctx, key, telemetryDB); err != nil {
		if appengine.IsOverQuota(err) {
			// return 503 and a text similar to what GAE is returning as well
			addPlainTextError(response, http.StatusServiceUnavailable, "503 - Over Quota")
		} else {
			addPlainTextError(response, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response.WriteHeaderAndEntity(http.StatusCreated, telemetryDB)


}

func getTelemetry(request *restful.Request, response *restful.Response) {
	ctx := appengine.NewContext(request.Request)

	var date time.Time
	var err error
	if dateString := request.QueryParameter("dateFrom"); dateString != "" {
		date, err = time.Parse(time.RFC3339, dateString)
		if err != nil {
			addPlainTextError(response, http.StatusBadRequest, fmt.Sprint(err.Error(), " - Correct format is RFC3339"))
			return
		}
	} else {
		date = time.Time{}
	}

	q := datastore.NewQuery(telemetryDBEntity).Filter("CreateDate >=", date).Order("-CreateDate")

	var telemetryList TelemetryEntityGetAPIv1List

	var telemetryOnDBList []TelemetryEntity
	_, err = q.GetAll(ctx, &telemetryOnDBList)
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
	for _, telemetryDB := range telemetryOnDBList {
		var telemetryAPI TelemetryEntityGetAPIv1
		mapDBtoAPITelemetry(&telemetryDB, &telemetryAPI)
		telemetryList = append(telemetryList, telemetryAPI)
	}

	response.WriteHeaderAndEntity(http.StatusOK, telemetryList)
}

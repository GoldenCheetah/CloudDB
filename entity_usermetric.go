/*
 * Copyright (c) 2015 Joern Rischmueller (joern.rm@gmail.com)
 *
 *  This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU Affero General Public License as
 *  published by the Free Software Foundation, either version 3 of the
 *  License, or (at your option) any later version.
 *
 *  This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU Affero General Public License for more details.
 *
 *  You should have received a copy of the GNU Affero General Public License
 *  along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */


package main

import (
	"google.golang.org/appengine/log"
	"net/http"
	"time"
	"strconv"
	"fmt"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"

	"github.com/emicklei/go-restful"
)


// ---------------------------------------------------------------------------------------------------------------//
// Full Golden Cheetah usermetric definition (usermetricentity) which is stored in DB
// ---------------------------------------------------------------------------------------------------------------//
type UserMetricEntity struct {
	Header       CommonEntityHeader
	MetricXML    string       `datastore:",noindex"`
	CreatorNick  string       `datastore:",noindex"`
	CreatorEmail string       `datastore:",noindex"`
}

type UserMetricEntityHeaderOnly struct {
	Header CommonEntityHeader
}


// ---------------------------------------------------------------------------------------------------------------//
// API View Definition
// ---------------------------------------------------------------------------------------------------------------//

// Full structure for GET and PUT
type UserMetricAPIv1 struct {
	Header       CommonAPIHeaderV1 `json:"header"`
	MetricXML    string      `json:"metrictxml"`
	CreatorNick  string      `json:"creatorNick"`
	CreatorEmail string      `json:"creatorEmail"`
}

type UserMetricAPIv1List []UserMetricAPIv1

// Header only structure
type UserMetricAPIv1HeaderOnly struct {
	Header CommonAPIHeaderV1 `json:"header"`
}
type UserMetricAPIv1HeaderOnlyList []UserMetricAPIv1HeaderOnly



// ---------------------------------------------------------------------------------------------------------------//
// Data Storage View
// ---------------------------------------------------------------------------------------------------------------//

const usermetricDBEntity = "usermetricentity"
const usermetricDBEntityRootKey = "usermetricroot"

func mapAPItoDBUserMetric(api *UserMetricAPIv1, db *UserMetricEntity) {
	mapAPItoDBCommonHeader(&api.Header, &db.Header)
	db.MetricXML = api.MetricXML
	db.CreatorNick = api.CreatorNick
	db.CreatorEmail = api.CreatorEmail
}


func mapDBtoAPIUserMetric(db* UserMetricEntity, api *UserMetricAPIv1) {
	mapDBtoAPICommonHeader(&db.Header, &api.Header)
	api.MetricXML = db.MetricXML
	api.CreatorNick = db.CreatorNick
	api.CreatorEmail = db.CreatorEmail
}



// supporting functions

// usermetricEntityKey returns the key used for all usermetricEntity entries.
func usermetricEntityRootKey(ctx context.Context) *datastore.Key {
	return datastore.NewKey(ctx, usermetricDBEntity, usermetricDBEntityRootKey, 0, nil)
}

// ---------------------------------------------------------------------------------------------------------------//
// request/response handler
// ---------------------------------------------------------------------------------------------------------------//

func insertUserMetric(request *restful.Request, response *restful.Response) {
	ctx := appengine.NewContext(request.Request)

	metric := new(UserMetricAPIv1)
	if err := request.ReadEntity(metric); err != nil {
		addPlainTextError(response, http.StatusInternalServerError, err.Error())
		return
	}

	// No checks if the necessary fields are filled or not - since GoldenCheetah is
	// the only consumer of the APIs - any checks/response are to support this use-case

	metricDB := new(UserMetricEntity)
	mapAPItoDBUserMetric(metric, metricDB)

	// complete/set POST fields
	metricDB.Header.LastChanged = time.Now()
	metricDB.Header.Curated = false
	metricDB.Header.Deleted = false

	// auto-curate if a registered "curator" is adding user metric
	curatorQuery := datastore.NewQuery(curatorDBEntity).Filter("CuratorId =", metricDB.Header.CreatorId)
	counter, _ := curatorQuery.Count(ctx) // ignore errors/just leave uncurated
	if counter == 1 {
		metricDB.Header.Curated = true
	} else {
		metricDB.Header.Curated = false
	}

	// and now store it
	key := datastore.NewIncompleteKey(ctx, usermetricDBEntity, usermetricEntityRootKey(ctx))
	key, err := datastore.Put(ctx, key, metricDB)
	if err != nil {
		commonResponseErrorProcessing (response, err)
		return
	}

	// send back the key
	response.WriteHeaderAndEntity(http.StatusCreated, strconv.FormatInt(key.IntID(), 10))

}

func updateUserMetric(request *restful.Request, response *restful.Response) {
	ctx := appengine.NewContext(request.Request)

	metric := new(UserMetricAPIv1)
	if err := request.ReadEntity(metric); err != nil {
		addPlainTextError(response, http.StatusInternalServerError, err.Error())
		return
	}

	if metric.Header.Id == 0 {
		addPlainTextError(response, http.StatusBadRequest, "Mandatory Key for Update is missing or invalid")
		return
	}

	// No more checks if the necessary fields are filled or not - since GoldenCheetah is
	// the only consumer of the APIs - any checks/response are to support this use-case

	metricDB := new(UserMetricEntity)
	mapAPItoDBUserMetric(metric, metricDB)
	metricDB.Header.LastChanged = time.Now()

	// and now store it

	key := datastore.NewKey(ctx, usermetricDBEntity, "", metric.Header.Id, usermetricEntityRootKey(ctx))
	if _, err := datastore.Put(ctx, key, metricDB); err != nil {
		commonResponseErrorProcessing (response, err)
		return
	}

	// Response is Empty for 204
	response.WriteHeaderAndEntity(http.StatusNoContent, "")

}
func getUserMetricHeader(request *restful.Request, response *restful.Response) {
	ctx := appengine.NewContext(request.Request)

	var date time.Time
	var err error
	var dateString string
	if dateString = request.QueryParameter("dateFrom"); dateString != "" {
		date, err = time.Parse(time.RFC3339, dateString)
		if err != nil {
			addPlainTextError(response, http.StatusBadRequest, fmt.Sprint(err.Error(), " - Correct format is RFC3339"))
			return
		}
	} else {
		date = time.Time{}
	}

	const maxNumberOfHeadersPerCall = 200; // this has to be equal to GoldenCheetah - CloudDBUserMetric class

	q := datastore.NewQuery(usermetricDBEntity).Filter("Header.LastChanged >=", date).Order("Header.LastChanged").Limit(maxNumberOfHeadersPerCall)

	var metricHeaderList UserMetricAPIv1HeaderOnlyList

	var metricsOnDBList []UserMetricEntityHeaderOnly
	k, err := q.GetAll(ctx, &metricsOnDBList)
	if err != nil && !isErrFieldMismatch(err) {
		commonResponseErrorProcessing (response, err)
		return
	}

	// DB Entity needs to be mapped back
	for i, metricDB := range metricsOnDBList {
		var metric UserMetricAPIv1HeaderOnly
		mapDBtoAPICommonHeader(&metricDB.Header, &metric.Header)
		metric.Header.Id = k[i].IntID()
		metricHeaderList = append(metricHeaderList, metric)
	}

	// write Info Log
	log.Infof(ctx, "GetHeader from: %s", dateString )

	response.WriteHeaderAndEntity(http.StatusOK, metricHeaderList)

}

func getUserMetricHeaderCount(request *restful.Request, response *restful.Response) {
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

	q := datastore.NewQuery(usermetricDBEntity).Filter("Header.LastChanged >=", date).Order("-Header.LastChanged")
	counter, _ := q.Count(ctx)

	response.WriteHeaderAndEntity(http.StatusOK, counter)

}

func getUserMetricById(request *restful.Request, response *restful.Response) {
	ctx := appengine.NewContext(request.Request)

	id := request.PathParameter("id")
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		commonResponseErrorProcessing (response, err)
		return
	}

	key := datastore.NewKey(ctx, usermetricDBEntity, "", i, usermetricEntityRootKey(ctx))

	metricDB := new(UserMetricEntity)
	err = datastore.Get(ctx, key, metricDB)
	if err != nil && !isErrFieldMismatch(err) {
		commonResponseErrorProcessing (response, err)
		return
	}


	// now map and respond
	metric := new(UserMetricAPIv1)
	mapDBtoAPIUserMetric(metricDB, metric)
	metric.Header.Id= key.IntID()

	response.WriteHeaderAndEntity(http.StatusOK, metric)
}

func deleteUserMetricById(request *restful.Request, response *restful.Response) {

	changeUserMetricById(request, response, true, false, true)

}

func curateUserMetricById(request *restful.Request, response *restful.Response) {

	newStatusString := request.QueryParameter("newStatus")
	b, err := strconv.ParseBool(newStatusString)
	if err != nil {
		addPlainTextError(response, http.StatusBadRequest, err.Error())
		return
	}
	changeUserMetricById(request, response, false, true, b)

}

// ------------------- supporting functions ------------------------------------------------

func changeUserMetricById(request *restful.Request, response *restful.Response, changeDeleted bool, changeCurated bool, newStatus bool) {
	c := appengine.NewContext(request.Request)

	id := request.PathParameter("id")
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		addPlainTextError(response, http.StatusBadRequest, err.Error())
		return
	}

	key := datastore.NewKey(c, usermetricDBEntity, "", i, usermetricEntityRootKey(c))

	metricDB := new(UserMetricEntity)
	err = datastore.Get(c, key, metricDB)
	if err != nil && !isErrFieldMismatch(err) {
		commonResponseErrorProcessing (response, err)
		return
	}

	// now update like requested

	if changeDeleted {
		metricDB.Header.Deleted = newStatus
		if newStatus {
			metricDB.MetricXML = ""
		}
		metricDB.Header.LastChanged = time.Now()
	}

	if changeCurated {
		metricDB.Header.Curated = newStatus
		metricDB.Header.LastChanged = time.Now()
	}

	if _, err := datastore.Put(c, key, metricDB); err != nil {
		if appengine.IsOverQuota(err) {
			// return 503 and a text similar to what GAE is returning as well
			addPlainTextError(response, http.StatusServiceUnavailable, "503 - Over Quota")
		} else {
			addPlainTextError(response, http.StatusInternalServerError, err.Error())
		}
		return
	}

	// Response is Empty for 204
	response.WriteHeaderAndEntity(http.StatusNoContent, "")

}


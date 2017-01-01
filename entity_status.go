/*
 * Copyright (c) 2015, 2016 Joern Rischmueller (joern.rm@gmail.com)
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


package goldencheetah

import (
	"net/http"
	"strconv"
	"time"
	"fmt"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/memcache"

	"github.com/emicklei/go-restful"
)


// ---------------------------------------------------------------------------------------------------------------//
// Golden Cheetah curator (statusentity) which is stored in DB
// ---------------------------------------------------------------------------------------------------------------//
type StatusEntity struct {
	Status     int
	ChangeDate time.Time
}

// Constants defined for documentation purposes - as they are set by GC
const (
	Status_Ok = 10
	Status_PartialFailure = 20
	Status_Outage = 30
)



type StatusEntityText struct {
	Text string                 `datastore:",noindex"`
}

// ---------------------------------------------------------------------------------------------------------------//
// API View Definition
// ---------------------------------------------------------------------------------------------------------------//

// Full structure for POST/PUT
type StatusEntityPostAPIv1 struct {
	Id         int64        `json:"id"`
	Status     int        `json:"status"`
	ChangeDate string        `json:"changeDate"`
	Text       string       `json:"text"`
}

type StatusEntityGetAPIv1 struct {
	Id         int64        `json:"id"`
	Status     int        `json:"status"`
	ChangeDate string        `json:"changeDate"`
}

type StatusEntityGetTextAPIv1 struct {
	Id   int64        `json:"id"`
	Text string       `json:"text"`
}

type StatusEntityGetAPIv1List []StatusEntityGetAPIv1

// ---------------------------------------------------------------------------------------------------------------//
// Memcache constants
// ---------------------------------------------------------------------------------------------------------------//

const statusMemcacheKey = "currentstatus"

// ---------------------------------------------------------------------------------------------------------------//
// Data Storage View
// ---------------------------------------------------------------------------------------------------------------//

const statusDBEntityRootKey = "statusroot"
const statusDBEntity = "statusentity"
const statusDBEntityText = "statusText"

func mapAPItoDBStatus(api *StatusEntityPostAPIv1, db *StatusEntity) {
	db.Status = api.Status
	if api.ChangeDate != "" {
		db.ChangeDate, _ = time.Parse(dateTimeLayout, api.ChangeDate)
	} else {
		db.ChangeDate = time.Now()
	}

}

func mapDBtoAPIStatus(db *StatusEntity, api *StatusEntityGetAPIv1) {
	api.Status = db.Status
	api.ChangeDate = db.ChangeDate.Format(dateTimeLayout)
}


// supporting functions

// statusEntityKey returns the key used for all statusEntity entries.
func statusEntityRootKey(ctx context.Context) *datastore.Key {
	return datastore.NewKey(ctx, statusDBEntity, statusDBEntityRootKey, 0, nil)
}

// ---------------------------------------------------------------------------------------------------------------//
// request/response handler
// ---------------------------------------------------------------------------------------------------------------//

func insertStatus(request *restful.Request, response *restful.Response) {
	ctx := appengine.NewContext(request.Request)

	status := new(StatusEntityPostAPIv1)
	if err := request.ReadEntity(status); err != nil {
		addPlainTextError(response, http.StatusInternalServerError, err.Error())
		return
	}

	// No checks if the necessary fields are filed or not - since GoldenCheetah is
	// the only consumer of the APIs - any checks/response are to support this use-case

	statusDB := new(StatusEntity)
	mapAPItoDBStatus(status, statusDB)

	// and now store it
	key := datastore.NewIncompleteKey(ctx, statusDBEntity, statusEntityRootKey(ctx))
	key, err := datastore.Put(ctx, key, statusDB);
	if err != nil {
		if appengine.IsOverQuota(err) {
			// return 503 and a text similar to what GAE is returning as well
			addPlainTextError(response, http.StatusServiceUnavailable, "503 - Over Quota")
		} else {
			addPlainTextError(response, http.StatusInternalServerError, err.Error())
		}
		return
	}

	if status.Text != "" {
		statusDBText := new(StatusEntityText)
		statusDBText.Text = status.Text
		// and now store it as child of statusEntry
		key := datastore.NewIncompleteKey(ctx, statusDBEntityText, key)
		key, err := datastore.Put(ctx, key, statusDBText);
		if err != nil {
			if appengine.IsOverQuota(err) {
				// return 503 and a text similar to what GAE is returning as well
				addPlainTextError(response, http.StatusServiceUnavailable, "503 - Over Quota")
			} else {
				addPlainTextError(response, http.StatusInternalServerError, err.Error())
			}
			return
		}
	}

	var in StatusEntityGetAPIv1
	in.Id = key.IntID()
	in.Status = status.Status
	in.ChangeDate = status.ChangeDate

	// flush the memcache
	memcache.Flush(ctx)

	// send back the key
	response.WriteHeaderAndEntity(http.StatusCreated, strconv.FormatInt(key.IntID(), 10))

}

func getStatus(request *restful.Request, response *restful.Response) {
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

	q := datastore.NewQuery(statusDBEntity).Filter("ChangeDate >=", date).Order("-ChangeDate")

	var statusList StatusEntityGetAPIv1List

	var statusOnDBList []StatusEntity
	k, err := q.GetAll(ctx, &statusOnDBList)
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
	for i, statusDB := range statusOnDBList {
		var statusAPI StatusEntityGetAPIv1
		mapDBtoAPIStatus(&statusDB, &statusAPI)
		statusAPI.Id = k[i].IntID()
		statusList = append(statusList, statusAPI)
	}

	response.WriteHeaderAndEntity(http.StatusOK, statusList)
}

func getCurrentStatus(request *restful.Request, response *restful.Response) {
	ctx := appengine.NewContext(request.Request)

	var statusAPI StatusEntityGetAPIv1

	// first check Memcache
	if _, err := memcache.Gob.Get(ctx, statusMemcacheKey, &statusAPI); err == nil {
		response.WriteHeaderAndEntity(http.StatusOK, statusAPI)
		return
	}

	q := datastore.NewQuery(statusDBEntity).Order("-ChangeDate").Limit(1)

	var statusOnDBList []StatusEntity
	k, err := q.GetAll(ctx, &statusOnDBList)
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
	mapDBtoAPIStatus(&statusOnDBList[0], &statusAPI)
	statusAPI.Id = k[0].IntID()

	// add to memcache / overwrite existing / ignore errors
	item := &memcache.Item{
		Key:   statusMemcacheKey,
		Object: statusAPI,
	}
	memcache.Gob.Set(ctx, item)

	response.WriteHeaderAndEntity(http.StatusOK, statusAPI)
}

func getStatusTextById(request *restful.Request, response *restful.Response) {
	ctx := appengine.NewContext(request.Request)

	id := request.PathParameter("id")
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		addPlainTextError(response, http.StatusBadRequest, err.Error())
		return
	}

	statusKey := datastore.NewKey(ctx, statusDBEntity, "", i, statusEntityRootKey(ctx))

	q := datastore.NewQuery(statusDBEntityText).Ancestor(statusKey).Limit(1) // we have max. 1 Text per status

	var statusTextOnDBList []StatusEntityText
	k, err := q.GetAll(ctx, &statusTextOnDBList)
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
	var statusAPI StatusEntityGetTextAPIv1
	statusAPI.Id = k[0].IntID()
	statusAPI.Text = statusTextOnDBList[0].Text

	response.WriteHeaderAndEntity(http.StatusOK, statusAPI)

}

//---------------------------------------------------------------------------------------
// internal functions
//---------------------------------------------------------------------------------------

func internalGetCurrentStatus(ctx context.Context) int {

	// first check Memcache
	if item, err := memcache.Get(ctx, statusMemcacheKey); err == nil {
		if i64, err := strconv.ParseInt(string(item.Value), 10, 0); err == nil {
			return int(i64)
		}
	}

	q := datastore.NewQuery(statusDBEntity).Order("-ChangeDate").Limit(1)

	var statusOnDBList []StatusEntity
	_, err := q.GetAll(ctx, &statusOnDBList)
	if (err != nil && !isErrFieldMismatch(err)) || len(statusOnDBList) == 0 {
		// we are not blocking to due problems in Status Management
		return Status_Ok
	}

	return statusOnDBList[0].Status
}




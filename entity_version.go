/*
 * Copyright (c) 2016 Joern Rischmueller (joern.rm@gmail.com)
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
// Golden Cheetah curator (versionentity) which is stored in DB
// ---------------------------------------------------------------------------------------------------------------//
type VersionEntity struct {
	Version    int
	ChangeDate time.Time
	Type       int
}

// Constants defined for documentation purposes - as they are set by GC
const (
	Version_Release = 10
	Version_Release_Candidate = 20
	Version_Development_Build = 30
	Version_Not_Found = 9999
)


type VersionEntityText struct {
	Text string                 `datastore:",noindex"`
}

// ---------------------------------------------------------------------------------------------------------------//
// API View Definition
// ---------------------------------------------------------------------------------------------------------------//

// Full structure for POST/PUT
type VersionEntityPostAPIv1 struct {
	Id         int64        `json:"id"`
	Version    int          `json:"version"`
	ChangeDate string       `json:"changeDate"`
	Text       string       `json:"text"`
}

type VersionEntityGetAPIv1 struct {
	Id         int64        `json:"id"`
	Version    int          `json:"version"`
	ChangeDate string       `json:"changeDate"`
}

type VersionEntityGetTextAPIv1 struct {
	Id   int64        `json:"id"`
	Text string       `json:"text"`
}

type VersionEntityGetAPIv1List []VersionEntityGetAPIv1

// ---------------------------------------------------------------------------------------------------------------//
// Memcache constants
// ---------------------------------------------------------------------------------------------------------------//

const versionMemcacheKey = "latestversion"

// ---------------------------------------------------------------------------------------------------------------//
// Data Storage View
// ---------------------------------------------------------------------------------------------------------------//

const versionDBEntityRootKey = "versionroot"
const versionDBEntity = "versionentity"
const versionDBEntityText = "versionText"

func mapAPItoDBVersion(api *VersionEntityPostAPIv1, db *VersionEntity) {
	db.Version = api.Version
	if api.ChangeDate != "" {
		db.ChangeDate, _ = time.Parse(dateTimeLayout, api.ChangeDate)
	} else {
		db.ChangeDate = time.Now()
	}

}

func mapDBtoAPIVersion(db *VersionEntity, api *VersionEntityGetAPIv1) {
	api.Version = db.Version
	api.ChangeDate = db.ChangeDate.Format(dateTimeLayout)
}


// supporting functions

// versionEntityKey returns the key used for all versionEntity entries.
func versionEntityRootKey(ctx context.Context) *datastore.Key {
	return datastore.NewKey(ctx, versionDBEntity, versionDBEntityRootKey, 0, nil)
}

// ---------------------------------------------------------------------------------------------------------------//
// request/response handler
// ---------------------------------------------------------------------------------------------------------------//

func insertVersion(request *restful.Request, response *restful.Response) {
	ctx := appengine.NewContext(request.Request)

	version := new(VersionEntityPostAPIv1)
	if err := request.ReadEntity(version); err != nil {
		addPlainTextError(response, http.StatusInternalServerError, err.Error())
		return
	}

	// No checks if the necessary fields are filed or not - since GoldenCheetah is
	// the only consumer of the APIs - any checks/response are to support this use-case

	versionDB := new(VersionEntity)
	mapAPItoDBVersion(version, versionDB)

	// and now store it
	key := datastore.NewIncompleteKey(ctx, versionDBEntity, versionEntityRootKey(ctx))
	key, err := datastore.Put(ctx, key, versionDB)
	if err != nil {
		if appengine.IsOverQuota(err) {
			// return 503 and a text similar to what GAE is returning as well
			addPlainTextError(response, http.StatusServiceUnavailable, "503 - Over Quota")
		} else {
			addPlainTextError(response, http.StatusInternalServerError, err.Error())
		}
		return
	}

	if version.Text != "" {
		versionDBText := new(VersionEntityText)
		versionDBText.Text = version.Text
		// and now store it as child of versionEntry
		key := datastore.NewIncompleteKey(ctx, versionDBEntityText, key)
		key, err := datastore.Put(ctx, key, versionDBText)
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

	var in VersionEntityGetAPIv1
	in.Id = key.IntID()
	in.Version = version.Version
	in.ChangeDate = version.ChangeDate

	// flush the memcache
	memcache.Flush(ctx)

	// send back the key
	response.WriteHeaderAndEntity(http.StatusCreated, strconv.FormatInt(key.IntID(), 10))

}

func getVersion(request *restful.Request, response *restful.Response) {
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

	q := datastore.NewQuery(versionDBEntity).Filter("ChangeDate >=", date).Order("-ChangeDate")

	var versionList VersionEntityGetAPIv1List

	var versionOnDBList []VersionEntity
	k, err := q.GetAll(ctx, &versionOnDBList)
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
	for i, versionDB := range versionOnDBList {
		var versionAPI VersionEntityGetAPIv1
		mapDBtoAPIVersion(&versionDB, &versionAPI)
		versionAPI.Id = k[i].IntID()
		versionList = append(versionList, versionAPI)
	}

	response.WriteHeaderAndEntity(http.StatusOK, versionList)
}

func getLatestVersion(request *restful.Request, response *restful.Response) {
	ctx := appengine.NewContext(request.Request)

	var versionAPI VersionEntityGetAPIv1

	// first check Memcache
	if _, err := memcache.Gob.Get(ctx, versionMemcacheKey, &versionAPI); err == nil {
		response.WriteHeaderAndEntity(http.StatusOK, versionAPI)
		return
	}

	q := datastore.NewQuery(versionDBEntity).Order("-ChangeDate").Limit(1)

	var versionOnDBList []VersionEntity
	k, err := q.GetAll(ctx, &versionOnDBList)
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
	mapDBtoAPIVersion(&versionOnDBList[0], &versionAPI)
	versionAPI.Id = k[0].IntID()

	// add to memcache / overwrite existing / ignore errors
	item := &memcache.Item{
		Key:   versionMemcacheKey,
		Object: versionAPI,
	}
	memcache.Gob.Set(ctx, item)

	response.WriteHeaderAndEntity(http.StatusOK, versionAPI)
}

func getVersionTextById(request *restful.Request, response *restful.Response) {
	ctx := appengine.NewContext(request.Request)

	id := request.PathParameter("id")
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		addPlainTextError(response, http.StatusBadRequest, err.Error())
		return
	}

	versionKey := datastore.NewKey(ctx, versionDBEntity, "", i, versionEntityRootKey(ctx))

	q := datastore.NewQuery(versionDBEntityText).Ancestor(versionKey).Limit(1) // we have max. 1 Text per version

	var versionTextOnDBList []VersionEntityText
	k, err := q.GetAll(ctx, &versionTextOnDBList)
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
	var versionAPI VersionEntityGetTextAPIv1
	versionAPI.Id = k[0].IntID()
	versionAPI.Text = versionTextOnDBList[0].Text

	response.WriteHeaderAndEntity(http.StatusOK, versionAPI)

}






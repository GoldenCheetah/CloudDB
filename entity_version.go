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
	"fmt"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"

	"github.com/emicklei/go-restful"
)


// ---------------------------------------------------------------------------------------------------------------//
// Golden Cheetah curator (versionentity) which is stored in DB
// ---------------------------------------------------------------------------------------------------------------//
type VersionEntity struct {
	Version     int
	Type        int          `datastore:",noindex"`
	URL         string       `datastore:",noindex"`
	Text        string       `datastore:",noindex"`
	VersionText string       `datastore:",noindex"`
}

// Constants defined for documentation purposes - as they are set by GC
const (
	Version_Release = 10
	Version_Release_Candidate = 20
	Version_Development_Build = 30
)


type VersionEntityText struct {
	Text string                 `datastore:",noindex"`
}

// ---------------------------------------------------------------------------------------------------------------//
// API View Definition
// ---------------------------------------------------------------------------------------------------------------//

// Full structure for POST/PUT
type VersionEntityPostAPIv1 struct {
	Version     int          `json:"version"`
	Type        int          `json:"releaseType"`
	URL         string       `json:"downloadURL"`
	VersionText string       `json:"versionText"`
	Text        string       `json:"text"`
}

type VersionEntityGetAPIv1 struct {
	Id          int64        `json:"id"`
	Version     int          `json:"version"`
	Type        int          `json:"releaseType"`
	URL         string       `json:"downloadURL"`
	VersionText string       `json:"versionText"`
	Text        string       `json:"text"`
}


type VersionEntityGetAPIv1List []VersionEntityGetAPIv1


// ---------------------------------------------------------------------------------------------------------------//
// Data Storage View
// ---------------------------------------------------------------------------------------------------------------//

const versionDBEntityRootKey = "versionroot"
const versionDBEntity = "versionentity"

func mapAPItoDBVersion(api *VersionEntityPostAPIv1, db *VersionEntity) {
	db.Version = api.Version
	db.Type = api.Type
	db.URL = api.URL
	db.Text = api.Text
	db.VersionText = api.Text
}

func mapDBtoAPIVersion(db *VersionEntity, api *VersionEntityGetAPIv1) {
	api.Version = db.Version
	api.Type = db.Type
	api.URL = db.URL
	api.Text = db.Text
	api.VersionText = db.VersionText
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

	// send back the key
	response.WriteHeaderAndEntity(http.StatusCreated, strconv.FormatInt(key.IntID(), 10))

}

func getVersion(request *restful.Request, response *restful.Response) {
	ctx := appengine.NewContext(request.Request)

	var version int
	var err error
	if versionString := request.QueryParameter("version"); versionString != "" {
		if version, err = strconv.Atoi(versionString); err != nil {
			addPlainTextError(response, http.StatusBadRequest, fmt.Sprint(err.Error(), " - Version: No integer string"))
			return
		}
	}

	q := datastore.NewQuery(versionDBEntity).Filter("Version >", version).Order("-Version")

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

	q := datastore.NewQuery(versionDBEntity).Order("-Version").Limit(1)

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

	response.WriteHeaderAndEntity(http.StatusOK, versionAPI)
}







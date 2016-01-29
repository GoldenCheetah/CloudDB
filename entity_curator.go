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


package goldencheetah

import (
	"net/http"
	"strconv"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"

	"github.com/emicklei/go-restful"

)


// ---------------------------------------------------------------------------------------------------------------//
// Golden Cheetah curator (curatorentity) which is stored in DB
// ---------------------------------------------------------------------------------------------------------------//
type CuratorEntity struct {
	CuratorId       string
	Nickname	    string
	Email           string

}

// ---------------------------------------------------------------------------------------------------------------//
// API View Definition
// ---------------------------------------------------------------------------------------------------------------//

// Full structure for GET and PUT
type CuratorAPIv1 struct {
	Id				int64		`json:"id"`
	CuratorId       string      `json:"curatorId"`
	Nickname        string      `json:"nickname"`
	Email           string      `json:"email"`
}

type CuratorAPIv1List []CuratorAPIv1


// ---------------------------------------------------------------------------------------------------------------//
// Data Storage View
// ---------------------------------------------------------------------------------------------------------------//

const curatorDBEntity = "curatorentity"
const curatorDBEntityRootKey = "curatorroot"


func mapAPItoDBCurator(api *CuratorAPIv1, db *CuratorEntity) {
	db.CuratorId = api.CuratorId
	db.Nickname = api.Nickname
	db.Email = api.Email
}


func mapDBtoAPICurator(db *CuratorEntity, api *CuratorAPIv1 ) {
	api.CuratorId = db.CuratorId
	api.Nickname = db.Nickname
	api.Email = db.Email
}


// supporting functions

// curatorEntityKey returns the key used for all curatorEntity entries.
func curatorEntityRootKey(c context.Context) *datastore.Key {
	return datastore.NewKey(c, curatorDBEntity, curatorDBEntityRootKey, 0, nil)
}

// ---------------------------------------------------------------------------------------------------------------//
// request/response handler
// ---------------------------------------------------------------------------------------------------------------//

func insertCurator(request *restful.Request, response *restful.Response) {
	ctx := appengine.NewContext(request.Request)

	curator := new(CuratorAPIv1)
	if err := request.ReadEntity(curator); err != nil {
		addPlainTextError(response,http.StatusInternalServerError, err.Error())
		return
	}

	// No checks if the necessary fields are filed or not - since GoldenCheetah is
	// the only consumer of the APIs - any checks/response are to support this use-case

	curatorDB := new(CuratorEntity)
	mapAPItoDBCurator(curator, curatorDB)

	// and now store it
	key := datastore.NewIncompleteKey(ctx, curatorDBEntity, curatorEntityRootKey(ctx))
	key, err := datastore.Put(ctx, key, curatorDB);
	if  err != nil {
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


func getCurator(request *restful.Request, response *restful.Response) {
	ctx := appengine.NewContext(request.Request)

	curatorString := request.QueryParameter("curatorId");

	var q *datastore.Query
	if curatorString != "" {
		q = datastore.NewQuery(curatorDBEntity).Filter("CuratorId =", curatorString)
	} else {
		q = datastore.NewQuery(curatorDBEntity)

	}
	var curatorOnDBList []CuratorEntity
	k, err := q.GetAll(ctx, &curatorOnDBList)
	if err != nil {
		if appengine.IsOverQuota(err) {
			// return 503 and a text similar to what GAE is returning as well
			addPlainTextError(response, http.StatusServiceUnavailable, "503 - Over Quota")
		} else {
			addPlainTextError(response, http.StatusInternalServerError, err.Error())
		}
		return
	}

	// DB Entity needs to be mapped back
	var curatorList CuratorAPIv1List
	for i, curatorDB := range curatorOnDBList {
		var curator CuratorAPIv1
		mapDBtoAPICurator(&curatorDB, &curator)
		curator.Id = k[i].IntID()
		curatorList = append (curatorList, curator)
	}

	response.WriteEntity(curatorList)
}


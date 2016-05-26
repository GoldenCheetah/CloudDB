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
	"time"
	"strconv"
	"fmt"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"

	b64 "encoding/base64"

	"github.com/emicklei/go-restful"
)


// ---------------------------------------------------------------------------------------------------------------//
// Full Golden Cheetah chart definition (gchartentity) which is stored in DB
// ---------------------------------------------------------------------------------------------------------------//
type GChartEntity struct {
	Header       CommonEntityHeader
	ChartType    string       `datastore:",noindex"`
	ChartView    string       `datastore:",noindex"`
	ChartDef     string       `datastore:",noindex"`
	Image        []byte       `datastore:",noindex"`
	CreatorNick  string       `datastore:",noindex"`
	CreatorEmail string       `datastore:",noindex"`
}

type GChartEntityHeaderOnly struct {
	Header CommonEntityHeader
}


// ---------------------------------------------------------------------------------------------------------------//
// API View Definition
// ---------------------------------------------------------------------------------------------------------------//

// Full structure for GET and PUT
type GChartAPIv1 struct {
	Header       CommonAPIHeaderV1 `json:"header"`
	ChartType    string      `json:"chartType"`
	ChartView    string      `json:"chartView"`
	ChartDef     string      `json:"chartDef"`
	Image        string      `json:"image"`
	CreatorNick  string      `json:"creatorNick"`
	CreatorEmail string      `json:"creatorEmail"`
}

type GChartAPIv1List []GChartAPIv1

// Header only structure
type GChartAPIv1HeaderOnly struct {
	Header CommonAPIHeaderV1 `json:"header"`
}
type GChartAPIv1HeaderOnlyList []GChartAPIv1HeaderOnly



// ---------------------------------------------------------------------------------------------------------------//
// Data Storage View
// ---------------------------------------------------------------------------------------------------------------//

const gChartDBEntity = "gchartentity"
const gChartDBEntityRootKey = "gchartsroot"

func mapAPItoDBGChart(api *GChartAPIv1, db *GChartEntity) {
	mapAPItoDBCommonHeader(&api.Header, &db.Header)
	db.ChartType = api.ChartType
	db.ChartView = api.ChartView
	db.ChartDef = api.ChartDef
	data, err := b64.StdEncoding.DecodeString(api.Image)
	if err != nil {
		data = nil
	} else {
		db.Image = data
	}
	db.CreatorNick = api.CreatorNick
	db.CreatorEmail = api.CreatorEmail
}


func mapDBtoAPIGChart(db *GChartEntity, api *GChartAPIv1) {
	mapDBtoAPICommonHeader(&db.Header, &api.Header)
	api.ChartType = db.ChartType
	api.ChartView = db.ChartView
	api.ChartDef = db.ChartDef
	api.Image = b64.StdEncoding.EncodeToString(db.Image)
	api.CreatorNick = db.CreatorNick
	api.CreatorEmail = db.CreatorEmail
}



// supporting functions

// chartEntityKey returns the key used for all chartEntity entries.
func gchartEntityRootKey(ctx context.Context) *datastore.Key {
	return datastore.NewKey(ctx, gChartDBEntity, gChartDBEntityRootKey, 0, nil)
}

// ---------------------------------------------------------------------------------------------------------------//
// request/response handler
// ---------------------------------------------------------------------------------------------------------------//

func insertGChart(request *restful.Request, response *restful.Response) {
	ctx := appengine.NewContext(request.Request)

	chart := new(GChartAPIv1)
	if err := request.ReadEntity(chart); err != nil {
		addPlainTextError(response, http.StatusInternalServerError, err.Error())
		return
	}

	// No checks if the necessary fields are filed or not - since GoldenCheetah is
	// the only consumer of the APIs - any checks/response are to support this use-case

	chartDB := new(GChartEntity)
	mapAPItoDBGChart(chart, chartDB)

	// complete/set POST fields
	chartDB.Header.LastChanged = time.Now()
	chartDB.Header.Curated = false
	chartDB.Header.Deleted = false

	// auto-curate if a registered "curator" is adding a gchart
	curatorQuery := datastore.NewQuery(curatorDBEntity).Filter("CuratorId =", chartDB.Header.CreatorId)
	counter, _ := curatorQuery.Count(ctx) // ignore errors/just leave uncurated
	if counter == 1 {
		chartDB.Header.Curated = true
	} else {
		chartDB.Header.Curated = false
	}

	// and now store it
	key := datastore.NewIncompleteKey(ctx, gChartDBEntity, gchartEntityRootKey(ctx))
	key, err := datastore.Put(ctx, key, chartDB);
	if err != nil {
		commonResponseErrorProcessing (response, err)
		return
	}

	// send back the key
	response.WriteHeaderAndEntity(http.StatusCreated, strconv.FormatInt(key.IntID(), 10))

}

func updateGChart(request *restful.Request, response *restful.Response) {
	ctx := appengine.NewContext(request.Request)

	chart := new(GChartAPIv1)
	if err := request.ReadEntity(chart); err != nil {
		addPlainTextError(response, http.StatusInternalServerError, err.Error())
		return
	}

	if (chart.Header.Id == 0) {
		addPlainTextError(response, http.StatusBadRequest, "Mandatory Id for Update is missing or invalid")
		return
	}

	// No more checks if the necessary fields are filed or not - since GoldenCheetah is
	// the only consumer of the APIs - any checks/response are to support this use-case

	chartDB := new(GChartEntity)
	mapAPItoDBGChart(chart, chartDB)

	chartDB.Header.LastChanged = time.Now()

	// and now store it

	key := datastore.NewKey(ctx, gChartDBEntity, "", chart.Header.Id, gchartEntityRootKey(ctx))
	if _, err := datastore.Put(ctx, key, chartDB); err != nil {
		commonResponseErrorProcessing (response, err)
		return
	}

	// Response is Empty for 204
	response.WriteHeaderAndEntity(http.StatusNoContent, "")

}
func getGChartHeader(request *restful.Request, response *restful.Response) {
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

	const maxNumberOfHeadersPerCall = 200; // this has to be equal to GoldenCheetah - CloudDBChartClient class

	q := datastore.NewQuery(gChartDBEntity).Filter("Header.LastChanged >=", date).Order("Header.LastChanged").Limit(maxNumberOfHeadersPerCall)

	var chartHeaderList GChartAPIv1HeaderOnlyList

	var chartsOnDBList []GChartEntityHeaderOnly
	k, err := q.GetAll(ctx, &chartsOnDBList)
	if err != nil && !isErrFieldMismatch(err) {
		commonResponseErrorProcessing (response, err)
		return
	}

	// DB Entity needs to be mapped back
	for i, chartDB := range chartsOnDBList {
		var chart GChartAPIv1HeaderOnly
		mapDBtoAPICommonHeader(&chartDB.Header, &chart.Header)
		chart.Header.Id = k[i].IntID()
		chartHeaderList = append(chartHeaderList, chart)
	}

	response.WriteHeaderAndEntity(http.StatusOK, chartHeaderList)

}

func getGChartHeaderCount(request *restful.Request, response *restful.Response) {
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

	q := datastore.NewQuery(gChartDBEntity).Filter("Header.LastChanged >=", date).Order("-Header.LastChanged")
	counter, _ := q.Count(ctx)

	response.WriteHeaderAndEntity(http.StatusOK, counter)

}

func getGChartById(request *restful.Request, response *restful.Response) {
	ctx := appengine.NewContext(request.Request)

	id := request.PathParameter("id")
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		commonResponseErrorProcessing (response, err)
		return
	}

	key := datastore.NewKey(ctx, gChartDBEntity, "", i, gchartEntityRootKey(ctx))

	chartDB := new(GChartEntity)
	err = datastore.Get(ctx, key, chartDB)
	if err != nil && !isErrFieldMismatch(err) {
		commonResponseErrorProcessing (response, err)
		return
	}

	// now map and respond
	chart := new(GChartAPIv1)
	mapDBtoAPIGChart(chartDB, chart)
	chart.Header.Id = key.IntID()

	response.WriteHeaderAndEntity(http.StatusOK, chart)
}

func deleteGChartById(request *restful.Request, response *restful.Response) {

	changeGChartById(request, response, true, false, true)

}

func curateGChartById(request *restful.Request, response *restful.Response) {

	newStatusString := request.QueryParameter("newStatus")
	b, err := strconv.ParseBool(newStatusString)
	if err != nil {
		commonResponseErrorProcessing (response, err)
		return
	}
	changeGChartById(request, response, false, true, b)

}

// ------------------- supporting functions ------------------------------------------------

func changeGChartById(request *restful.Request, response *restful.Response, changeDeleted bool, changeCurated bool, newStatus bool) {
	ctx := appengine.NewContext(request.Request)

	id := request.PathParameter("id")
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		addPlainTextError(response, http.StatusBadRequest, err.Error())
		return
	}

	key := datastore.NewKey(ctx, gChartDBEntity, "", i, gchartEntityRootKey(ctx))

	chartDB := new(GChartEntity)
	err = datastore.Get(ctx, key, chartDB)
	if err != nil && !isErrFieldMismatch(err) {
		commonResponseErrorProcessing (response, err)
		return
	}

	// now update like requested

	if changeDeleted {
		chartDB.Header.Deleted = newStatus
		if newStatus {
			chartDB.ChartType = ""
			chartDB.ChartView = ""
			chartDB.ChartDef = ""
			chartDB.Image = nil
		}
		chartDB.Header.LastChanged = time.Now()
	}

	if changeCurated {
		chartDB.Header.Curated = newStatus
		chartDB.Header.LastChanged = time.Now()
	}

	if _, err := datastore.Put(ctx, key, chartDB); err != nil {
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


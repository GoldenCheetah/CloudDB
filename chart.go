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
	"time"
	"strconv"
	"fmt"

	"appengine"
	"appengine/datastore"

	b64 "encoding/base64"

	"github.com/emicklei/go-restful"

)


// ---------------------------------------------------------------------------------------------------------------//
// Full Golden Cheetah chart definition (chartentity) which is stored in DB
// ---------------------------------------------------------------------------------------------------------------//
type ChartEntity struct {
	Name 	   		string       `datastore:",noindex"`
	Description		string       `datastore:",noindex"`
	Language		string       `datastore:",noindex"`
	GcVersion 		string
	ChartXML	    string       `datastore:",noindex"`
	ChartVersion    int
	Image			[]byte       `datastore:",noindex"`
	StatusId		int
	LastChanged 	time.Time
	CreatorId		string
	CreatorNick		string       `datastore:",noindex"`
	CreatorEmail	string       `datastore:",noindex"`
	Curated			bool
}


// ---------------------------------------------------------------------------------------------------------------//
// More re-use structures for the API and DB view
// ---------------------------------------------------------------------------------------------------------------//

// any status defined here must be in sync with the Client Status definitions
const (
    created = 0
)


// ---------------------------------------------------------------------------------------------------------------//
// API View Definition
// ---------------------------------------------------------------------------------------------------------------//

// Full structure for GET and PUT
type ChartAPIv1 struct {
	Id				int64		`json:"id"`
	Name            string   	`json:"name"`
	Description     string      `json:"description"`
	Language        string      `json:"language"`
	GcVersion       string      `json:"gcversion"`
	ChartXML        string      `json:"chartxml"`
	ChartVersion    int			`json:"chartversion"`
	Image           string      `json:"image"`
	StatusId 		int         `json:"statusId"`
	LastChanged     string      `json:"lastChange"`
	CreatorId       string      `json:"creatorId"`
	CreatorNick     string      `json:"creatorNick"`
	CreatorEmail    string      `json:"creatorEmail"`
	Curated			bool		`json:"curated"`
}

type ChartAPIv1List []ChartAPIv1


// ---------------------------------------------------------------------------------------------------------------//
// Data Storage View
// ---------------------------------------------------------------------------------------------------------------//

const chartDBEntity = "chartentity"
const chartDBEntityRootKey = "chartsroot"


func mapAPItoDBChart(api *ChartAPIv1, db *ChartEntity) {
	db.Name = api.Name
	db.Description = api.Description
	db.Language = api.Language
	db.GcVersion = api.GcVersion
	db.ChartXML = api.ChartXML
	db.ChartVersion = api.ChartVersion
	data, err := b64.StdEncoding.DecodeString(api.Image)
	if err != nil{
		data = nil
	} else {
		db.Image = data
	}
	db.StatusId = api.StatusId
	db.LastChanged, _ = time.Parse(dateTimeLayout, api.LastChanged)
	db.CreatorId = api.CreatorId
	db.CreatorNick = api.CreatorNick
	db.CreatorEmail = api.CreatorEmail
	db.Curated = api.Curated

}



func mapDBtoAPIChart(db *ChartEntity, api *ChartAPIv1 ) {
	api.Name = db.Name
	api.Description = db.Description
	api.Language = db.Language
	api.GcVersion = db.GcVersion
	api.ChartXML = db.ChartXML
	api.Image = b64.StdEncoding.EncodeToString(db.Image)
	api.StatusId = db.StatusId
	api.LastChanged = db.LastChanged.Format(dateTimeLayout)
	api.Curated = db.Curated
	api.CreatorId = db.CreatorId
	api.CreatorNick = db.CreatorNick
	api.CreatorEmail = db.CreatorEmail
}



// supporting functions

// chartEntityKey returns the key used for all chartEntity entries.
func chartEntityRootKey(c appengine.Context) *datastore.Key {
	return datastore.NewKey(c, chartDBEntity, chartDBEntityRootKey, 0, nil)
}

// ---------------------------------------------------------------------------------------------------------------//
// request/response handler
// ---------------------------------------------------------------------------------------------------------------//

func insertChart(request *restful.Request, response *restful.Response) {
	c := appengine.NewContext(request.Request)

	chart := new(ChartAPIv1)
	if err := request.ReadEntity(chart); err != nil {
		addPlainTextError(response,http.StatusInternalServerError, err.Error())
		return
	}

	// No checks if the necessary fields are filed or not - since GoldenCheetah is
	// the only consumer of the APIs - any checks/response are to support this use-case

	chartDB := new(ChartEntity)
	mapAPItoDBChart(chart, chartDB)

	// complete/set POST fields
	chartDB.ChartVersion = 1
	chartDB.LastChanged = time.Now()
	chartDB.Curated = false

	// and now store it
	key := datastore.NewIncompleteKey(c, chartDBEntity, chartEntityRootKey(c))
	key, err := datastore.Put(c, key, chartDB);
	if  err != nil {
		if appengine.IsOverQuota(err) {
			addPlainTextError(response, http.StatusPaymentRequired, err.Error())
		} else {
			addPlainTextError(response, http.StatusInternalServerError, err.Error())
		}
		return
	}

	// send back the key
	response.WriteHeaderAndEntity(http.StatusCreated, strconv.FormatInt(key.IntID(), 10))

}

func updateChart(request *restful.Request, response *restful.Response) {
	c := appengine.NewContext(request.Request)

	chart := new(ChartAPIv1)
	if 	err := request.ReadEntity(chart); err != nil {
		addPlainTextError(response, http.StatusInternalServerError, err.Error())
		return
	}

	if (chart.Id == 0) {
		addPlainTextError(response, http.StatusBadRequest, "Mandatory Id/Key for Update is missing or invalid")
		return
	}

	// No more checks if the necessary fields are filed or not - since GoldenCheetah is
	// the only consumer of the APIs - any checks/response are to support this use-case

	chartDB := new(ChartEntity)
	mapAPItoDBChart(chart, chartDB)

	// and now store it

	key := datastore.NewKey(c, chartDBEntity, "", chart.Id, chartEntityRootKey(c))
	if 	_, err := datastore.Put(c, key, chartDB); err != nil {
		if appengine.IsOverQuota(err) {
			addPlainTextError(response, http.StatusPaymentRequired, err.Error())
		} else {
			addPlainTextError(response, http.StatusInternalServerError, err.Error())
		}
		return
	}

	// send back the key
	response.WriteHeaderAndEntity(http.StatusOK, strconv.FormatInt(key.IntID(), 10))

}


func getCharts(request *restful.Request, response *restful.Response) {
	c := appengine.NewContext(request.Request)

	var date time.Time
	var err  error
	if dateString := request.QueryParameter("dateFrom"); dateString != "" {
		date, err = time.Parse(time.RFC3339, dateString)
		if err != nil {
			addPlainTextError(response, http.StatusBadRequest, fmt.Sprint(err.Error(), " - Correct format is RFC3339"))
			return
		}
	} else {
		date = time.Time{}
	}

	q:= datastore.NewQuery(chartDBEntity).Filter("LastChanged >=", date  )
	var chartsOnDBList []ChartEntity
	k, err := q.GetAll(c, &chartsOnDBList)
	if err != nil {
		if appengine.IsOverQuota(err) {
			addPlainTextError(response, http.StatusPaymentRequired, err.Error())
		} else {
			addPlainTextError(response, http.StatusInternalServerError, err.Error())
		}
	}

	// DB Entity needs to be mapped back
	var chartList ChartAPIv1List
	for i, chartDB := range chartsOnDBList {
		var chart ChartAPIv1
		mapDBtoAPIChart(&chartDB, &chart)
		chart.Id = k[i].IntID()
		chartList = append (chartList, chart)
	}

	response.WriteEntity(chartList)
}

func getChartById(request *restful.Request, response *restful.Response) {
	c := appengine.NewContext(request.Request)

	id := request.PathParameter("id")
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		addPlainTextError(response, http.StatusBadRequest, err.Error())
		return
	}

	key := datastore.NewKey(c, chartDBEntity, "", i, chartEntityRootKey(c))

	chartDB := new(ChartEntity)
	err = datastore.Get(c, key, chartDB)
	if err != nil {
		switch {
		case appengine.IsOverQuota(err):
			addPlainTextError(response, http.StatusPaymentRequired, err.Error())
		case err == datastore.ErrNoSuchEntity:
			addPlainTextError(response, http.StatusNotFound, err.Error())
		default:
			addPlainTextError(response, http.StatusBadRequest, err.Error())
		}
		return
	}

	// now map and respond
	chart := new(ChartAPIv1)
	mapDBtoAPIChart(chartDB, chart)
	chart.Id = key.IntID()

	response.WriteEntity(chart)
}
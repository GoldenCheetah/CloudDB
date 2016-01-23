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

	"os"
	"fmt"
	"net/http"

    "github.com/emicklei/go-restful"  // @Version Tag  v1.2
)

// init the Webserver within the GAE framework
func init() {

	ws := new(restful.WebService)

	// ----------------------------------------------------------------------------------
	// setup the charts endpoints - processing see "charts.go"
	// ----------------------------------------------------------------------------------
	ws.
	Path("/v1").
	Doc("Manage Charts").
	Consumes(restful.MIME_JSON).
	Produces(restful.MIME_JSON) // you can specify this per route as well

	ws.Route(ws.POST("/chart/").Filter(basicAuthenticate).To(insertChart).
	// docs
	Doc("creates a chart").
	Operation("createChart").
	Reads(ChartAPIv1{})) // from the request

	ws.Route(ws.PUT("/chart/").Filter(basicAuthenticate).To(updateChart).
	// docs
	Doc("updates a chart").
	Operation("updatedChart").
	Reads(ChartAPIv1{})) // from the request

	ws.Route(ws.GET("/chart/{id}").Filter(basicAuthenticate).To(getChartById).
	// docs
	Doc("get a chart").
	Operation("getChartbyId").
	Param(ws.PathParameter("id", "identifier of the chart").DataType("string")).
	Writes(ChartAPIv1{})) // on the response

	ws.Route(ws.DELETE("/chart/{id}").Filter(basicAuthenticate).To(deleteChartById).
	// docs
	Doc("delete a chart by setting the deleted status").
	Operation("deleteChartbyId").
	Param(ws.PathParameter("id", "identifier of the chart").DataType("string")))

	ws.Route(ws.PUT("/chartcuration/{id}").Filter(basicAuthenticate).To(curateChartById).
	// docs
	Doc("set the curation status of the chart to {newStatus} which must be 'true' or 'false' ").
	Operation("updateChartCurationStatus").
	Param(ws.PathParameter("id", "identifier of the chart").DataType("string")).
	Param(ws.QueryParameter("newStatus", "true/false curation status").DataType("bool")))

	// Endpoint for ChartHeader only (no JPG or LTMSettings)
	ws.Route(ws.GET("/chartheader").Filter(basicAuthenticate).To(getChartHeader).
	// docs
	Doc("gets a collection of charts header - in buckets of 500 charts - table sort is new to old").
	Operation("getChartHeader").
	Param(ws.QueryParameter("dateFrom", "Date of last change").DataType("string")).
	Writes(ChartAPIv1HeaderOnlyList{})) // on the response

	// Count Chart Headers to be retrieved
	ws.Route(ws.GET("/chartheader/count").Filter(basicAuthenticate).To(getChartHeaderCount).
	// docs
	Doc("gets the number of chart headers for testing,... selection").
	Operation("getChartHeader").
	Param(ws.QueryParameter("dateFrom", "Date of last change").DataType("string")))


	// ----------------------------------------------------------------------------------
	// setup the curator endpoints - processing see "charts.go"
	// ----------------------------------------------------------------------------------
	ws.Route(ws.GET("/curator").Filter(basicAuthenticate).To(getCurator).
	// docs
	Doc("gets a collection of curators").
	Operation("getCurator").
	Param(ws.QueryParameter("curatorId", "UUid of the Curator").DataType("string")).
	Writes(CuratorAPIv1List{})) // on the response

	ws.Route(ws.POST("/curator").Filter(basicAuthenticate).To(insertCurator).
	// docs
	Doc("creates a curator").
	Operation("createCurator").
	Reads(CuratorAPIv1{})) // from the request


	// all routes defined - let's go

	restful.Add(ws)

} // init()


// global declarations
const basicauth = "Basic_Auth"
const authorization = "Authorization"
const dateTimeLayout = "2006-01-02T15:04:05Z"


func basicAuthenticate(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	headerClientId := req.Request.Header.Get(authorization)
	if secretClientId := os.Getenv(basicauth); secretClientId != "" {
		if fmt.Sprint("Basic ",secretClientId) != headerClientId {
			resp.AddHeader("WWW-Authenticate", "Basic realm=Protected Area")
			resp.WriteErrorString(http.StatusUnauthorized, "Not Authorized")
			return
		}
	} else {
		resp.AddHeader("WWW-Authenticate", "Basic realm=Protected Area")
		resp.WriteErrorString(http.StatusInternalServerError, "Authorization configuration missing on Server")
		return
	}

	chain.ProcessFilter(req, resp)
} // basicAuthenticate

// Convenience functions for error handling
func addPlainTextError( r *restful.Response, httpStatus int, errorReason string ) {
	r.AddHeader("Content-Type", "text/plain")
	r.WriteErrorString(httpStatus, errorReason)
}

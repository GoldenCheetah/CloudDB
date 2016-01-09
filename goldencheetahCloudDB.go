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

	ws.Route(ws.POST("/chart").Filter(basicAuthenticate).To(insertChart).
	// docs
	Doc("creates a chart").
	Operation("createChart").
	Reads(ChartAPIv1{})) // from the request

	ws.Route(ws.PUT("/chart").Filter(basicAuthenticate).To(updateChart).
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


	// Endpoint for ChartHeader only (no JPG or LTMSettings)
	ws.Route(ws.GET("/chartheader").Filter(basicAuthenticate).To(getChartHeader).
	// docs
	Doc("gets a collection of charts header").
	Operation("getChartHeader").
	Param(ws.QueryParameter("dateFrom", "Date of last change").DataType("string")).
	Param(ws.QueryParameter("curated", "Curated true/false").DataType("bool")).

	Writes(ChartAPIHeaderV1List{})) // on the response

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

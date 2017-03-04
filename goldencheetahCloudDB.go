/*
 * Copyright (c) 2015, 2016, 2017 Joern Rischmueller (joern.rm@gmail.com)
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

	"google.golang.org/appengine"

	"github.com/emicklei/go-restful"  // @Version Tag  v1.2
)

// init the Webserver within the GAE framework
func init() {

	ws := new(restful.WebService)

	// ----------------------------------------------------------------------------------
	// setup the gcharts endpoints - processing see "entity_gcharts.go"
	// ----------------------------------------------------------------------------------
	ws.
	Path("/v1").
	Doc("Manage GCharts").
	Consumes(restful.MIME_JSON).
	Produces(restful.MIME_JSON) // you can specify this per route as well

	ws.Route(ws.POST("/gchart/").Filter(basicAuthenticate).Filter(filterCloudDBStatus).To(insertGChart).
	// docs
	Doc("creates a gchart").
	Operation("createGChart").
	Reads(GChartPostAPIv1{})) // from the request

	ws.Route(ws.PUT("/gchart/").Filter(basicAuthenticate).Filter(filterCloudDBStatus).To(updateGChart).
	// docs
	Doc("updates a gchart").
	Operation("updatedGChart").
	Reads(GChartPostAPIv1{})) // from the request

	ws.Route(ws.GET("/gchart/{id}").Filter(basicAuthenticate).Filter(filterCloudDBStatus).To(getGChartById).
	// docs
	Doc("get a gchart").
	Operation("getGChartbyId").
	Param(ws.PathParameter("id", "identifier of the gchart").DataType("string")).
	Writes(GChartGetAPIv1{})) // on the response

	ws.Route(ws.PUT("/gchartuse/{id}").Filter(basicAuthenticate).Filter(filterCloudDBStatus).To(incrementGChartUsageById).
	// docs
	Doc("increments the DL use counter for a chart by 1").
	Operation("incrementUsageCounterById").
	Param(ws.PathParameter("id", "identifier of the gchart").DataType("string")))

	ws.Route(ws.DELETE("/gchart/{id}").Filter(basicAuthenticate).Filter(filterCloudDBStatus).To(deleteGChartById).
	// docs
	Doc("delete a gchart by setting the deleted status").
	Operation("deleteGChartbyId").
	Param(ws.PathParameter("id", "identifier of the chart").DataType("string")))

	ws.Route(ws.PUT("/gchartcuration/{id}").Filter(basicAuthenticate).Filter(filterCloudDBStatus).To(curateGChartById).
	// docs
	Doc("set the curation status of the gchart to {newStatus} which must be 'true' or 'false' ").
	Operation("updateGChartCurationStatus").
	Param(ws.PathParameter("id", "identifier of the gchart").DataType("string")).
	Param(ws.QueryParameter("newStatus", "true/false curation status").DataType("bool")))

	// Endpoint for GChartHeader only (no JPG or Definition)
	ws.Route(ws.GET("/gchartheader").Filter(basicAuthenticate).Filter(filterCloudDBStatus).To(getGChartHeader).
	// docs
	Doc("gets a collection of gcharts header - in buckets of x charts - table sort is new to old").
	Operation("getGChartHeader").
	Param(ws.QueryParameter("dateFrom", "Date of last change").DataType("string")).
	Writes(GChartAPIv1HeaderOnlyList{})) // on the response

	// Count Chart Headers to be retrieved
	ws.Route(ws.GET("/gchartheader/count").Filter(basicAuthenticate).Filter(filterCloudDBStatus).To(getGChartHeaderCount).
	// docs
	Doc("gets the number of gchart headers for testing,... selection").
	Operation("getGChartHeader").
	Param(ws.QueryParameter("dateFrom", "Date of last change").DataType("string")))


	// ----------------------------------------------------------------------------------
	// setup the usermetric endpoints - processing see "entity_usermetric.go"
	// ----------------------------------------------------------------------------------
	ws.
	Path("/v1").
	Doc("Manage User Metrics").
	Consumes(restful.MIME_JSON).
	Produces(restful.MIME_JSON) // you can specify this per route as well

	ws.Route(ws.POST("/usermetric/").Filter(basicAuthenticate).Filter(filterCloudDBStatus).To(insertUserMetric).
	// docs
	Doc("creates a usermetric").
	Operation("createUserMetric").
	Reads(UserMetricAPIv1{})) // from the request

	ws.Route(ws.PUT("/usermetric/").Filter(basicAuthenticate).Filter(filterCloudDBStatus).To(updateUserMetric).
	// docs
	Doc("updates a usermetric").
	Operation("updateUserMetric").
	Reads(UserMetricAPIv1{})) // from the request

	ws.Route(ws.GET("/usermetric/{id}").Filter(basicAuthenticate).Filter(filterCloudDBStatus).To(getUserMetricByKey).
	// docs
	Doc("get a usermetric").
	Operation("getUserMetricbyId").
	Param(ws.PathParameter("key", "identifier of the user metric").DataType("string")).
	Writes(UserMetricAPIv1{})) // on the response

	ws.Route(ws.DELETE("/usermetric/{key}").Filter(basicAuthenticate).Filter(filterCloudDBStatus).To(deleteUserMetricByKey).
	// docs
	Doc("delete a usermetric by setting the deleted status").
	Operation("deleteUserMetricbyKey").
	Param(ws.PathParameter("key", "identifier of the usermetric").DataType("string")))

	ws.Route(ws.PUT("/usermetriccuration/{key}").Filter(basicAuthenticate).Filter(filterCloudDBStatus).To(curateUserMetricByKey).
	// docs
	Doc("set the curation status of the usermetric to {newStatus} which must be 'true' or 'false' ").
	Operation("updateUserMetricCurationStatus").
	Param(ws.PathParameter("key", "identifier of the usermetric").DataType("string")).
	Param(ws.QueryParameter("newStatus", "true/false curation status").DataType("bool")))

	// Endpoint for ChartHeader only (no JPG or LTMSettings)
	ws.Route(ws.GET("/usermetricheader").Filter(basicAuthenticate).Filter(filterCloudDBStatus).To(getUserMetricHeader).
	// docs
	Doc("gets a collection of usermetric header - in buckets of x headers - table sort is new to old").
	Operation("getUserMetricHeader").
	Param(ws.QueryParameter("dateFrom", "Date of last change").DataType("string")).
	Writes(UserMetricAPIv1HeaderOnlyList{})) // on the response

	// Count Chart Headers to be retrieved
	ws.Route(ws.GET("/usermetricheader/count").Filter(basicAuthenticate).Filter(filterCloudDBStatus).To(getUserMetricHeaderCount).
	// docs
	Doc("gets the number of usermetric headers for testing,... selection").
	Operation("getUserMetricHeader").
	Param(ws.QueryParameter("dateFrom", "Date of last change").DataType("string")))

	// ----------------------------------------------------------------------------------
	// setup the curator endpoints - processing see "entity_curator.go"
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

	// ----------------------------------------------------------------------------------
	// setup the status endpoints - processing see "entity_status.go"
	// ----------------------------------------------------------------------------------

	ws.Route(ws.POST("/status").Filter(basicAuthenticate).To(insertStatus).
	// docs
	Doc("creates a new status entity").
	Operation("createStatus").
	Reads(StatusEntityPostAPIv1{})) // from the request

	ws.Route(ws.GET("/status").Filter(basicAuthenticate).To(getStatus).
	// docs
	Doc("gets a collection of status").
	Operation("getStatus").
	Param(ws.QueryParameter("dateFrom", "Status Validity").DataType("string")).
	Writes(StatusEntityGetAPIv1List{})) // on the response

	ws.Route(ws.GET("/status/latest").Filter(basicAuthenticate).To(getCurrentStatus).
	// docs
	Doc("gets the current/latest status").
	Operation("getStatus").
	Writes(StatusEntityGetAPIv1{})) // on the response

	ws.Route(ws.GET("/statustext/{id}").Filter(basicAuthenticate).To(getStatusTextById).
	// docs
	Doc("gets the text for a specific status entity").
	Operation("getStatusText").
	Param(ws.PathParameter("id", "identifier of the version text").DataType("string")).
	Writes(StatusEntityGetTextAPIv1{})) // on the response

	// ----------------------------------------------------------------------------------
	// setup the version endpoints - processing see "entity_version.go"
	// ----------------------------------------------------------------------------------

	ws.Route(ws.POST("/version").Filter(basicAuthenticate).To(insertVersion).
	// docs
		Doc("creates a new version entity").
		Operation("createVersion").
		Reads(VersionEntityPostAPIv1{})) // from the request

	ws.Route(ws.GET("/version").Filter(basicAuthenticate).To(getVersion).
	// docs
		Doc("gets a collection of versions").
		Operation("getVersion").
		Param(ws.QueryParameter("version", "Version").DataType("string")).
		Writes(VersionEntityGetAPIv1List{})) // on the response

	ws.Route(ws.GET("/version/latest").Filter(basicAuthenticate).To(getLatestVersion).
	// docs
		Doc("gets the latest version").
		Operation("getVersion").
		Writes(VersionEntityGetAPIv1{})) // on the response

	// ----------------------------------------------------------------------------------
	// setup the telemetry endpoints - processing see "entity_telemetry.go"
	// ----------------------------------------------------------------------------------

	ws.Route(ws.PUT("/telemetry").Filter(basicAuthenticate).To(upsertTelemetry).
	// docs
		Doc("stores location,... of the call based on IP adress").
		Operation("post telemetry data").
		Reads(TelemetryEntityPostAPIv1{})) // from the request

	ws.Route(ws.GET("/telemetry").Filter(basicAuthenticate).To(getTelemetry).
	// docs
		Doc("gets a collection of versions").
		Operation("get All Telemetry Data").
		Param(ws.QueryParameter("createdAfter", "Telemetry created after").DataType("string")).
		Param(ws.QueryParameter("updatedAfter", "Telemetry last updated after").DataType("string")).
		Param(ws.QueryParameter("os", "Operating System").DataType("string")).
		Param(ws.QueryParameter("version", "GoldenCheetah Version").DataType("string")).
		Writes(TelemetryEntityGetAPIv1List{})) // on the response

	// all routes defined - let's go

	restful.Add(ws)

} // init()


// global declarations
const basicauth = "Basic_Auth"
const authorization = "Authorization"
const dateTimeLayout = "2006-01-02T15:04:05Z"
const (
	http_UnprocessableEntity = 422
)
const status_unprocessable = "Error - CloudDB Status does not allow processing the request"


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

func filterCloudDBStatus(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	ctx := appengine.NewContext(req.Request)

	if internalGetCurrentStatus(ctx) != Status_Ok {
		addPlainTextError(resp, http_UnprocessableEntity, status_unprocessable)
		return
	}

	chain.ProcessFilter(req, resp)
}


// Convenience functions for error handling
func addPlainTextError( r *restful.Response, httpStatus int, errorReason string ) {
	r.AddHeader("Content-Type", "text/plain")
	r.WriteErrorString(httpStatus, errorReason)
}

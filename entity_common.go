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

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"

	"github.com/emicklei/go-restful"
)


// ---------------------------------------------------------------------------------------------------------------//
// Common Structures and Functions for all CloudDB entities
// ---------------------------------------------------------------------------------------------------------------//
type CommonEntityHeader struct {
	Name        string       `datastore:",noindex"`
	Description string       `datastore:",noindex"`
	Language    string       `datastore:",noindex"`
	GcVersion   string
	LastChanged time.Time
	CreatorId   string
	Curated     bool
	Deleted     bool
}

// Internal Structure for Header
type CommonAPIHeaderV1 struct {
	Id          int64                `json:"id"`
	Name        string        `json:"name"`
	Description string      `json:"description"`
	GcVersion   string      `json:"gcversion"`
	LastChanged string      `json:"lastChange"`
	CreatorId   string      `json:"creatorId"`
	Language    string      `json:"language"`
	Curated     bool                `json:"curated"`
	Deleted     bool                `json:"deleted"`
}

func mapAPItoDBCommonHeader(api *CommonAPIHeaderV1, db *CommonEntityHeader) {
	db.Name = api.Name
	db.Description = api.Description
	db.Language = api.Language
	db.GcVersion = api.GcVersion
	db.LastChanged, _ = time.Parse(dateTimeLayout, api.LastChanged)
	db.CreatorId = api.CreatorId
	db.Curated = api.Curated
	db.Deleted = api.Deleted
}

func mapDBtoAPICommonHeader(db *CommonEntityHeader, api *CommonAPIHeaderV1) {
	api.Name = db.Name
	api.Description = db.Description
	api.GcVersion = db.GcVersion
	api.Language = db.Language
	api.LastChanged = db.LastChanged.Format(dateTimeLayout)
	api.CreatorId = db.CreatorId
	api.Curated = db.Curated
	api.Deleted = db.Deleted
}

func commonResponseErrorProcessing(response *restful.Response, err error) {
	switch {
	case appengine.IsOverQuota(err):
		// return 503 and a text similar to what GAE is returning as well
		addPlainTextError(response, http.StatusServiceUnavailable, "503 - Over Quota")
	case err == datastore.ErrNoSuchEntity:
		addPlainTextError(response, http.StatusNotFound, err.Error())
	default:
		addPlainTextError(response, http.StatusBadRequest, err.Error())
	}
}



// ignore missing fields error when mapping to Header struct
func isErrFieldMismatch(err error) bool {
	_, ok := err.(*datastore.ErrFieldMismatch)
	return ok
}
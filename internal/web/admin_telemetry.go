package web

import (
	"net/http"
	"strconv"

	"github.com/GoMudEngine/GoMud/internal/telemetry"
)

func adminTelemetry(w http.ResponseWriter, r *http.Request) {
	serveAdminTemplate(w, r, "telemetry.html", nil)
}

func adminTelemetryAPI(w http.ResponseWriter, r *http.Request) {
	serveAdminTemplate(w, r, "telemetry-api.html", nil)
}

// GET /admin/api/v1/telemetry
// Query params: category, itemId, mobId, roomId, zone, raceId, topic, date, dateFrom, dateTo,
//
//	groupby (mob|item|zone|room|date|category|race|topic), sort (asc|desc), limit (int)
func apiV1GetTelemetry(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	qb := telemetry.Query()

	if v := q.Get("category"); v != "" {
		qb = qb.Category(v)
	}
	if v := q.Get("zone"); v != "" {
		qb = qb.Zone(v)
	}
	if v := q.Get("date"); v != "" {
		qb = qb.Date(v)
	}
	if v := q.Get("dateFrom"); v != "" {
		qb = qb.DateFrom(v)
	}
	if v := q.Get("dateTo"); v != "" {
		qb = qb.DateTo(v)
	}
	if v := q.Get("itemId"); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			qb = qb.ItemId(id)
		}
	}
	if v := q.Get("mobId"); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			qb = qb.MobId(id)
		}
	}
	if v := q.Get("roomId"); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			qb = qb.RoomId(id)
		}
	}
	if v := q.Get("raceId"); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			qb = qb.RaceId(id)
		}
	}
	if v := q.Get("topic"); v != "" {
		qb = qb.Topic(v)
	}

	switch q.Get("groupby") {
	case "mob", "mobid":
		qb = qb.GroupBy(telemetry.GroupByMobId)
	case "item", "itemid":
		qb = qb.GroupBy(telemetry.GroupByItemId)
	case "zone":
		qb = qb.GroupBy(telemetry.GroupByZone)
	case "room", "roomid":
		qb = qb.GroupBy(telemetry.GroupByRoomId)
	case "date":
		qb = qb.GroupBy(telemetry.GroupByDate)
	case "category":
		qb = qb.GroupBy(telemetry.GroupByCategory)
	case "race", "raceid":
		qb = qb.GroupBy(telemetry.GroupByRaceId)
	case "topic":
		qb = qb.GroupBy(telemetry.GroupByTopic)
	}

	if q.Get("sort") == "asc" {
		qb = qb.SortAsc()
	} else {
		qb = qb.SortDesc()
	}

	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			qb = qb.Limit(n)
		}
	}

	results := qb.Results()
	if results == nil {
		results = []telemetry.Record{}
	}

	writeJSON(w, http.StatusOK, APIResponse[[]telemetry.Record]{
		Success: true,
		Data:    results,
	})
}

// DELETE /admin/api/v1/telemetry
// Query params: category, zone, raceId, topic, date, dateFrom, dateTo (all optional - omit to clear all)
func apiV1DeleteTelemetry(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	category := q.Get("category")
	zone := q.Get("zone")
	date := q.Get("date")
	dateTo := q.Get("dateTo")

	// A bare "date" param sets both bounds to that single day.
	// dateFrom/dateTo allow clearing a range.
	if date != "" && dateTo == "" {
		dateTo = date
	}
	if v := q.Get("dateFrom"); v != "" && date == "" {
		date = v
	}

	var raceId int
	if v := q.Get("raceId"); v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			raceId = id
		}
	}
	topic := q.Get("topic")

	telemetry.Clear(category, zone, date, dateTo, 0, 0, 0, raceId, topic)

	if err := telemetry.Save(); err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse[struct{}]{Success: true})
}

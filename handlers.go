package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type API struct{ store *Store }

func (api *API) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Path == "/" {
		if request.Method != http.MethodGet {
			writeError(writer, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		writeJSON(writer, http.StatusOK, map[string]string{
			"service": "campaign events API",
			"events":  "POST /events",
			"stats":   "GET /campaigns/{campaign_id}/stats",
			"listing": "GET /campaigns/{campaign_id}/events",
		})
		return
	}
	if request.URL.Path == "/events" {
		if request.Method != http.MethodPost {
			writeError(writer, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		api.postEvents(writer, request)
		return
	}
	const campaignPrefix = "/campaigns/"
	if strings.HasPrefix(request.URL.Path, campaignPrefix) {
		parts := strings.Split(strings.TrimPrefix(request.URL.Path, campaignPrefix), "/")
		if len(parts) == 2 && parts[0] != "" && parts[1] == "stats" && request.Method == http.MethodGet {
			writeJSON(writer, http.StatusOK, api.store.Stats(parts[0]))
			return
		}
		if len(parts) == 2 && parts[0] != "" && parts[1] == "events" && request.Method == http.MethodGet {
			api.getEvents(writer, request, parts[0])
			return
		}
	}
	writeError(writer, http.StatusNotFound, "endpoint not found")
}

func (api *API) postEvents(writer http.ResponseWriter, request *http.Request) {
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	var envelopes []eventEnvelope
	if err := decoder.Decode(&envelopes); err != nil || envelopes == nil {
		writeError(writer, http.StatusBadRequest, "request body must be a JSON array of events")
		return
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		writeError(writer, http.StatusBadRequest, "request body must contain one JSON array")
		return
	}
	events := make([]Event, 0, len(envelopes))
	for _, envelope := range envelopes {
		event, err := envelope.event()
		if err != nil || validateEvent(event) != nil {
			writeError(writer, http.StatusBadRequest, "batch contains an invalid event")
			return
		}
		events = append(events, event)
	}
	inserted, duplicates, err := api.store.AddBatch(events)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, errEventConflict) {
			status = http.StatusConflict
		}
		writeError(writer, status, err.Error())
		return
	}
	writeJSON(writer, http.StatusCreated, map[string]int{"accepted": inserted, "duplicates": duplicates})
}

func (api *API) getEvents(writer http.ResponseWriter, request *http.Request, campaignID string) {
	limit, offset := 100, 0
	var err error
	if value := request.URL.Query().Get("limit"); value != "" {
		limit, err = strconv.Atoi(value)
		if err != nil || limit < 1 || limit > 1000 {
			writeError(writer, http.StatusBadRequest, "limit must be between 1 and 1000")
			return
		}
	}
	if value := request.URL.Query().Get("offset"); value != "" {
		offset, err = strconv.Atoi(value)
		if err != nil || offset < 0 {
			writeError(writer, http.StatusBadRequest, "offset must be zero or greater")
			return
		}
	}
	writeJSON(writer, http.StatusOK, api.store.Events(campaignID, limit, offset))
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}

func writeError(writer http.ResponseWriter, status int, message string) {
	writeJSON(writer, status, map[string]string{"error": message})
}
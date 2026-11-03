package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

const testEvent = `[{"event_id":"evt-1","campaign_id":"cmp-1","contact_id":"ct-1","type":"opened","timestamp":"2026-08-10T09:15:04Z"}]`

func TestPostEventsIsIdempotent(t *testing.T) {
	api := &API{store: NewStore()}
	first := httptest.NewRecorder()
	api.ServeHTTP(first, httptest.NewRequest(http.MethodPost, "/events", bytes.NewBufferString(testEvent)))
	second := httptest.NewRecorder()
	api.ServeHTTP(second, httptest.NewRequest(http.MethodPost, "/events", bytes.NewBufferString(testEvent)))
	if first.Code != http.StatusCreated || second.Code != http.StatusCreated {
		t.Fatalf("unexpected statuses: %d, %d", first.Code, second.Code)
	}
	stats := api.store.Stats("cmp-1")
	if stats.Opened != 1 {
		t.Fatalf("opened count = %d, want 1", stats.Opened)
	}
}

func TestInvalidBatchDoesNotPartiallyCommit(t *testing.T) {
	api := &API{store: NewStore()}
	body := `[{"event_id":"good","campaign_id":"cmp-1","contact_id":"ct-1","type":"sent","timestamp":"2026-08-10T09:15:04Z"},{"event_id":"bad","campaign_id":"cmp-1","contact_id":"ct-2","type":"unknown","timestamp":"2026-08-10T09:15:04Z"}]`
	recorder := httptest.NewRecorder()
	api.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/events", bytes.NewBufferString(body)))
	if recorder.Code != http.StatusBadRequest || api.store.Stats("cmp-1").Sent != 0 {
		t.Fatalf("invalid batch was accepted: status=%d stats=%+v", recorder.Code, api.store.Stats("cmp-1"))
	}
}

func TestConcurrentDuplicatePostsCountOnce(t *testing.T) {
	api := &API{store: NewStore()}
	var wait sync.WaitGroup
	for i := 0; i < 20; i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			recorder := httptest.NewRecorder()
			api.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/events", bytes.NewBufferString(testEvent)))
		}()
	}
	wait.Wait()
	if got := api.store.Stats("cmp-1").Opened; got != 1 {
		t.Fatalf("opened count = %d, want 1", got)
	}
}

func TestStatsResponseShape(t *testing.T) {
	api := &API{store: NewStore()}
	recorder := httptest.NewRecorder()
	api.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/campaigns/cmp-unknown/stats", nil))
	var stats Stats
	if err := json.Unmarshal(recorder.Body.Bytes(), &stats); err != nil || stats.CampaignID != "cmp-unknown" {
		t.Fatalf("unexpected stats response: %s", recorder.Body.String())
	}
}
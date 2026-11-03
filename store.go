package main

import (
	"errors"
	"sort"
	"sync"
)

var (
	errInvalidEvent = errors.New("event has missing or invalid required fields")
	errEventConflict = errors.New("event_id already exists with different event data")
)

type Store struct {
	mu     sync.RWMutex
	events map[string]Event
}

func NewStore() *Store {
	return &Store{events: make(map[string]Event)}
}

// AddBatch validates and commits the whole batch under one lock.
// Replaying an identical event is idempotent; a reused ID with different data is rejected.
func (store *Store) AddBatch(events []Event) (inserted, duplicates int, err error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	pending := make(map[string]Event, len(events))
	for _, event := range events {
		if err := validateEvent(event); err != nil {
			return 0, 0, err
		}
		if previous, ok := pending[event.EventID]; ok {
			if !sameEvent(previous, event) {
				return 0, 0, errEventConflict
			}
			continue
		}
		pending[event.EventID] = event
	}
	for eventID, event := range pending {
		if existing, ok := store.events[eventID]; ok {
			if !sameEvent(existing, event) {
				return 0, 0, errEventConflict
			}
			continue
		}
	}
	for _, event := range pending {
		if _, ok := store.events[event.EventID]; ok {
			duplicates++
			continue
		}
		store.events[event.EventID] = event
		inserted++
	}
	return inserted, duplicates, nil
}

func sameEvent(left, right Event) bool {
	if left.EventID != right.EventID || left.CampaignID != right.CampaignID || left.ContactID != right.ContactID || left.Type != right.Type || !left.Timestamp.Equal(right.Timestamp) || len(left.Metadata) != len(right.Metadata) {
		return false
	}
	for key, value := range left.Metadata {
		if right.Metadata[key] != value {
			return false
		}
	}
	return true
}

func (store *Store) Stats(campaignID string) Stats {
	store.mu.RLock()
	defer store.mu.RUnlock()
	stats := Stats{CampaignID: campaignID}
	for _, event := range store.events {
		if event.CampaignID != campaignID {
			continue
		}
		switch event.Type {
		case "sent":
			stats.Sent++
		case "delivered":
			stats.Delivered++
		case "opened":
			stats.Opened++
		case "clicked":
			stats.Clicked++
		}
	}
	return stats
}

func (store *Store) Events(campaignID string, limit, offset int) []Event {
	store.mu.RLock()
	defer store.mu.RUnlock()
	result := make([]Event, 0)
	for _, event := range store.events {
		if event.CampaignID == campaignID {
			result = append(result, event)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.Before(result[j].Timestamp)
	})
	if offset >= len(result) {
		return []Event{}
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end]
}
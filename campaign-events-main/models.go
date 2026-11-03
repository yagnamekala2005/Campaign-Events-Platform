package main

import "time"

var allowedEventTypes = map[string]bool{
	"sent":      true,
	"delivered": true,
	"opened":    true,
	"clicked":   true,
}

type Event struct {
	EventID    string            `json:"event_id"`
	CampaignID string            `json:"campaign_id"`
	ContactID  string            `json:"contact_id"`
	Type       string            `json:"type"`
	Timestamp  time.Time         `json:"timestamp"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

type Stats struct {
	CampaignID string `json:"campaign_id"`
	Sent       int    `json:"sent"`
	Delivered  int    `json:"delivered"`
	Opened     int    `json:"opened"`
	Clicked    int    `json:"clicked"`
}

type eventEnvelope struct {
	EventID    string            `json:"event_id"`
	CampaignID string            `json:"campaign_id"`
	ContactID  string            `json:"contact_id"`
	Type       string            `json:"type"`
	Timestamp  string            `json:"timestamp"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

func (en eventEnvelope) event() (Event, error) {
	timestamp, err := time.Parse(time.RFC3339, en.Timestamp)
	if err != nil {
		return Event{}, err
	}
	return Event{
		EventID: en.EventID, CampaignID: en.CampaignID, ContactID: en.ContactID,
		Type: en.Type, Timestamp: timestamp, Metadata: en.Metadata,
	}, nil
}

func validateEvent(event Event) error {
	if event.EventID == "" || event.CampaignID == "" || event.ContactID == "" || event.Type == "" {
		return errInvalidEvent
	}
	if !allowedEventTypes[event.Type] {
		return errInvalidEvent
	}
	if event.Timestamp.IsZero() {
		return errInvalidEvent
	}
	return nil
}
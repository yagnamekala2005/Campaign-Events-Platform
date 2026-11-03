# Campaign Events Tracking Backend

Completed Go backend for Relay campaign event ingestion and campaign statistics.

## Structure

```text
ASSIGNMENT.md       supplied assignment
main.go             completed HTTP server
handlers.go         routing and JSON handlers
models.go           event models and validation
store.go            synchronized in-memory storage
main_test.go        backend tests
starter/            supplied starter skeleton and seed data
debugging/          supplied debugging program with minimal fixes
NOTES.md            decisions, scale memo, and marketer analysis
BUGS.md             debugging findings
AI_USAGE.md         AI disclosure
```

## Prerequisites

Install Go 1.22 or newer and ensure `go` is on `PATH`. No third-party dependencies are required.

## Run

From the repository root:

```powershell
go run .
```

The API listens on `http://localhost:8080`. If port 8080 is already in use, use the existing server or stop it with `Ctrl+C` before starting another copy.

## API

### POST `/events`

Accepts one JSON array of events. Required fields are `event_id`, `campaign_id`, `contact_id`, `type`, and an RFC3339 `timestamp`. Supported types are `sent`, `delivered`, `opened`, and `clicked`; metadata is optional.

The batch is atomic: if any event is malformed, invalid, or conflicts with an existing event ID, no new event in that batch is committed. An identical replay returns `201` with `duplicates` and does not change statistics. A conflicting event ID returns `409`.

```powershell
$body = Get-Content .\starter\seed\events.json -Raw
Invoke-RestMethod -Uri http://localhost:8080/events -Method Post -ContentType "application/json" -Body $body
```

Success response:

```json
{"accepted":1,"duplicates":0}
```

### GET `/campaigns/{campaign_id}/stats`

```text
http://localhost:8080/campaigns/cmp_summer_sale/stats
```

Example response:

```json
{"campaign_id":"cmp_summer_sale","sent":1,"delivered":1,"opened":1,"clicked":1}
```

Each unique event ID counts once. Arrival order and event timestamp do not affect counts. Unknown campaigns return zero counts with HTTP 200.

### GET `/campaigns/{campaign_id}/events`

Optional event listing, sorted by timestamp ascending. `limit` defaults to 100 and accepts 1-1000; `offset` defaults to 0.

```text
http://localhost:8080/campaigns/cmp_summer_sale/events?limit=20&offset=0
```

### Root discovery

`GET /` returns the available endpoint paths as JSON.

## Tests and verification

```powershell
go test ./...
go test -race ./...
go vet ./...
go build ./...
```

Run the supplied debugging program from its directory:

```powershell
Push-Location debugging
go run . events.jsonl
go run -race . events.jsonl
Pop-Location
```

Both outputs match `debugging/expected_output.txt`.

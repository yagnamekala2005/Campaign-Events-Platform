# Engineering notes

## Understanding

This service accepts campaign lifecycle events, deduplicates them by `event_id`, and exposes per-campaign counters. The implementation uses Go's standard HTTP server, `encoding/json`, and an in-memory store protected by a read/write mutex. The optional campaign event listing endpoint is included.

## Assumptions and counting semantics

- `event_id`, `campaign_id`, `contact_id`, `type`, and an RFC3339 `timestamp` are required.
- Only `sent`, `delivered`, `opened`, and `clicked` are accepted.
- Each unique `event_id` counts once, regardless of timestamp order or arrival time.
- Replaying the same event is a successful idempotent operation. Reusing an ID for different data returns `409 Conflict`.
- A batch is atomic: if any event is malformed, invalid, or conflicts, none of its new events are committed.
- Missing metadata is valid; metadata is not used in aggregation.
- Counters are event counts, not unique-contact counts. An event ID is the identity boundary.
- Stats for an unknown campaign return zeroes and HTTP 200.

## Ambiguities and product questions

The assignment does not define whether partial batches should be accepted, whether event IDs are globally unique, whether counters mean events or unique contacts, or whether timestamps may be future-dated. Questions for the product manager: should a conflicting replay be quarantined or rejected, should batches be partial, what retention period is required, and should stats support time zones, date windows, or unique-contact metrics?

## Priorities

Correct validation and idempotency come first because bad events permanently distort reporting. Atomic locking comes next because concurrent ingestion must not corrupt state. The stats endpoint follows because it is the primary read contract; event listing and pagination are useful operational visibility but secondary.

## Scale memo

The current in-memory map is appropriate for a small demo but loses data on restart, is limited to one process, and requires a full scan for every stats query. First improvement: move events and a uniqueness constraint into SQLite for durable local operation, or PostgreSQL when multiple instances and durable throughput are needed. Add an index/unique constraint on `event_id`, and an index on `(campaign_id, type)` plus timestamp if event listing and time windows are common.

At 100 million events per day, ingestion should be accepted by a durable queue and processed by horizontally scalable consumers. The database should partition or shard by campaign/time, and a unique event-ID constraint or idempotency table must remain authoritative across instances. Consumer retries require at-least-once-safe writes; dead-letter storage and replay tools make failures recoverable. Pre-aggregated counters can make reads cheap, but updates must be transactional with event deduplication or rebuilt from the event log to avoid drift.

The public API can remain the same for dashboard callers, but ingestion acknowledgement semantics should be made explicit if a queue is introduced: `201` can mean durably queued rather than fully aggregated, and a status/lag contract may be needed. The current synchronous API does not need to change for the first storage migration.

Operationally, record structured logs and metrics for accepted, duplicate, invalid, conflicted, delayed, and failed events; monitor queue lag, database latency, error rates, and counter reconciliation. Improve durability, uniqueness, indexes, and observability before adding distributed complexity. Postpone sharding, materialized analytics pipelines, and exactly-once claims until measured throughput and query needs justify them.

## Angry marketer scenario

The opened value can fall from 250,000 to 248,000 while delivered rises if the two dashboard snapshots are not monotonic views of one immutable dataset. Late or out-of-order events can raise delivered before open events arrive, while a correction or deletion can remove previously counted opens. A reprocessing job may deduplicate opens differently, an aggregation job may use a different time window, or cached dashboard panels may have different refresh times. Query consistency can also produce a mixed-time snapshot. A rising delivered value is not suspicious by itself: providers can report delivery later than send time, and delivery and open events have independent arrival delays. It is suspicious if the underlying immutable event set and query window are identical but the delivered count is not monotonic.

Other possibilities include duplicate cleanup, event backfills assigned to a different campaign, a timezone boundary, or an implementation bug such as replacing rather than incrementing counters, filtering by the wrong timestamp, or losing updates under concurrency. I would first compare raw event counts and unique IDs at both snapshot times, inspect ingestion/deletion/replay logs, and check cache timestamps and query windows. A stable event log with a unique-ID audit and a repeatable query distinguishes expected late/corrected data from a non-deterministic aggregation or race bug.

## What I completed / what I intentionally skipped

I completed the root HTTP service, validation, atomic batch policy, idempotent event storage, statistics, optional event listing, tests, supplied seed handling, the four supplied debugging fixes, the scale memo, the marketer analysis, and the required documentation. I intentionally kept in-memory storage because it is allowed by the assignment and avoids third-party dependencies. I did not add authentication, deployment, a frontend, or a queue because those are outside the requested scope.

## If I had another day

I would add persistent SQLite storage with a unique event-ID constraint, expand table-driven HTTP tests for every malformed input case, and add reconciliation tooling for aggregated counters. I would also benchmark batch sizes and document the exact ingestion acknowledgement semantics before introducing asynchronous queue processing.
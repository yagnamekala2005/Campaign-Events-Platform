# Debugging findings

The supplied debugging program had four independent bugs. Fixes were kept local to `debugging/main.go`.

1. **Deduplication reset for every batch.** `processBatch` created a new `seen` map for each 200-event batch, so duplicate IDs in different batches were counted again. The fix is the package-level `seenEventIDs` map, which is checked for the whole input. Verification: duplicate IDs in the 20,000-line fixture no longer inflate totals.
2. **Unique opens were not campaign-scoped.** `openedBy` used only `contact_id`, so one contact opening in campaign A prevented that contact from counting in campaign B. The fix keys the map by `campaign_id` and `contact_id`. Verification: the output matches the supplied expected unique-open counts for all campaigns.
3. **Daily delivery buckets used local time.** `Timestamp.Local()` made the day depend on the machine timezone instead of the event's UTC date. The fix uses `Timestamp.UTC().Format("2006-01-02")`. Verification: daily buckets match `expected_output.txt` and are timezone-independent.
4. **Worker updates raced and output could vary.** Worker goroutines incremented shared counters without synchronization. The fix protects `apply` with `statsMu`. Campaign and day names were already sorted before printing, so output is deterministic once counter races are removed. Verification: `go run -race . events.jsonl` completes without race reports, and repeated output matches `expected_output.txt`.

Commands run from this directory:

```text
go run . events.jsonl
go run -race . events.jsonl
```

Both commands completed successfully, and stdout matched `expected_output.txt` exactly.

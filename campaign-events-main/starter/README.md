# Campaign events starter

This directory is the original minimal skeleton supplied with the assignment. The completed service is implemented at the repository root; the starter is retained for reference.

## Requirements

- Go 1.22 or later.
- Standard library only.

## Run the starter

From this directory:

```text
go run .
```

The starter handlers intentionally return `501 Not Implemented`. Run the completed service from the repository root instead.

## Seed data

`seed/events.json` contains realistic provider traffic, including retries, conflicts, malformed timestamps, missing fields, and unsupported event types. The completed root service accepts or rejects the full batch according to its documented atomic-batch policy.

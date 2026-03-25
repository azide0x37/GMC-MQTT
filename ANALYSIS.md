# Analysis Checklist (Scored & Ranked)

Scoring: `2 = complete`, `1 = partial`, `0 = missing`.
Ranking: `1` is highest priority for improvement.

## Task Breakdown
1. [x] Add unit tests for serial parsing, discovery payloads, config updates, and state calculations.
2. [x] Expand README with discovery config + MQTT config update JSON format.
3. [x] Add config validation for required fields.

## Core Functionality (Current State)
1. [x] Serial communication with GMC-300s (Score: 2) - `gmc/gmc.go` implements commands and serial I/O.
2. [x] Periodic querying loop (Score: 2) - ticker-based polling in `cmd/main.go`.
3. [x] MQTT publish of state (Score: 2) - JSON state published to `state_topic` in `cmd/main.go`.
4. [x] Home Assistant discovery payloads (Score: 2) - discovery messages published in `cmd/main.go` + `mqtt/discovery.go`.
5. [x] Config loading + alignment of `publish_topic`/`state_topic` (Score: 2) - normalization mirrors legacy field.

## Missing / Partial Features (Ranked)
1. [x] MQTT config updates (temporary + permanent) (Score: 2)
   - Implemented with JSON updates and persistence.
2. [x] Persistent config updates (Score: 2)
   - Uses atomic write via temp file + rename.
3. [x] Basic reconnect + timeouts (Score: 2)
   - MQTT auto-reconnect and serial read timeouts + reopen logic.
4. [x] Runtime reconfiguration of query interval/topics (Score: 2)
   - Live updates to interval and subscription topics supported.

## Quality & Ops (Ranked)
1. [x] Tests beyond config load (Score: 2)
   - Added unit coverage for serial parsing, discovery payloads, config updates, and state calculations.
2. [x] Documentation accuracy (Score: 2)
   - `config.toml` fixed and README aligned for topics, discovery, and update JSON.
3. [x] Clear config defaults/validation (Score: 2)
   - Normalization and validation added for required fields.

## Report Card (Snapshot)
- Testing: `B`
- Documentation: `B`

## Skill: Make realtime scheduling byte-accurate and authoritative

Use this skill when changing the realtime protocol so scheduling, not encoding, owns packet budget decisions.

## Goal

Make `scheduler.go` the only place that decides whether a realtime candidate is included, deferred, or dropped for packet-size reasons.

After this change:

```text
planner builds candidates
-> candidates receive exact or defensible encoded byte costs
-> scheduler decides include/defer/drop using those byte costs
-> active.go encodes and sends included candidates only
-> active.go only fails on true encode/send errors, not budget policy
```

## Problem to remove

Do not allow this flow to remain:

```text
planner builds candidates
-> scheduler estimates candidate cost
-> scheduler includes candidate
-> active.go encodes candidate
-> active.go rejects hot-lane packet after scheduling
```

`active.go` must not silently skip already-scheduled hot-lane packets because of encoded size.

## Required files

Expected files:

```text
services/game-server/internal/protocol/realtime/candidate_size.go
services/game-server/internal/protocol/realtime/planner.go
services/game-server/internal/protocol/realtime/scheduler.go
services/game-server/internal/protocol/realtime/active.go
services/game-server/internal/protocol/realtime/scheduler_test.go
services/game-server/internal/protocol/realtime/active_test.go
```

Touch only additional files that are strictly required by compile errors or existing test placement.

## Step 1: Add candidate sizing

Create:

```text
services/game-server/internal/protocol/realtime/candidate_size.go
```

Add helpers that measure the encoded byte size of a realtime candidate before scheduling.

The initial implementation may encode the candidate wire packet to JSON and measure `len(bytes)`. Correctness is more important than estimator optimization for this pass.

Required responsibilities:

```go
func EstimateCandidateEncodedBytes(candidate RealtimeLaneCandidate) int
func EstimateWirePacketEncodedBytes(packet map[string]any) int
func CandidateExceedsHardPacketLimit(candidate RealtimeLaneCandidate) bool
```

Use existing candidate-to-wire-packet behavior rather than duplicating packet construction logic.

If encoding fails, return a conservative oversized value or expose the error through the smallest existing path. Do not silently return zero.

## Step 2: Replace fake schedule estimates

Find `scheduleRecordForCandidate()` in `planner.go`.

Replace broad packet-family estimates such as:

```go
EstimatedBytes: EstimatePacketBytes(packetFamily, 1, 0)
```

with exact candidate sizing:

```go
EstimatedBytes: EstimateCandidateEncodedBytes(candidate)
```

Keep existing lane, priority, delivery class, packet family, and metadata behavior unchanged.

Do not split `planner.go` in this task.

## Step 3: Move hot-lane hard-limit policy into scheduler

Update `scheduler.go` so `SelectSendPlan()` owns oversize decisions.

The scheduler should be able to distinguish:

```text
included
deferred_budget
dropped_oversize
```

For hot-lane records, if `EstimatedBytes` exceeds the configured hard encoded packet limit, classify the record as dropped because it is oversized.

The scheduler must not require `active.go` to perform a second size-policy check after scheduling.

Required handling:

```text
asteroids hot lane -> apply hard encoded packet byte limit
bullets hot lane -> apply hard encoded packet byte limit
other lanes -> preserve existing required/baseline behavior unless a current policy already says otherwise
```

Do not silently drop required non-hot packets unless existing code already has a defined failure/resync policy for that case.

## Step 4: Remove post-schedule hot packet rejection

In `active.go`, remove any logic that encodes a scheduled candidate and then rejects it because the encoded byte length exceeds a hot-lane packet limit.

Allowed behavior in `active.go`:

```text
candidate was not scheduled -> do not send
encode failed -> report encode failure
send failed -> report send failure
scheduled and encoded -> send
```

Disallowed behavior in `active.go`:

```text
scheduled packet is too large -> skip it here
```

That decision belongs in `scheduler.go`.

## Step 5: Make diagnostics truthful

Update diagnostics so scheduler results can distinguish:

```text
included
deferred_budget
dropped_oversize
send_failed
encode_failed
```

The important invariant is that logs and diagnostics must not report a packet as included if it was later skipped by `active.go` for budget reasons.

If the current diagnostics structure cannot represent all five states cleanly, add the smallest field or enum needed.

## Step 6: Add focused tests

Add or update tests in the realtime protocol package.

Required test coverage:

```text
small hot-lane candidate is included
oversize hot-lane candidate is dropped by scheduler
oversize hot-lane candidate is not dropped by active.go after scheduling
required non-hot candidate is not silently dropped by active.go for size policy
schedule record EstimatedBytes comes from encoded candidate size, not fake packet-family estimate
```

Most important regression test:

```text
Given a hot-lane candidate with EstimatedBytes over the hard limit,
SelectSendPlan records it as dropped_oversize,
and active encode/send is not responsible for dropping it.
```

Prefer direct unit tests around scheduler records and active send behavior. Avoid broad integration rewrites.

## Step 7: Report verification

After the edit pass, report:

```text
Changed files
What scheduler decision now owns
What active.go no longer owns
Existing project test command the user should run
Any compile/test blocker encountered
```

Use the existing server realtime test command from the game-server module root:

```bash
go test ./internal/protocol/realtime/...
```

Also report the broader existing server package test command:

```bash
go test ./internal/...
```

Do not run custom ad hoc verifier scripts for this task.

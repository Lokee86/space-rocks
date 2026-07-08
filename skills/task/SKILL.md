## Realtime Planner Extraction Skill

Use this skill when refactoring the server realtime protocol planner to reduce `planner.go` responsibility without changing behavior.

## Goal

Turn `services/game-server/internal/protocol/realtime/planner.go` from a mixed-responsibility protocol brain into a thin orchestration file.

The planner should coordinate candidate assembly. It should not own candidate policy, diagnostics, lane-specific full/delta construction, lifecycle packet construction, or lane quantizer placement.

This is a structural cleanup. Preserve behavior.

## Scope

Work only in:

* `services/game-server/internal/protocol/realtime/`

Expected production file additions:

* `candidate_types.go`
* `candidate_policy.go`
* `candidate_projection.go`
* `candidate_diagnostics.go`
* `lane_candidate_world.go`
* `lane_candidate_overlay.go`
* `lane_candidate_session.go`
* `lane_candidate_event.go`
* `lane_candidate_lifecycle.go`
* `quantize_overlay.go`
* `quantize_session.go`

Test files may be split only after production extraction is stable.

## Package Boundary Rule

Do not create new subpackages for this pass.

Keep all extracted planner pieces in package `realtime`.

Rationale:

* Planner code depends on core realtime types, packet structs, metadata, deltas, quantizers, lane state, hot-lane policy, and packet-family constants.
* Creating subpackages now would likely introduce import cycles or premature `core/common` packages.
* Same-package file seams make ownership visible without changing dependency shape.

The existing `quantize/` subpackage is fine. Do not move planner candidate code into it.

## Required Behavior Preservation

Preserve all current behavior, including:

* full packet fallback when no usable baseline projection exists
* delta omission when projection has not changed
* chained projection storage for delta candidates
* world hot-lane splitting
* asteroid and bullet hot-lane candidate creation
* asteroid and bullet lifecycle candidate creation
* hot-lane cohort state mutation through `sessionState`
* event batch creation without draining pending events
* candidate delivery class, priority, and packet-family selection
* schedule record construction
* candidate diagnostics output
* quantization behavior

Do not change scheduling behavior in this pass.

Do not change hot-lane chunking behavior in this pass.

Do not replace `RealtimeLaneCandidate.Full`, `RealtimeLaneCandidate.Delta`, or `RealtimeLaneCandidate.Projection` with typed interfaces in this pass.

## Target Planner Shape

After extraction, `planner.go` should mostly contain:

* `AssembleRealtimeLaneCandidates`
* `assembleRealtimeLaneCandidates`
* possibly `prepareRealtimeSendPlan`, unless moving it is a simple pure move

The private assembly function should append results from lane-specific builders:

* `buildWorldLaneCandidates`
* `buildOverlayLaneCandidates`
* `buildSessionLaneCandidates`
* `buildEventLaneCandidates`

## Extraction Order

### 1. Extract Candidate Types

Create `candidate_types.go`.

Move from `planner.go`:

* `RealtimeLaneCandidateKind`
* `RealtimeLaneCandidateKindFull`
* `RealtimeLaneCandidateKindDelta`
* `RealtimeLaneCandidateKindEventBatch`
* `RealtimeLaneCandidate`
* `RealtimeLanePlan`
* `RealtimeSendPrepared`

Do not change names or exported status.

### 2. Extract Candidate Policy

Create `candidate_policy.go`.

Move from `planner.go`:

* `packetFamilyForCandidate`
* `deliveryClassForCandidate`
* `priorityForCandidate`
* `scheduleRecordForCandidate`

Preserve exact logic.

Do not change delivery classes.

Do not change priorities.

Do not change `EstimatedBytes`.

Do not change `PayloadRef` behavior.

### 3. Extract Candidate Projection

Create `candidate_projection.go`.

Move from `planner.go`:

* `CandidateProjection`

Preserve exact logic.

Leave `CandidateMetadata` in `baseline.go` for this pass unless a later bounded task specifically moves metadata helpers.

### 4. Extract Candidate Diagnostics

Create `candidate_diagnostics.go`.

Move from `planner.go`:

* `CandidateWriteDiagnostics`
* `CandidateWriteDiagnosticsFor`
* `hotPacketCadenceForDiagnostics`
* `hotPacketCadenceLabel`
* `hotLaneModesForDiagnostics`
* `hotLaneCountsForDiagnostics`

Preserve exact diagnostics fields and values.

### 5. Extract Overlay Candidate Builder

Create `lane_candidate_overlay.go`.

Move overlay candidate construction out of `assembleRealtimeLaneCandidates`.

Add:

* `buildOverlayLaneCandidates(snapshot game.GameplayPresentationSnapshot, state RealtimeSessionState) []RealtimeLaneCandidate`

This function should own overlay:

* lane state lookup
* baseline readiness check
* next sequence calculation
* full packet build
* quantized full packet build
* baseline projection lookup
* full candidate fallback
* delta candidate creation
* unchanged projection omission
* chained projection creation

Return zero, one, or more candidates matching current behavior.

### 6. Extract Session Candidate Builder

Create `lane_candidate_session.go`.

Move session candidate construction out of `assembleRealtimeLaneCandidates`.

Add:

* `buildSessionLaneCandidates(snapshot game.GameplayPresentationSnapshot, state RealtimeSessionState) []RealtimeLaneCandidate`

This function should own session:

* lane state lookup
* baseline readiness check
* next sequence calculation
* full packet build
* quantized full packet build
* baseline projection lookup
* full candidate fallback
* delta candidate creation
* unchanged projection omission
* chained projection creation

Return zero, one, or more candidates matching current behavior.

### 7. Extract Event Candidate Builder

Create `lane_candidate_event.go`.

Move event candidate construction out of `assembleRealtimeLaneCandidates`.

Add:

* `buildEventLaneCandidates(snapshot game.GameplayPresentationSnapshot, state RealtimeSessionState) []RealtimeLaneCandidate`

Current behavior must remain:

* no candidate when `snapshot.PendingEvents` is empty
* event state sequence comes from `LaneEvent`
* candidate kind remains `RealtimeLaneCandidateKindEventBatch`
* packet is built with `BuildEventBatchPacket`
* pending events are not drained or mutated

### 8. Extract Lifecycle Candidate Helpers

Create `lane_candidate_lifecycle.go`.

Move asteroid and bullet lifecycle candidate construction out of inline world planning.

Add helpers:

* `buildBulletLifecycleCandidate(delta WorldWireDeltaPacket, state RealtimeSessionState) (RealtimeLaneCandidate, bool)`
* `buildAsteroidLifecycleCandidate(delta WorldWireDeltaPacket, state RealtimeSessionState) (RealtimeLaneCandidate, bool)`

Each helper should own:

* create/delete presence check
* lifecycle lane state lookup
* next lifecycle sequence calculation
* metadata lane replacement
* metadata sequence replacement
* lifecycle snapshot ID
* lifecycle snapshot kind
* chunk metadata
* lifecycle packet family
* candidate construction

Return `false` when no lifecycle creates or deletes exist.

### 9. Extract World Candidate Builder

Create `lane_candidate_world.go`.

Move world candidate construction out of `assembleRealtimeLaneCandidates`.

Add:

* `buildWorldLaneCandidates(snapshot game.GameplayPresentationSnapshot, state RealtimeSessionState, sessionState *RealtimeSessionState) []RealtimeLaneCandidate`

This function should own world:

* lane state lookup
* baseline readiness check
* next sequence calculation
* full packet build
* quantized full packet build
* baseline projection lookup
* full candidate fallback
* world delta creation
* hot-lane split
* lifecycle candidate extraction
* clearing lifecycle creates/deletes from world delta after lifecycle candidates are emitted
* asteroid hot candidate metadata stamping
* bullet hot candidate metadata stamping
* world delta change detection
* world projection chaining
* asteroid hot candidate creation
* bullet hot candidate creation

Preserve current `sessionState` mutation:

* if `sessionState != nil`, assign `sessionState.HotLaneCohorts = split.CohortState`

Do not change hot-lane policy.

Do not change hot-lane allowed/present conditions.

### 10. Move Overlay and Session Quantizers

Create `quantize_overlay.go`.

Move:

* `quantizeOverlayFullPacket`

Create `quantize_session.go`.

Move:

* `quantizeSessionFullPacket`

Preserve exact quantization field paths and error handling.

Do not alter `quantize_world.go`.

### 11. Shrink Planner Orchestration

Update `assembleRealtimeLaneCandidates` so it only coordinates:

* initialize candidate slice
* append world candidates
* append overlay candidates
* append session candidates
* append event candidates
* return `RealtimeLanePlan`

Preserve candidate order from current behavior:

1. world-related candidates, including lifecycle/hot candidates emitted by world builder
2. overlay
3. session
4. event

Candidate order is part of observable planner behavior through tests and scheduling.

## Test Handling

Do not split tests until production code extraction compiles.

After production extraction is stable, `planner_test.go` may be split into smaller same-package test files.

Recommended split:

* `planner_world_test.go`
* `planner_overlay_test.go`
* `planner_session_test.go`
* `planner_event_test.go`
* `planner_hot_lane_test.go`
* `candidate_policy_test.go`
* `candidate_diagnostics_test.go`

Do not rewrite test intent.

Do not remove coverage.

Move tests only when it reduces file size and preserves the exact assertions.

## Constraints

* Preserve behavior.
* No subpackages.
* No unrelated cleanup.
* No scheduler changes.
* No compact wire changes.
* No hot-lane chunking changes.
* No lifecycle sequencing changes.
* No candidate typing redesign.
* No package-wide architecture rewrite.
* If another file is required, touch only the smallest necessary additional file and report why.
* If the task balloons, stop.

## Verification

The expected verification command is:

* `go test ./services/game-server/internal/protocol/realtime/...`

If the module path requires running from `services/game-server`, use:

* `cd services/game-server && go test ./internal/protocol/realtime/...`

Report:

* changed files
* whether behavior was intended to remain unchanged
* test command result
* any tests moved or not moved
* any extraction that was intentionally deferred

# Participation And Joining
Parent index: [Gameplay Planning](./!INDEX.md)

## Purpose

This doc is the authoritative P4 planning owner for match participation and joining semantics.

It defines participation states, joining, leaving, AFK removal, spectating, reconnect, capacity, team assignment, participant counts, and result eligibility, plus the handoffs to lifecycle, modes, teams, lives, spawning, encounters, objectives, and results. It defines policy boundaries without claiming that implementation exists.

## Ownership Boundary

This doc owns planning for:

```text
match participation-state semantics
join eligibility and lobby-admission baseline
spectator-first late-entry policy
voluntary leave, forfeiture, and AFK-removal semantics
participation and result-eligibility defaults
participant-count change policy handoff
capacity/team-validation and re-entry/reconnect boundaries
match-over participation lock boundary
```

This doc does not own:

```text
transport, connection, room-membership, or capacity-reservation execution
reconnect-slot reservation execution or lifecycle timers/transitions
selectable mode policy
team assignment/balancing mechanics
life, death, elimination, respawn, or spawn placement/protection
encounter spawning, objective evaluation, or match-end orchestration
result locking/presentation or packet, storage, and UI schemas
```

Multiplayer Lifecycle owns execution of transport, room membership, capacity reservations, reconnect machinery, and transitions applying this policy. Modes own selectable participation, joining, re-entry, team-assignment, participant-count, and result-eligibility policy. Teams, lives/death/elimination/respawn, spawning, encounters, objectives, and results provide or consume their own authoritative facts rather than becoming alternate participation authorities.

## Settled Shared State Model

The shared participation states are:

```text
pending_spawn
active
pending_respawn
eliminated
spectating
disconnected
removed
```

They describe match participation independently from room membership and transport. A lobby member has not entered a match state; a spectator occupies room capacity without being an active participant. `pending_spawn` precedes the first appropriate active spawn; `active` is eligible for active gameplay; `pending_respawn` waits for a mode-approved return and shares the active AFK timer; `eliminated` records loss of the mode-defined active-life or eligibility path and normally transitions to `spectating`; `disconnected` is the preserved reconnectable participation state for Multiplayer Lifecycle V2; `removed` ends active evaluation while retaining result facts. In initial P4, a disconnect removes active participation and follows the interim `removed`/historical-facts behavior because active reconnect is not implemented.

The system must distinguish:

```text
room occupancy -> whether the member consumes capacity
match participation -> active, pending, eliminated, spectating, disconnected, or removed
result eligibility -> whether locked mode policy includes the member
```

## Settled Joining Model

Initial P4 joining is lobby-only; new players cannot enter an already-running match. Lobby eligibility is validated before match start through lifecycle admission.

Multiplayer Lifecycle V2 makes joining mode-defined and must support:

```text
new players; reconnecting players; spectators; invite/restricted-room admission
```

Joining is never permitted after the authoritative match-over lock. A locked/completed match is immutable: joins, leaves, reconnects, and participant changes cannot reopen, reverse, or mutate its outcome. GameOver/result viewing is a lifecycle path, not a completed-match join.

A permitted in-game joiner enters `spectating` first and becomes `active` only when the mode says the spawn/entry point is appropriate. Lifecycle executes admission and transitions; mode policy controls eligibility and activation. Late-join starting state and reconnect-slot reservation details are deferred to V2.

In-game join team assignment is Auto-balanced or mode-controlled. A join is rejected for mode, capacity, or team violations. No valid team assignment makes the join invalid; this should be rare and must be logged as a notable rule/configuration rejection. V2 reserves reconnect capacity, and reconnect restores preserved match state rather than using the fresh-join path. Reservation mechanism, lifetime, and exact capacity accounting are lifecycle implementation details.

## Spectating And Re-entry

Spectating is a capacity-consuming match state, not an implicit room leave. It is the normal first state for an admitted in-game joiner and the normal destination for an eliminated player. A spectator may return to active play only when mode policy allows it and the required entry condition is met.

Re-entry is not a fresh join: it preserves identity and historical facts and uses the mode-defined lives, elimination, team, objective, and result handoffs. Reconnect must not silently reset lives, active/eliminated state, pending-respawn state, team facts, or historical result facts unless mode policy explicitly says so. Participation does not define a universal re-entry trigger or spawn location.

## Leaving, Forfeiture, And Removal

Voluntary leaving is forfeiture plus immediate active removal. The participant stops contributing to active evaluation when authoritative leave is accepted, while historical participation, score, attribution, team, and other result facts remain available.

Team forfeiture occurs only when no remaining team members participate; one member leaving alone does not forfeit the team. Team rules provide normalized remaining-member facts and mode rules decide the outcome consequence.

AFK removal is forfeiture. Any relevant action resets the timer; active and pending-respawn share it; lobby and spectator states have no AFK timer. Action vocabulary, duration, warnings, and lifecycle execution are implementation-level decisions, but expiry must not become a non-forfeiting disconnect. Removal, disconnection, and forfeiture remain distinct facts even when all end active evaluation.

## Capacity, Team Validation, And Participant Counts

Spectators count against `player_capacity`, not a separate spectator-capacity pool. A future mode-specific spectator limit may be enforced within `player_capacity`, but it is not current policy. Lifecycle owns capacity occupancy, reservations, ordering, release, and enforcement, while distinguishing occupancy from active participant count. V2 joins require final authoritative validation of mode, capacity, admission, and team constraints; invitation or external assignment never bypasses game-server validation.

Participant changes may rescale objectives/targets and immediately affect encounter spawning unless the mode freezes participant count at match start. Objectives and encounters consume this resolved policy and own their calculations and spawn mechanics.

No minimum participation is required by default. Late, departed, or absent participants may still win and receive results or rewards under locked mode policy. Modes may explicitly override participation eligibility, including minimum active participation, presence, ranking, or reward conditions; eligibility is not inferred from final connection or room state.

## Match-Over Lock

The authoritative match-over lock is immutable. After it, new joins cannot enter, leaves/reconnects/spectator changes cannot revise the decision, and participant-count changes cannot rescale the locked result. Room membership, result viewing, return-to-lobby, and cleanup may continue under lifecycle policy, but participation and result facts cannot be reopened or reversed.

## System Handoffs

```text
room/lobby admission
-> lifecycle final validation
-> resolved mode participation/join policy
-> pending_spawn or spectating
-> mode-approved active entry or re-entry
-> normalized participation facts
-> teams / lives / spawning / objectives / encounters
-> locked eligibility and participation history
-> immutable MatchDecision and results
```

### Modes And Match Rules

Modes select joining, spectator, re-entry, participation, team-assignment, participant-count, and result-eligibility policy. They decide active entry, spectator return, participant rescaling/freezing, and result conditions; they do not execute transport, room membership, reservations, reconnect, spawning, or result locking.

### Multiplayer Session And Lifecycle

Lifecycle owns transport, room membership, admission, capacity reservations, reconnect machinery, timing, and removal execution. It applies resolved policy, preserves reconnect identity, and reports normalized facts; it does not invent mode, team, life, or result policy.

### Teams And Team Rules

Teams provide assignment/balance, relationship, and team-forfeiture facts from normalized participating members. In-game joining uses Auto-balanced or mode-controlled assignment; participation does not duplicate team authority.

### Lives, Death, Elimination, And Respawn

Lives/death/elimination/respawn owns life and respawn policy and consumes participation facts. It supplies life, death, pending-respawn, elimination, recovery, and re-entry facts. Eliminated players normally become spectators; later return is mode-approved re-entry, not a fresh join.

### Player Spawn Profiles

Spawn profiles own placement, safety, and protection. Participation supplies the eligible participant and entry context; an in-game joiner remains spectator until mode and spawn policy allow active entry.

### Objectives And Encounters

Objectives and encounters consume resolved participant-count policy. They own scaling, targets, calculations, and spawn mechanics; participation does not mutate them directly.

### Match Outcomes And Results

Outcomes lock the decision. Results consume historical participation, forfeiture, team, life, elimination, and eligibility facts and cannot reopen a match after later participant changes.

### Player Experience

The client may display joining, all seven participation states, AFK warnings, and result eligibility and collect relevant input. It cannot admit players, change authoritative participation, reset AFK without a recognized action, grant re-entry, or decide eligibility.

## Implementation Direction

The first P4 slice keeps joining lobby-only while establishing this policy seam:

```text
1. Define shared participation-state identifiers and normalized facts.
2. Validate lobby admission against mode, room, capacity, and team rules.
3. Separate occupancy, active participation, and result eligibility.
4. Apply voluntary leave and AFK removal as forfeiting active removal.
5. Route eliminated players to spectating without losing history.
6. Feed participant-count policy to objectives and encounters.
7. Preserve the immutable match-over lock and locked eligibility.
8. Reserve V2 seams for spectator-first joining, re-entry, state-restoring reconnect, and reconnect capacity.
9. Keep lifecycle execution, mode policy, and gameplay facts on their owning sides.
```

Lifecycle V2 adds new/reconnect admission, spectator admission within `player_capacity`, invite/restricted-room validation, reconnect reservations, and mode-approved spectator-to-active activation. Late-join starting state and reservation details remain V2 implementation decisions, not P4 product gaps. No product-level participation/joining decisions block P4 planning.

## Testing Direction

Important future checks:

```text
initial joining is lobby-only; all seven states are representable/distinguishable
occupancy differs from active participation and result eligibility
spectators consume capacity and eliminated players become spectators
spectators return active only by mode policy; in-game joiners enter spectator first
active entry waits for mode-approved spawn/entry; team assignment is Auto-balanced or mode-controlled
mode/capacity/team-invalid joins are rejected; no-valid-team rejection is rare, logged, and notable
V2 reconnect capacity is reserved within `player_capacity`; reconnect restores preserved state
voluntary leave immediately removes active evaluation and records forfeiture
team forfeiture requires no remaining participating team members
relevant actions reset AFK; active/pending-respawn share it; lobby/spectator have none; AFK is forfeiture
no minimum participation by default; late/departed/absent players may win and receive results/rewards
modes explicitly override participation/result eligibility
participant changes rescale objectives/targets and affect encounters unless count is frozen at start
locked/completed matches cannot reopen/reverse; post-lock joins are rejected
invite/restricted admission still receives final authoritative validation
```

## Related Docs

- [Gameplay Planning](./!INDEX.md)
- [Modes And Match Rules](modes-and-match-rules.md)
- [Teams And Team Rules](teams-and-team-rules.md)
- [Lives, Death, Elimination, And Respawn](lives-death-elimination-and-respawn.md)
- [Player Spawn Profiles](player-spawn-profiles.md)
- [Objectives And Objective Runtime](objectives-and-objective-runtime.md)
- [Encounter Lifecycle And Despawn](encounter-lifecycle-and-despawn.md)
- [Match Outcomes And Results](match-outcomes-and-results.md)
- [Multiplayer Session And Lifecycle](../platform/multiplayer-session-and-lifecycle.md)

## Remaining Implementation-Level Decisions

- Exact participation-state type, transition, event, snapshot, and storage fields.
- Exact separation/storage of occupancy, active participation, and result eligibility.
- Exact lobby admission and final-validation handoff shape.
- Exact AFK action vocabulary, duration, warnings, persistence, and timeout behavior.
- Exact pending-spawn behavior before first active spawn.
- Exact spectator admission/accounting and reservation release order within `player_capacity`.
- Exact V2 reconnect reservation capacity, lifetime, ownership, and expiry.
- Exact late-join starting state and spectator-to-active handshake.
- Exact invite/restricted admission fields and validation errors.
- Exact team-assignment request and rejection event shape.
- Exact participant-count handoff to objectives, targets, and encounters.
- Exact mode override for minimum participation and result/reward eligibility.
- Exact historical retention/result projection for leave, AFK removal, disconnect, and re-entry.
- Exact match-over lock event, persistence, and post-lock cleanup handoff.
- Exact package, packet, and storage boundaries at implementation time.

There are no remaining product-level participation or joining decisions blocking P4 system planning.

## Core Invariants

```text
Participation And Joining is the authoritative P4 planning owner.
Initial P4 joining is lobby-only; V2 joining is mode-defined.
V2 supports new, reconnecting, spectator, and invite/restricted-room admission.
Joining is never permitted after the authoritative match-over lock.
States are pending_spawn, active, pending_respawn, eliminated, spectating, disconnected, removed.
In-game joiners enter spectator first; active entry requires mode-appropriate spawn/entry.
Late-join starting state and reconnect reservations are deferred to V2.
Spectators consume `player_capacity`, not a separate spectator pool; eliminated players become spectators; spectators may return by mode policy.
Voluntary leave and AFK removal are forfeiture plus immediate active removal.
Team forfeiture requires no remaining participating team members.
Relevant actions reset AFK; active/pending-respawn share it; lobby/spectator have none.
No minimum participation is required by default; modes may override eligibility.
Late/departed/absent participants may still win and receive results/rewards under locked policy.
Participant changes may rescale objectives/targets and affect encounters unless count is frozen at match start.
Locked/completed matches are immutable and cannot reopen or reverse.
V2 rejects mode/capacity/team-invalid joins and logs rare no-valid-team rejection.
In-game team assignment is Auto-balanced or mode-controlled.
Reconnect restores preserved match state, not a fresh join.
Lifecycle owns execution, transport, room membership, reservations, and reconnect machinery.
Modes own selectable policy; gameplay systems provide or consume their own facts.
No product-level participation/joining decisions block P4 planning.
```

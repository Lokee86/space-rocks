<!-- policy-exempt: roadmap sequencing document -->

# Development Roadmap

Parent index: [Planning](./!INDEX.md)

## Purpose

This document defines the actionable development sequence for Space Rocks after the documentation and planning overhaul.

It coordinates implementation priority across technical foundations, public web presence, gameplay systems, platform systems, progression, multiplayer, and launch-facing surfaces.

Detailed ownership remains in the system-specific planning documents. This document owns sequencing, dependency order, phase gates, and priority relationships.

## Current Baseline

The current project has a useful vertical-slice baseline:

```text
local pilot/profile flow
match result presentation
stats refresh
authenticated multiplayer admission
room create/join/ready/start flow
single-player flow
player-data routing
initial devtools telemetry overlay
known gameplay packet pressure
```

The next development work should stop treating isolated feature slices as enough. Several future systems now depend on shared technical seams:

```text
packet budget
realtime protocol
mode rules
match result finalization
player-data contracts
progression grants
account trust
multiplayer lifecycle
public website surfaces
```

## Top Priorities

The top priority systems are:

```text
network observability and packet budget
realtime protocol architecture
```

Recommended order:

```text
1. Network observability and packet budget.
2. Realtime protocol architecture.
```

The packet-budget evidence checkpoint selected Phase P2 realtime protocol work as the current architectural next step. Remaining telemetry and logging work stays deferred until it is useful during P2 validation or after packet-size reduction.

Network observability and realtime protocol work are architectural blockers for serious gameplay expansion, larger multiplayer, enemies, bullet hell, and richer runtime events.

## Roadmap Rules

```text
Do not add entity-heavy gameplay before packet observability and realtime protocol work.

Do not add progression rewards before trusted match summaries, integrity decisions, idempotent grants, and player-data contracts are stable.

Do not add leaderboards before durable eligible results and public profile identity exist.

Do not add ranked PvP matchmaking before Combat Rating and ranked eligibility exist.

Do not turn the V0 devlog site into the launch website.

Do not start with protobuf before lanes, snapshots, deltas, priority policy, and packet ownership are proven.

Do not let devtooling-suite planning block the roadmap, but do implement telemetry required by packet-budget work.
```

## Phase P0 - Completed V0 Public-Devlog Baseline

### Goal

The first public Space Rocks web presence is complete as a V0 public-devlog baseline and must stay capped away from the future launch website.

### Scope

Completed baseline:

```text
homepage
devlog archive
individual devlog post pages
content-driven devlog entries
current Astro static devlog implementation
```

The V0 devlog site must not turn into the future launch website.

### Non-Goals

```text
accounts
account portal
CMS runtime
comments
newsletter backend
commerce
Steam linking
download access
leaderboards
profiles
support portal
appeal portal
admin portal
```

### Completion Criteria

```text
static build succeeds
homepage renders
archive renders
published posts render
draft posts are excluded
internal links resolve
static hosting is possible without runtime server
```

Remaining web polish and SEO/shareability planning live in the web planning docs, not in this roadmap sequence.

## Phase P1 - Network Observability And Packet Budget

### Goal

Make packet pressure measurable before adding systems that increase entity count, event count, or realtime state size.

This phase measures how far the current JSON/full-state packet model is from the lane-scoped realtime target.

Preferred frequent realtime packets should land around 300-600 bytes. Frequent packets over about 1KB should require justification, lower frequency, splitting, or later protocol work.

This phase is measurement and diagnostics, not optimization.

### Scope

```text
gameplay packet byte measurement
large-packet diagnostics
slow-write diagnostics
contributor counts
client inbound message byte tracking
devtools packet telemetry display
packet-pressure smoke scenario
```

### Completion Criteria

```text
large gameplay packets explain their contributors
slow writes include useful route context
client and server packet metrics can be compared
World Telemetry Overlay shows packet pressure
manual smoke can demonstrate packet growth as entities increase
no packet format has changed
no gameplay behavior has changed
```

### Decision Gate

```text
If normal gameplay packets are too large for the target budget:
-> proceed directly to realtime protocol architecture.

If packet size is acceptable but timing, jitter, or build/write cost is unclear:
-> harden runtime observability before protocol changes.

If packet size stays under budget and timing is clean:
-> platform/account work may move ahead before deeper optimization.
```
## Phase P2 - Realtime Protocol Architecture

### Goal

Replace the current full-state-per-tick delivery model with an explicit realtime protocol boundary.

### Ownership Rule

```text
networking/outbound owns delivery mechanics.
protocol/realtime owns delivery policy.
protocol/packetcodec owns byte representation.
```

### Scope

```text
server protocol/realtime package
services/game-server/internal/protocol/realtime/packets_generated.go
client protocol/realtime scripts
lane vocabulary
full snapshots
delta snapshots
baseline tracking
sequence numbers
create/update/delete records
priority policy
resync path
shadow verification
cutover
old state deletion
```

### P2 Validation Support

Deferred network telemetry and logging work can resume during P2 when it helps validate protocol changes:

```text
client inbound packet byte tracking when useful
World Telemetry Overlay packet display when useful
client/server packet comparison if needed
packet-pressure smoke checks for protocol changes
logging refinements needed to validate packet-size reduction
```

This support work belongs to P2 when it helps validate lanes, snapshots, deltas, baseline handling, packet-size improvements, or realtime protocol behavior.

### Deferred Next-Phase Codec And Compact Representation Work

```text
client packetcodec relocation
quantization rules
bit-packing rules
protobuf
binary/bitpacking work targets the new lane protocol, not old state
compact representation targets the new lane protocol, not old state
```

### Implementation Sequence

```text
1. Schema/data-sync packet families and planning doc normalization.
2. Server projections/full/delta/baselines plus non-draining snapshot/event API. Projection, scheduling, and encoding do not drain; shadow may peek/copy pending events but never drains.
3. Priority scheduler and budget planner.
4. Generic lane metrics and removal of legacy state packet warning noise.
5. Client protocol/realtime lane caches and compatibility read model.
6. Shadow encode/measure/parity with no send and no event drain. Shadow never drains; active `event_batch` drains only after socket write/enqueue success.
7. Runtime cutover behind a temporary dev-only switch.
8. Delete old `state` path and temporary switch.
9. Replace compatibility read model with lane-native presentation adapters.
10. Next phase: codec move plus compact/binary representation.
```

### Completion Criteria

```text
realtime protocol policy is separate from outbound delivery
full and delta snapshots exist
per-session baselines exist
shared-world and per-session overlay baselines exist
sequence handling exists
priority policy exists
lane/priority metrics exist
old state path is removed after cutover
high-frequency state no longer depends on one full combined packet every tick
stable event identity exists before event_batch cutover
event_batch duplicate suppression and control-path/event-drain ordering are defined after active socket write/enqueue success
```
## Phase P3 - Technical Release Foundation

### Goal

Make future work release-shaped instead of only editor/dev-runner-shaped.

### Scope

```text
verification and quality gates
build/release/environment matrix
compatibility, versioning, and migrations
operational readiness and failure modes
runtime performance and scale budget
observability, logging, and diagnostics
```

### Priority Order

```text
1. Verification and quality gates.
2. Build/release/environment matrix.
3. Compatibility, versioning, and migrations.
4. Operational readiness and failure modes.
5. Runtime performance and scale budget.
6. Logging and diagnostics hardening.
```

### Completion Criteria

```text
local development sanity gate exists
documentation and contract gate exists
local packaged single-player beta gate exists
dev-hosted multiplayer gate exists
hosted staging gate is defined
production candidate blockers are explicit
runtime-heavy features require measurement before release-shaped expansion
```

## Phase P4 - Player Experience Foundation

### Goal

Define what a match is, how rules are resolved, how results are finalized, and how player builds enter a match.

### Priority Order

```text
1. Modes and match rules.
2. Match outcomes and results.
3. Player build and loadouts.
4. Inventory and hangar foundation.
5. Player-experience presentation seams.
```

### Required Mode Slice

```text
survival_arcade
score_attack
ModePreset
RoomModeConfig
ResolvedMatchRules
configured lives
target_score for score_attack
mode identity in match result
```

### Required Match-End Slice

```text
EndOfMatchFlow
MatchSummary
MatchSummaryDispatcher
persistence slice
presentation-safe result slice
future progression slice
future achievement fact slice
```

### Required Build Slice

```text
ShipVariant
BuildEligibility
EligibleBuildOptions
LoadoutSelection
ResolvedPlayerBuild
RuntimeEquipmentState boundary
weapon-point rules
module slots
shield support
```

### Completion Criteria

```text
current play works through survival_arcade
score_attack proves the rules seam
match end locks once
result output is presentation-safe
build eligibility has an authoritative seam
runtime mutable state is not stored as loadout state
```

## Phase P5 - Trusted Results, Progression, And Player-Data Grants

### Goal

Make durable rewards safe, idempotent, and correctly routed.

### Priority Order

```text
1. Player-data contract enforcement.
2. MatchSummary to persistence-compatible result slice.
3. IntegrityEvaluation and EligibilityDecision.
4. GrantAward and Grant model.
5. Stable award_id and grant_id rules.
6. Idempotent player-data grant application.
7. XP, level, rank, and title derivation.
8. Earned Orebits grants.
9. Unlock grants.
10. Inventory item grants.
11. Achievement and milestone fact pipeline.
```

### Required Preconditions

```text
mode-aware trusted result facts
EndOfMatchFlow
player identity routing
result idempotency
integrity classification
player-data contract stability
```

### Deferred Until Later

```text
leaderboards
seasonal boards
public rankings
commerce-backed inventory
rare persistent drops
large economy sinks
```

### Completion Criteria

```text
progression emits GrantAward records
player-data owns storage routing and application
grants are idempotent
currency grants cannot duplicate on retry
debug/test/ineligible results do not silently enter normal rewards
guest, local profile, and authenticated account routes remain distinct
```

## Phase P6 - Account, Identity, Trust, And Moderation Foundation

### Goal

Finish the online identity and trust surface before expanding online platform systems.

### Priority Order

```text
1. Account display identity policy.
2. Strict display-name moderation.
3. Manual signup and login.
4. Email verification.
5. Online multiplayer block for unverified manual accounts.
6. Password reset and account recovery.
7. Token/session upgrade.
8. Google OAuth.
9. Provider linking and unlinking.
10. Account deletion and deactivation behavior.
11. Development-only auth bypass, build-flagged and environment-gated.
```

### Parallel Security/Admin Work

```text
game-integrity classification
room_title moderation
audit logs
appeal/review support for major actions
admin/devtool-like enforcement visibility
```

### Completion Criteria

```text
production online play requires Authenticated Account identity
display names are moderated
room titles are moderated
manual accounts cannot bypass verification
dev auth bypass cannot exist on live deployments
integrity decisions can classify result eligibility
```

## Phase P7 - Multiplayer Lifecycle V2

### Goal

Make room/session lifecycle robust before matchmaking tries to place players into rooms automatically.

### Priority Order

```text
1. Clarify SessionID, AccountID, MemberID, and PlayerID roles.
2. Add join order tracking.
3. Add disconnected member state.
4. Route active disconnect through the pause seam.
5. Add reconnect claim and active ship-control restoration.
6. Make Starting a real synchronized handoff state.
7. Add loading confirmation and timeout behavior.
8. Add queued join reservations.
9. Add mid-session join structure.
10. Add spectator capacity and lifecycle state.
11. Add member-local return-to-lobby.
12. Add GameOver result-viewing join behavior.
13. Add no-action timeouts outside lobby and queue states.
14. Split kick and room-lifetime ban.
15. Add owner transfer by join order.
16. Add lifecycle diagnostics.
```

### Completion Criteria

```text
disconnect is not leave
active reconnect works
Starting is a real lock/loading handoff
return-to-lobby is member-local
queued joins reserve capacity
kick and ban are separate
results do not mutate during reconnect, return, or cleanup
```

## Phase P8 - Matchmaking And Room Discovery

### Goal

Add browser, queue, assignment, and discovery after lifecycle/admission can safely receive assigned players.

### Priority Order

```text
1. API-server-owned matchmaking, search, and queue boundary.
2. Game-server-owned room registry summary boundary.
3. RoomDiscoverySummary.
4. room_title or room_name naming.
5. Joinable-room browser.
6. Initial filters.
7. Queue state and status.
8. Assignment target.
9. User confirmation timeout.
10. Assignment token or reservation semantics.
11. Fallback room creation.
12. Room title moderation handoff.
13. Social and Discord invite seams.
14. Hosted registry, region, and capacity prep.
```

### Deferred Until Ratings Exist

```text
ranked PvP matchmaking
rating-band matching
party rating aggregation
Combat Rating matchmaking
```

### Completion Criteria

```text
API server owns discovery and assignment
game server remains authoritative for room instances and final joins
browser lists requester-visible joinable rooms by default
queue can create fallback rooms quickly
assignment requires confirmation
final joins use the normal game-server lifecycle path
```

## Phase P9 - Metered Gameplay Expansion

### Goal

Add only enough gameplay expansion to support, test, and justify the systems around it.

The near-term purpose of gameplay expansion is not to build the final game content suite. The near-term purpose is to create enough real gameplay pressure to validate the portfolio systems around it:

```text
packet budget
realtime protocol
runtime performance gates
mode rules
match outcomes
progression grants
inventory/loadouts
leaderboards
website presentation
devtools and diagnostics
```

Actual broad gameplay expansion can wait until the surrounding systems justify and support it.

### Metering Rule

Gameplay expansion should be metered. Add the smallest useful amount of new gameplay content that proves a system seam.

Examples:

```text
one second baseline mode instead of many modes
one enemy family instead of a full bestiary
one boss prototype instead of a boss roster
one mission shape instead of a campaign
one loadout expansion path instead of a full arsenal
one rare persistent reward path instead of a full loot table
one bullet-pressure scenario instead of full bullet hell content
```

### Priority Order

```text
1. Safer ship, weapon, and module expansion.
2. Minimal enemy and encounter proof.
3. Minimal level/mission/content structure proof.
4. Minimal boss proof.
5. Bullet-pressure scenario.
6. Drones, mines, or radial timed area effects only when they prove specific systems.
7. Runtime rare drops.
8. Persistent rare drops only after progression and inventory grants are stable.
```

### Hard Gates

```text
No entity-heavy gameplay expansion before packet observability and realtime protocol work.

No progression-bearing gameplay expansion before trusted results, integrity, grants, and player-data routing are stable.

No leaderboard-facing gameplay expansion before ranking eligibility and board definitions exist.

No large content suite before the systems around it are proven.
```

### Completion Criteria

```text
new gameplay content proves at least one planned system seam
packet and runtime pressure are measurable
mode/result/progression effects are explicit
content remains small enough to replace or expand later
portfolio-relevant systems are strengthened by the gameplay slice
```

## Phase P10 - Leaderboards, Rankings, Seasons, And Campaign Structure

### Goal

Expose durable comparison and seasonal/campaign play after results are trusted and modes are mature enough.

### Priority Order

```text
1. Board definition model and source of truth.
2. public_profile_id behavior.
3. Ranking privacy settings.
4. Initial board catalog from mature modes.
5. Local profile rankings.
6. Online account rankings.
7. Post-match ranking impact.
8. In-game ranking browser.
9. Website leaderboard and public profile surfaces.
10. Seasonal and archived board lifecycle.
11. Skill aggregate boards with sample thresholds.
12. Pilot Rating.
13. Combat Rating.
14. Ad-hoc team boards.
```

### Season/Campaign Preconditions

```text
mode rules exist
mission/content structure exists
progression rewards exist
leaderboard/ranking policy exists where relevant
website can present season/campaign pages
```

### Exclusions

```text
no account resets
no seasonal exclusivity
no RMT seasonal pressure
no XP reward spam
```

### Completion Criteria

```text
boards derive from eligible durable facts
client and website do not maintain separate board catalogs
privacy settings affect public display
local and online ranking contexts remain distinct
season/campaign surfaces are FOMO-light and RMT-free
```

## Phase P11 - Launch Website And Commerce Platform

### Goal

Grow beyond the V0 devlog site into the full launch web and product platform surface.

This phase is separate from the V0 devlog site.

### Priority Order

```text
1. Product homepage.
2. Roadmap/status pages.
3. Media, gallery, lore, and deeper content.
4. CMS scaffold if still justified.
5. Account portal.
6. Ownership status.
7. Direct purchase surface.
8. Payment-provider handoff.
9. Steam linking presentation.
10. Steam ownership verification presentation.
11. Perpetual direct-download entitlement presentation.
12. Account-gated download access.
13. Support and recovery routes.
14. Legal, policy, and disclosure pages.
15. Launch analytics and conversion measurement.
16. Leaderboard, profile, season, and campaign website surfaces when their platform systems are ready.
```

### Required Dependencies

```text
account identity
commerce and economy policy
build/release artifact policy
entitlement model
support/admin visibility
safe failure states
```

### Completion Criteria

```text
launch homepage is product-first, not devlog-first
account portal exists
ownership status is visible
direct purchase flow has safe handoff
Steam ownership verification can grant direct-download entitlement
download access is account-gated
support/recovery routes exist
website is not authoritative for payment, entitlement, account, ranking, or moderation state
```

## Dependency Chain

The intended dependency chain is:

```text
P1 Network observability and packet budget
-> P2 Realtime protocol architecture
-> P3 Technical release foundation
-> P4 Player experience foundation
-> P5 Trusted results, progression, and player-data grants
-> P6 Account, identity, trust, and moderation foundation
-> P7 Multiplayer lifecycle V2
-> P8 Matchmaking and room discovery
-> P9 Metered gameplay expansion
-> P10 Leaderboards, rankings, seasons, and campaigns
-> P11 Launch website and commerce platform
```

P0 remains the completed public-devlog baseline and is not part of the active dependency chain.

Some phases can overlap, but dependency rules should not be violated.

## Related Docs

* [Network Observability And Packet Budget](domains/technical/network-observability-and-packet-budget.md)
* [Realtime Protocol Architecture](protocol/realtime-protocol-architecture.md)
* [Devlog Static Site](../services/web/devlog-static-site.md)
* [Website And Web Presence](domains/web/website-and-web-presence.md)
* [Verification And Quality Gates](domains/technical/verification-and-quality-gates.md)
* [Build Release And Environment Matrix](domains/technical/build-release-and-environment-matrix.md)
* [Modes And Match Rules](domains/gameplay/modes-and-match-rules.md)
* [Match Outcomes And Results](domains/gameplay/match-outcomes-and-results.md)
* [Player Build And Loadouts](domains/gameplay/player-build-and-loadouts.md)
* [Progression And Rewards](domains/gameplay/progression-and-rewards.md)
* [Account And Identity Systems](domains/platform/account-and-identity-systems.md)
* [Multiplayer Session And Lifecycle](domains/platform/multiplayer-session-and-lifecycle.md)
* [Matchmaking And Room Discovery](domains/platform/matchmaking-and-room-discovery.md)
* [Leaderboards And Rankings](domains/platform/leaderboards-and-rankings.md)
* [Season Format And Progression](domains/platform/season-format-and-progression.md)

## Notes

This roadmap is not a feature backlog.

It should remain a sequencing and dependency document. Detailed scope belongs in the owner documents for each domain, service, protocol, data, or devtools area.

When implementation changes make a planned system current, update the relevant current documentation instead of expanding this roadmap with implementation details.




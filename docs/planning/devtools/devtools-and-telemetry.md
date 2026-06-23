# Devtools Completed Suite

Parent index: [Devtools Planning](./!INDEX.md)

## Purpose

This document plans the completed Space Rocks developer tooling, admin tooling, telemetry, runtime diagnostics, durable-state test tooling, content tooling, and release-verification support.

It defines the final suite shape for current and future systems without duplicating detailed plans already owned elsewhere, especially realtime protocol lane planning.

## Overview

Space Rocks devtooling should become a controlled suite for inspection, mutation, simulation, diagnostics, content validation, and release verification.

The current implementation already provides strong runtime gameplay devtools:

```text
client devtools window
debug hotkeys
placement tools
spawn tools
respawn tools
freeze tools
score/lives tools
clear-entity tools
world telemetry overlay
player dev labels
server hitbox overlay
debug status output
debug shape catalog output
API-server Bruno smoke tests
data-sync validation and generation checks
```

The final suite must cover more than current match-runtime debugging. Future systems will require tooling for:

```text
mode rules
match outcomes
integrity classification
player-data routing
progression grants
inventory and hangar state
loadout eligibility
achievements and milestones
economy and commerce
admin/support corrections
content and source-of-truth validation
scenario and load testing
release gates
```

Devtools may inspect and mutate state, including durable state, but mutation must pass through the owning system’s real seam and must be gated, tagged, and auditable according to environment.

## Current status

Active planning.

Current runtime devtools are implemented around the Godot client, Go game-server devtools package, generated debug packets, and game-owned `export_devtools` seams.

Current gaps:

```text
future systems lack readout and mutation coverage
durable-state developer tooling is not complete
admin and developer gates are not separated
full telemetry needs a dedicated window instead of only an overlay
scenario/load-test tooling is not established
integrity/devtools taint propagation is not implemented
normal-event simulation for rewards and commerce is not implemented
```

## Ownership boundary

This document owns planning for the final developer/admin/tooling suite shape.

It owns:

```text
devtools suite categories
developer versus admin gating expectations
readout and mutation coverage expectations
telemetry window and configurable overlay direction
durable-state mutation rules for tools
normal-event simulation policy
scenario/load-test tooling direction
release-verification tooling expectations
```

It does not own:

```text
packet lane design
realtime protocol format
specific packet schemas
specific API endpoint design
physical player-data schema
reward formulas
economy balance
achievement definitions
admin enforcement policy
content file formats
test implementation details
```

Those details remain in their owning protocol, service, data, gameplay, platform, or technical planning docs.

## Core model

Final tooling is split into five categories:

```text
Runtime Developer Tools
Durable Developer Tools
Admin Tools
Content And Source-Of-Truth Tools
Verification And Release Tools
```

These categories may share infrastructure, but they do not share the same permission model.

```text
Developer tooling
-> diagnosis, reproduction, mutation, simulation, live-server investigation where allowed

Admin tooling
-> support, review, correction, repair, retry, enforcement-adjacent actions

Content tooling
-> authored data, generated outputs, tuning, source-of-truth validation

Verification tooling
-> repeatable checks, scenario runs, release gates, packaged-build confidence
```

## Authority rules

Client devtools can request and display. They must not become gameplay authority.

Server, player-data, API, progression, inventory, commerce, achievement, and integrity systems remain authoritative for their own state.

Allowed mutation pattern:

```text
developer/admin tool action
-> gated command or API route
-> owning domain/service seam
-> normal validation/idempotency/audit path
-> state mutation
-> readout, event, log, or result confirmation
```

Disallowed mutation pattern:

```text
developer/admin tool action
-> raw database write
-> bypass owning service
-> skip source tags
-> skip idempotency
-> silently create trusted durable state
```

## Tooling logs and audit

Every state-changing developer or admin action must be logged.

Required action log fields include actor identity, actor gate, environment, tool surface, action type, target kind, target id, source tag, reason where available, request accepted/rejected/applied, failure reason where applicable, and correlation id where available.

High-frequency telemetry samples should stay in metrics or devtools telemetry, not normal logs.

Audit-grade records are required for durable state, economy, enforcement, eligibility, admin corrections, devtools normal-source simulation, and production developer actions.

## Gate model

Admin and developer permissions must be separate.

| Gate          | Role                                          | Expected capability                                    |
| ------------- | --------------------------------------------- | ------------------------------------------------------ |
| Player        | Normal gameplay.                              | No devtools or admin mutation.                         |
| Support/Admin | Support, review, correction, retry, repair.   | Audited account/player-data/admin actions.             |
| Developer     | Diagnosis, reproduction, simulation, testing. | Controlled runtime and durable mutation where allowed. |
| Release/CI    | Automated verification.                       | Checks, scenarios, contract validation.                |
| Local Debug   | Local iteration.                              | Broadest devtools capability.                          |

Admins do not generally need gameplay mutation tools such as kill, spawn, freeze, invincibility, or arbitrary normal-reward simulation.

Developers may need dangerous tools on live or staging servers for investigation. Live developer access should require:

```text
developer role
environment allowlist
explicit reason when practical
audit log
source tag
session-limited or elevated access where practical
```

## Environment policy

| Environment              | Developer tools                               | Admin tools                  |
| ------------------------ | --------------------------------------------- | ---------------------------- |
| Local Development        | Enabled/open.                                 | Optional/local only.         |
| Local Packaged Beta      | Likely enabled for diagnostics.               | Not normally needed.         |
| Dev-Hosted Multiplayer   | Developer-gated.                              | Admin-gated where useful.    |
| Hosted Staging           | Strongly developer-gated.                     | Strongly admin-gated.        |
| Hosted Production Client | Disabled completely.                          | No client admin tools.       |
| Hosted Production Server | Exceptional developer-gated diagnostics only. | Narrow audited admin subset. |

Production client builds must not include devtools capability.

Production server builds may retain strongly gated developer/admin paths. These paths must not be reachable through normal player capability.

## Runtime developer tools

Runtime developer tools cover live gameplay and simulation state.

Current tools remain:

```text
devtools window
debug hotkeys
spawn and placement tools
respawn tools
freeze tools
counter tools
clear-entity tools
hitbox overlays
player labels
raw local/target telemetry
debug status readouts
```

Future runtime tools should cover both readout and mutation for:

```text
mode rules
objective state
match-end conditions
scoring policy
spawn profiles
encounter profiles
damage traces
shield and health state
runtime equipment state
ammo and cooldowns
radial effects
drop-table rolls
domain events
enemy, boss, and wave state
```

Readout should usually land before mutation. Mutation should be added when the owning system has a clean seam.

## Telemetry window and overlay

Full telemetry should move into a dedicated telemetry window.

The world overlay should remain a configurable, glanceable subset.

```text
Telemetry Window
-> detailed packet, runtime, entity, network, and scenario metrics
-> checkbox controls for what appears in the world overlay

World Overlay
-> selected live metrics only
-> small enough to use while playing
-> configured from the telemetry window
```

Likely telemetry window metrics:

```text
latest gameplay packet bytes
max gameplay packet bytes
average gameplay packet bytes where cheap
large packet warning count
server tick duration
simulation phase timings
build / encode / write duration
room count
player count
entity counts
event count
collision count
active stream count
RTT
client FPS
client visible entity counts
```

Likely overlay metrics:

```text
packet bytes
RTT
FPS
entity counts
warning indicators
```

Packet budget and protocol-lane details remain owned by the relevant technical and protocol plans.

## Durable developer tools

Durable developer tools cover systems that outlive a single match.

They should support inspection and controlled mutation through owning systems.

Required coverage:

```text
player-data route inspection
profile snapshot inspection
guest/local/account route testing
match result sink status
result retry/idempotency checks
GrantAward construction
inventory and hangar mutation through grants
wallet mutation through transactions
achievement and milestone event injection
commerce purchase/refund/sellback simulation
integrity classification inspection
```

Durable mutation is allowed through explicit developer/admin channels that use player-data ownership and domain-owned seams.

Examples:

```text
grant owned item through GrantAward
grant unlock through GrantAward
simulate earned currency through reward path
simulate shop purchase through commerce path
complete achievement through evaluator/completion path
repair profile through player-data route
retry failed match-result write through result-reporting path
```

## Normal-event simulation

Developer tools must support both explicit test sources and simulated normal sources.

```text
devtools_test_grant
-> obvious debug source
-> useful for testing grant plumbing

simulated normal source
-> pretends to be match_completion, achievement_completion, shop_purchase, or another real source
-> useful for testing normal downstream behavior

admin_grant / admin_adjustment
-> explicit support/admin correction source
```

Local and development environments may simulate normal earned rewards.

Deployment builds must not expose normal-source simulation to players or ordinary admins.

Developer-gated staging or live-server diagnostics may allow normal-source simulation only when explicitly authorized and audited.

Normal-source simulation must record that it was simulated, even when it exercises the same downstream path as a real match, achievement, milestone, purchase, or entitlement.

## Economy and commerce tools

Commerce tools must test real flows, not only arbitrary balance edits.

Required developer coverage:

```text
simulate match-earned Orebits
simulate achievement reward
simulate milestone reward
simulate shop purchase
simulate refund
simulate entitlement grant
simulate sellback
simulate receipt replay
simulate duplicate request
simulate insufficient funds
inspect wallet transactions
inspect receipts
inspect entitlement records
```

Admin economy actions should be narrower:

```text
audited correction
refund/reversal
support grant
repair failed purchase or entitlement state
```

Developer simulation and admin correction may share infrastructure but must use different gates and source tags.

## Admin tools

Admin tools should focus on support, correction, review, and repair.

Expected admin capabilities:

```text
inspect account state
inspect player-data state
inspect match result status
inspect leaderboard entry status
inspect room joinability and moderation state
inspect closed or non-joinable rooms for diagnostics
inspect season participation and seasonal reward grants
inspect website/account portal support state
retry failed internal writes
inspect integrity verdicts
inspect audit history
apply support correction through source-tagged grant
hide or restore leaderboard entries through owning moderation/admin paths
repair profile/account state through owning service
```

Admin tools should not generally expose:

```text
kill player
spawn bullets
spawn asteroids
toggle invincibility
freeze world
force arbitrary score/lives
simulate normal earned rewards
```

Any exception should be treated as developer tooling, not normal admin tooling.

## Enforcement and community tooling

Abuse, moderation, reports, appeals, room-name moderation, display-name moderation, Discord/community handoff, and enforcement consequences belong in [Abuse And Enforcement Admin](../domains/platform/security-and-admin/abuse-and-enforcement-admin.md).

Developer tooling may help inspect or reproduce enforcement signals.

Production action tools must use admin gates, case/audit records, and the owning enforcement paths.

Automated decisions must remain distinguishable from human admin actions.

## Content and source-of-truth tools

Data-sync is part of the broader tooling ecosystem, but it should be treated as content and source-of-truth tooling rather than only runtime devtools.

Current data-sync already covers:

```text
constants
packets
drop tables
player-data logical schema validation
```

Constants already cover many tuning surfaces, including runtime, gameplay, weapon, pickup, client presentation, shell, and lobby values.

Future content tooling should decide case by case whether new authored domains belong in data-sync, another source-of-truth pipeline, or separate content tooling.

Likely content/tooling concerns:

```text
mode presets
ship variants
weapon profiles
module profiles
shop offers
achievement and milestone catalogs
enemy and boss profiles
encounter profiles
mission or campaign content
visual/effect tuning
drop tuning
```

Required content tooling capabilities:

```text
validate source files
check generated drift
diff generated output
push generated output
show owning source
show generated consumers
preview client/server generated views
fail stale generated files in release gates
```

## Scenario and load-test tools

Runtime-heavy features need repeatable scenarios before they become release-shaped.

Scenario tooling should be able to:

```text
create room
select mode config
spawn controlled entity sets
spawn bullet pressure
spawn pickup/drop scenarios
spawn enemy/boss/wave scenarios later
drive bot/load-test input later
force match start/end churn
record packet and runtime metrics
export summary diagnostics
```

Scenario categories:

```text
packet pressure
entity pressure
bullet pressure
radial/effect pressure
enemy/boss pressure
larger-room pressure
spectator pressure
progression-heavy result flow
disconnect/reconnect churn
match start/end churn
```

These tools should feed runtime-heavy feature gates, staging readiness, and production candidate checks.

## Coverage requirements

Every implemented system should eventually declare its tooling coverage.

Minimum coverage fields:

```text
system
readout coverage
mutation coverage
owning mutation seam
allowed environments
admin gate
developer gate
trust-sensitive status
source tag behavior
audit/log behavior
test owner
documentation owner
```

A system is not tooling-complete if it only has raw state visibility but no controlled way to force, simulate, or repair important states during development.

A system is also not tooling-complete if it can mutate state without owning-seam routing, gates, source tags, or audit where required.

## Implementation sequence

1. Keep this plan indexed in `docs/planning/devtools/!INDEX.md` and update it as future tooling decisions become current implementation.

2. Add a tooling coverage registry table for current systems.

3. Harden existing runtime devtools around command ownership, gates, and action logging.

4. Add a telemetry window and make the world overlay configurable from it.

5. Implement packet and runtime telemetry requirements from the packet-budget plan without duplicating protocol-lane work here.

6. Add mode/rules/result readouts when mode and match-result systems mature.

7. Add integrity readouts for automation lane, result category, debug/devtools usage, taint, verdict, and eligibility.

8. Add durable developer tools for player-data, GrantAward simulation, inventory, wallet, achievements, and commerce.

9. Add normal-event simulation modes for rewards and commerce in development-safe environments.

10. Add scenario/load-test tooling for runtime-heavy feature gates.

11. Separate production admin tooling from developer tooling.

12. Add release-gate checks for production client devtools removal and production server admin/developer gating.

## Open decisions

```text
exact filename for the tooling coverage registry
exact telemetry window layout
exact command result or acknowledgement model
exact devtools action audit schema
exact source-tag vocabulary shared by devtools, admin, progression, and commerce
exact developer live-server access policy
exact scenario runner format
which current server devtools become production admin tools
which future content domains belong in data-sync versus separate content tooling
how normal-source simulation is represented in audit records
```

## Core invariants

```text
Final devtooling is not a parallel game.
Client devtools never become gameplay authority.
Mutation must route through the owning system’s real seam.
Durable mutation is allowed only through player-data/domain-owned paths.
Developer tooling and admin tooling have separate gates.
Normal-event simulation is allowed for development and testing but not normal deployment access.
Telemetry belongs primarily in a telemetry window, with a configurable world overlay subset.
Content and source-of-truth tooling is related to devtooling but remains its own tooling category.
Production client devtools capability must be absent.
Production server developer/admin paths must be gated, tagged, and auditable.
```

## Related docs

* [Devtools Planning](./!INDEX.md)
* [Devtools](../../devtools/!INDEX.md)
* [Devtools Authority And Seams](../../devtools/design/devtools-authority-and-seams.md)
* [Devtools Packet Protocol](../../devtools/design/devtools-packet-protocol.md)
* [Client Devtools](../../devtools/client/!INDEX.md)
* [Server Devtools](../../devtools/server/!INDEX.md)
* [API Server Devtools](../../devtools/api-server/!INDEX.md)
* [API Product Surface](../protocol/api-product-surface.md)
* [Network Observability And Packet Budget](../domains/technical/network-observability-and-packet-budget.md)
* [Observability Logging And Diagnostics](../domains/technical/observability-logging-and-diagnostics.md)
* [Build Release And Environment Matrix](../domains/technical/build-release-and-environment-matrix.md)
* [Verification And Quality Gates](../domains/technical/verification-and-quality-gates.md)
* [Game Integrity Policy](../domains/platform/security-and-admin/game-integrity-policy.md)
* [Abuse And Enforcement Admin](../domains/platform/security-and-admin/abuse-and-enforcement-admin.md)
* [Leaderboards And Rankings](../domains/platform/leaderboards-and-rankings.md)
* [Matchmaking And Room Discovery](../domains/platform/matchmaking-and-room-discovery.md)
* [Multiplayer Session And Lifecycle](../domains/platform/multiplayer-session-and-lifecycle.md)
* [Season Format And Progression](../domains/platform/season-format-and-progression.md)
* [Website And Web Presence](../domains/web/website-and-web-presence.md)
* [Progression And Rewards](../domains/gameplay/progression-and-rewards.md)
* [Inventory And Hangar](../domains/gameplay/inventory-and-hangar.md)
* [Shop, Commerce, And Economy](../domains/gameplay/shop-commerce-and-economy.md)
* [Achievements And Milestones](../domains/gameplay/achievements-and-milestones.md)
* [Player Build And Loadouts](../domains/gameplay/player-build-and-loadouts.md)
* [Data Sync And SSoT Pipeline](../../data/data-sync-and-ssot-pipeline.md)

## Notes

This plan intentionally avoids packet-lane design details. Future devtools packet movement belongs to realtime protocol planning.

“Devtools” in this document is a broad planning term. Current implementation docs under `docs/devtools/` should continue to describe actual implemented client, server, and API tooling.

Data-sync remains content/source-of-truth tooling even though developers use it heavily. Content development, tuning, contract validation, and generated-output management should not be forced into the same bucket as runtime debug commands.

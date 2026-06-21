# Game Integrity Policy

Parent index: [Platform Planning](../!INDEX.md)

## Purpose

This document plans the future game integrity policy for Space Rocks.

It defines the software layer that classifies sessions, matches, result facts, automation use, debug/devtools effects, and anti-farming signals before downstream systems consume those facts for progression, rewards, achievements, leaderboards, public match history, or review.

This is a platform-domain planning document. It is not the current trust-and-eligibility policy, not the realtime protocol architecture, and not an enforcement or ban policy.

## Overview

Space Rocks should treat game integrity as a software classification and routing problem.

The planned integrity flow is:

```text
client intent
-> realtime protocol / networking validation
-> authoritative game-server simulation
-> match end / MatchSummary
-> IntegrityEvaluation
-> ResultCategory + IntegrityVerdict + EligibilityDecision
-> player-data / progression / achievements / leaderboards / audit
```

The client never proves trusted result facts.

The game server owns simulation, scoring, match end, result construction, and initial integrity classification.

Downstream systems consume integrity decisions instead of recreating their own anti-cheat, automation, debug, or anti-farming rules.

The integrity system should support normal human play, declared bot play, TAS play, load testing, internal integrity testing, and debug/devtools use without collapsing them into one “cheat” category.

## Current status

Active planning.

Current implementation has strong structural foundations:

* the game server owns authoritative gameplay simulation
* the client sends intent, not result facts
* the game server owns collisions, damage, score, lives, deaths, match-over state, and match-result summaries
* player-data routes match results by identity and mode
* `result_id` provides match-result idempotency
* Guest and Local Profile identities cannot create online-trusted account facts
* production multiplayer requires Authenticated Account identity

Current implementation does not yet have:

* an integrity context object
* result categories
* automation lanes
* bot/TAS result lanes
* integrity verdicts
* downstream eligibility decisions
* debug/devtools taint propagation
* anti-farming evaluation
* integrity audit tags
* review handoff signals
* lane-specific leaderboard eligibility
* reward eligibility decisions derived from integrity evaluation

## Ownership boundary

This document owns planning for:

```text
integrity metadata
result categories
automation lanes
bot/TAS/load-test/debug classification
debug/devtools taint policy
integrity verdicts
eligibility decisions
anti-farming signals
result/reward/leaderboard eligibility routing
integrity audit tags
review handoff signals
```

This document does not own:

```text
packet lanes
transport choice
WebSocket delivery policy
WebRTC delivery policy
sequence numbers
baselines
stale update discard
resync behavior
codec migration
quantization
bit packing
protobuf migration
identity trust policy
account auth implementation
account bans
suspensions
appeals
display-name moderation
reward formulas
leaderboard ranking formulas
physical database schema
transport encryption
```

Realtime protocol hardening belongs to [Realtime Protocol Architecture](../../../protocol/realtime-protocol-architecture.md).

Current identity, mode, and trusted-fact eligibility belongs to [Trust And Eligibility Policy](../../../../domains/platform/trust-and-eligibility-policy.md).

Future account capability fields belong to account and identity planning.

Future review queues, bans, suspensions, appeals, and admin workflows belong to a separate abuse and enforcement admin plan.

## Relationship to realtime protocol

Realtime protocol planning owns how realtime data moves.

Game integrity owns how accepted server-owned gameplay and result facts are classified after protocol and networking have accepted or rejected packets.

These concepts are separate:

```text
Transport
-> WebSocket, WebRTC DataChannel, hybrid transport, or future transport

Protocol lane
-> reliable control, realtime state, event, slow world, debug/telemetry

Automation lane
-> human, bot, TAS, load_test, integrity_test, debug
```

A TAS run over WebSocket is still:

```text
AutomationLane = tas
Transport = websocket
```

A bot run over WebRTC is still:

```text
AutomationLane = bot
Transport = webrtc
```

A normal human match over WebRTC is still:

```text
AutomationLane = human
Transport = webrtc
```

Transport choice must not affect result trust.

A result is not more trusted because it came over WebRTC, and not less trusted because it came over WebSocket. Result trust comes from server authority, identity and mode eligibility, integrity classification, and downstream eligibility policy.

## Relationship to trust and eligibility

Trust and eligibility policy decides whether an identity, mode, and fact source can become online-trusted.

Game integrity policy decides whether a server-produced session or result should be classified as normal, automated, TAS, debug, test, suspicious, capped, review-required, or rejected for downstream use.

Correct split:

```text
Trust And Eligibility Policy
-> is this identity/mode/fact source eligible to become online-trusted?

Game Integrity Policy
-> what kind of session/result is this, and which downstream lanes may consume it?
```

Guest and Local Profile restrictions remain owned by trust policy.

Authenticated Account requirements remain owned by trust policy.

Local/offline import restrictions remain owned by trust policy.

Game integrity consumes those decisions and adds result-specific classification.

## Core architecture

The planned integrity architecture is:

```text
MatchSummary or current MatchResultSummary-compatible data
-> IntegrityContext
-> IntegrityEvaluation
-> ResultCategory
-> IntegrityVerdict
-> EligibilityDecision
-> downstream consumers
```

The first implementation should plug into match finalization before broad packet-abuse detection.

Runtime packet abuse counters can later feed integrity tags, but the initial seam should classify completed results and expose downstream eligibility.

## IntegrityContext

`IntegrityContext` is the input object for integrity classification.

Planned shape:

```text
IntegrityContext
- environment
- account_flags
- room_flags
- mode_id
- automation_lane
- result_category
- debug_tainted
- devtools_used
- automation_declared
- integrity_test
- custom_room
- private_room
- suspicious_tags[]
```

Field meanings:

* `environment` identifies development, test, staging, or production behavior.
* `account_flags` includes automation or test capabilities attached to the authenticated account.
* `room_flags` includes room-level settings that affect eligibility.
* `mode_id` identifies the mode whose rules produced the result.
* `automation_lane` identifies the declared control lane for the session or room.
* `result_category` is the final category stamped onto the result.
* `debug_tainted` records whether debug/devtools mutation touched trust-sensitive facts.
* `devtools_used` records debug/devtools involvement even when taint is suppressed in development.
* `automation_declared` records whether bot/TAS/automation was declared before the run.
* `integrity_test` records internal test classification.
* `custom_room` records custom or modified room status.
* `private_room` records private room status.
* `suspicious_tags[]` records integrity signals for audit, review, and downstream policy.

## Automation lanes

Automation lanes classify the intended control type of a session, room, or run.

Planned lanes:

```text
human
bot
tas
load_test
integrity_test
debug
```

Lane meanings:

* `human` - normal player-controlled session.
* `bot` - declared bot-controlled play.
* `tas` - declared tool-assisted or scripted assisted run.
* `load_test` - traffic or load simulation.
* `integrity_test` - internal verification run for integrity, scoring, result, or eligibility systems.
* `debug` - debug/devtools-driven run.

Automation lanes are not protocol lanes.

Automation lanes do not decide how packets move. They decide how the result is classified and which downstream consumers may count it.

## Result categories

`ResultCategory` is stamped onto match and result output.

Planned categories:

```text
normal_result
bot_result
tas_result
load_test_result
integrity_test_result
debug_result
```

Default lane-to-category mapping:

```text
AutomationLane.human
-> normal_result

AutomationLane.bot
-> bot_result

AutomationLane.tas
-> tas_result

AutomationLane.load_test
-> load_test_result

AutomationLane.integrity_test
-> integrity_test_result

AutomationLane.debug
-> debug_result
```

The server stamps result category.

The client or API must not retroactively reclassify a finished normal result as bot, TAS, debug, or test after the result is produced.

## Automation and TAS policy

Automation is not inherently cheating when it is explicitly declared, capability-authorized, and routed into an eligible automation lane.

Declared automation should be classified into bot, TAS, load-test, or integrity-test lanes.

Undeclared automation in a human lane is an integrity violation or review signal.

The planned software rule is:

```text
Declared automation
-> classify into bot/TAS/test lanes
-> exclude from normal human competitive lanes
-> optionally include in bot/TAS-specific lanes

Undeclared automation in human lanes
-> suspicious, review_required, or rejected
```

Automation-friendly accounts do not bypass game authority.

Automation-friendly accounts still submit intent to the authoritative game server. The server still owns simulation, scoring, match results, rewards, and eligibility.

## Account capability flags

Future account capability flags should allow automation lanes without making those accounts generally exempt from integrity rules.

Planned account flags:

```text
automation_allowed
bot_allowed
tas_allowed
integrity_test_account
```

Policy:

```text
account flag allows the lane
session or room declaration selects the lane
server stamps the result category
downstream eligibility decides what the result can affect
```

An automation-friendly account should not automatically be:

```text
human leaderboard eligible
ranked eligible
normal reward eligible
public competitive eligible
anti-cheat exempt
```

It should be eligible only for explicitly allowed automation lanes and downstream surfaces.

## Room and session declaration

Automation lane should be selected before match start.

Planned room or session option:

```text
automation_lane = human | bot | tas | load_test | integrity_test | debug
```

Rules:

```text
automation_lane defaults to human
non-human lanes require account or environment capability
lane must be locked before match start
lane must be stamped into the result category
lane cannot be changed after result finalization
```

Bot/TAS support should be designed as first-class declared lanes rather than hidden exceptions.

## Debug and devtools taint

`debug_tainted` is a planned integrity marker for facts affected by debug or devtools mutation.

It should be config-sensitive.

Development behavior:

```text
debug_tainted may be disabled, suppressed, or replaced with debug/test result categories
devtools may exercise gameplay, grant, presentation, and result paths
debug and test results should remain visibly classified where useful
```

Production behavior:

```text
debug/devtools mutation of trust-sensitive facts must taint or reject affected results
debug_tainted results must not silently enter normal online-trusted lanes
production trust-sensitive flows must treat debug/devtools effects strictly
```

`debug_tainted` should not be used for every non-human run.

Use explicit result categories instead:

```text
debug_tainted
-> debug/devtools mutation touched the run

bot_result
-> declared bot lane

tas_result
-> declared TAS lane

integrity_test_result
-> internal verification lane

load_test_result
-> traffic/load simulation lane
```

## Integrity verdicts

`IntegrityVerdict` is the result of integrity evaluation.

Planned verdicts:

```text
accepted
accepted_with_flags
capped
cooldown
review_required
rejected
```

Verdict meanings:

* `accepted` - no integrity issue found.
* `accepted_with_flags` - result may persist, but audit signals exist.
* `capped` - result may persist, but reward or scoring effects are capped.
* `cooldown` - source is temporarily ineligible for repeat reward or repeat public credit.
* `review_required` - public, competitive, or reward-sensitive effects should be held or excluded pending review.
* `rejected` - result, reward, or public/competitive fact is not eligible.

These verdicts are not account punishments.

Ban, suspension, appeal, and moderation behavior belongs to future enforcement/admin systems.

## Integrity tags

Integrity tags record evidence and classification details.

Initial planned tags:

```text
debug_tainted
devtools_used
automation_declared
automation_undeclared
bot_lane
tas_lane
load_test_lane
integrity_test_lane
duplicate_result
malformed_result
impossible_state
invalid_identity_mode
invalid_participant
invalid_room_state
invalid_mode_result
request_rate_exceeded
packet_flood
malformed_packet_spike
short_match_farming
repeat_reward_pattern
custom_room_ineligible
private_room_ineligible
manual_review
admin_override
```

Tags should be useful for:

```text
logs
audit records
review surfaces
test names
leaderboard exclusion explanations
reward eligibility decisions
future admin tooling
```

Tags should not expose sensitive internal detection detail to public clients unless intentionally surfaced through safe user-facing messaging.

## EligibilityDecision

`EligibilityDecision` is the downstream-facing routing object.

Planned shape:

```text
EligibilityDecision
- persist_private_stats
- grant_progression
- grant_currency
- grant_achievements
- human_leaderboard_eligible
- bot_leaderboard_eligible
- tas_leaderboard_eligible
- public_match_history_eligible
- requires_review
```

This object exists because one result may be valid for one consumer and invalid for another.

Example policy:

```text
normal_result
-> persist_private_stats: true
-> grant_progression: true
-> grant_currency: true
-> grant_achievements: true
-> human_leaderboard_eligible: true
-> bot_leaderboard_eligible: false
-> tas_leaderboard_eligible: false
-> public_match_history_eligible: true

bot_result
-> persist_private_stats: true
-> grant_progression: policy-dependent
-> grant_currency: policy-dependent
-> grant_achievements: policy-dependent
-> human_leaderboard_eligible: false
-> bot_leaderboard_eligible: true
-> tas_leaderboard_eligible: false
-> public_match_history_eligible: policy-dependent

tas_result
-> persist_private_stats: true
-> grant_progression: policy-dependent
-> grant_currency: likely false for normal economy
-> grant_achievements: policy-dependent
-> human_leaderboard_eligible: false
-> bot_leaderboard_eligible: false
-> tas_leaderboard_eligible: true
-> public_match_history_eligible: policy-dependent

load_test_result
-> persist_private_stats: false or internal-only
-> grant_progression: false
-> grant_currency: false
-> grant_achievements: false
-> human_leaderboard_eligible: false
-> bot_leaderboard_eligible: false
-> tas_leaderboard_eligible: false
-> public_match_history_eligible: false

integrity_test_result
-> persist_private_stats: internal-only
-> grant_progression: false outside test
-> grant_currency: false outside test
-> grant_achievements: false outside test
-> human_leaderboard_eligible: false
-> bot_leaderboard_eligible: false
-> tas_leaderboard_eligible: false
-> public_match_history_eligible: false

debug_result
-> persist_private_stats: local/dev/test only
-> grant_progression: false in production
-> grant_currency: false in production
-> grant_achievements: false in production
-> human_leaderboard_eligible: false
-> bot_leaderboard_eligible: false
-> tas_leaderboard_eligible: false
-> public_match_history_eligible: false
```

Exact progression, achievement, and bot/TAS public history policy remains a product decision. The software seam must support separate decisions.

## Anti-farming policy

Anti-farming is part of integrity evaluation.

It is not a ban system.

Anti-farming detects reward, leaderboard, achievement, or progression abuse that may happen through valid gameplay paths.

Initial signals:

```text
very short repeated matches
same participants repeatedly generating rewards
same mode or room settings repeatedly producing high yield
private-room reward loops
custom-room reward loops
low-input completion loops
rare reward repetition outside expected envelope
disconnect/reconnect exploitation
challenge completion repetition
objective completion repetition outside period policy
```

Possible outputs:

```text
accepted_with_flags
capped
cooldown
review_required
rejected
```

Anti-farming may block or reduce:

```text
currency grants
XP grants
rare drops
achievement progress
challenge completion
leaderboard eligibility
public match history eligibility
```

Anti-farming should not directly disable accounts or issue punishments.

## Runtime abuse signals

Runtime packet and request abuse belongs primarily to networking and realtime protocol hardening, but integrity may consume summarized signals.

Possible summarized signals:

```text
packet_flood
request_rate_exceeded
malformed_packet_spike
respawn_spam
pause_spam
target_spam
room_churn
auth_failure_spike
disconnect_reconnect_churn
```

The integrity system should not own packet validation or transport mechanics.

It may consume counters or flags produced by networking/protocol systems when evaluating result eligibility.

## Service-to-service protection relationship

Game integrity assumes public clients cannot directly submit trusted result, reward, leaderboard, or grant facts.

Internal write paths should be protected by service-to-service boundaries.

Relevant policy:

```text
game-server match result writes are internal-only
player-data internal routes require service authorization
Rails/API internal writes require service authorization
idempotency keys are required for result and grant writes
result payloads are not accepted from public clients
admin/devtools writes use explicit source tags
```

Exact service-auth mechanism belongs to API, player-data, and service integration docs.

## Downstream consumers

### Player-data

Player-data should receive result metadata needed for persistence, audit, and downstream eligibility.

Player-data should not decide that bot, TAS, debug, or test results are normal human results.

Player-data should preserve enough metadata for future review and reporting.

### Progression and rewards

Progression and rewards consume `EligibilityDecision`.

Progression should not independently decide that debug, TAS, bot, load-test, or suspicious results are normal reward sources.

Reward formulas remain owned by progression and reward planning.

Integrity owns whether a result is eligible to enter those formulas.

### Achievements and milestones

Achievements consume `EligibilityDecision` plus achievement-definition flags.

Some achievements may allow bot/TAS/local/debug progress for testing or special categories. Normal account achievements should not accidentally count debug or load-test results.

Achievement definitions should be able to express intended eligibility lanes.

### Leaderboards and rankings

Leaderboards consume lane-specific eligibility.

Normal human leaderboards exclude:

```text
bot_result
tas_result
load_test_result
integrity_test_result
debug_result
review_required results
rejected results
```

Bot leaderboards may include `bot_result`.

TAS leaderboards may include `tas_result`.

Ranking formulas and board presentation belong to leaderboard planning.

Integrity owns category and eligibility input.

### Public match history

Public match history should consume eligibility decisions.

A result may be private-stat eligible but public-history ineligible.

Examples:

```text
debug_result
-> public-history ineligible

integrity_test_result
-> public-history ineligible

review_required
-> public-history held or excluded

bot_result / tas_result
-> public-history allowed only with clear category labeling
```

### Audit and review

Integrity tags, verdicts, and result categories should feed future review tools.

The review system may later support:

```text
manual review
admin override
result correction
reward reversal
leaderboard removal
account enforcement
appeals
```

Those workflows are not owned by this document.

## Planned package placement

Likely game-server package:

```text
services/game-server/internal/integrity/
```

Possible files:

```text
context.go
automation_lane.go
result_category.go
verdict.go
tags.go
eligibility.go
config.go
match_evaluator.go
```

Initial implementation should remain small and concrete.

Avoid a broad framework before there are consumers.

## First integration point

The first integration point should be match finalization.

Planned flow:

```text
room/game reaches match over
-> MatchSummary or current MatchResultSummary-compatible object is resolved
-> integrity context is built
-> integrity evaluation runs
-> result category is stamped
-> verdict is produced
-> eligibility decision is derived
-> downstream result reporting receives result + eligibility metadata
```

Do not start by trying to inspect every packet.

Packet and request abuse counters can feed integrity later, after the result classification seam exists.

## Implementation sequence

1. Replace the obsolete anti-cheat planning stub with this game integrity policy.

2. Link protocol hardening to [Realtime Protocol Architecture](../../../protocol/realtime-protocol-architecture.md) instead of duplicating packet, lane, sequence, codec, WebSocket, or WebRTC concerns.

3. Define the software vocabulary:

```text
AutomationLane
ResultCategory
IntegrityVerdict
IntegrityTag
IntegrityContext
EligibilityDecision
```

4. Define the default eligibility matrix for:

```text
normal_result
bot_result
tas_result
load_test_result
integrity_test_result
debug_result
```

5. Define `debug_tainted` as config-sensitive:

```text
optional or suppressible in development
strict in production trust-sensitive flows
```

6. Define account capability flags:

```text
automation_allowed
bot_allowed
tas_allowed
integrity_test_account
```

7. Define room/session automation declaration:

```text
automation_lane
selected before match start
locked during the match
server-stamped into result category
```

8. Add the first game-server `integrity` package with vocabulary and pure evaluation behavior.

9. Attach default integrity metadata to match result finalization with no behavior change.

10. Add debug/devtools taint marking or debug result classification behind development-safe config.

11. Add automation-lane result classification for bot and TAS lanes.

12. Add `EligibilityDecision` and pass it to match result reporting.

13. Let progression, leaderboards, achievements, and player-data consume `EligibilityDecision` as those systems mature.

14. Add anti-farming signals as integrity tags and verdict inputs.

15. Add audit and review handoff once admin/review systems are planned.

## Testing direction

Important future tests:

```text
human lane produces normal_result
bot lane produces bot_result
tas lane produces tas_result
load_test lane produces load_test_result
integrity_test lane produces integrity_test_result
debug lane produces debug_result

bot_result is not human leaderboard eligible
tas_result is not human leaderboard eligible
debug_result is not public competitive eligible
load_test_result is not progression eligible
integrity_test_result is internal/test only

debug_tainted can be suppressed in development config
debug_tainted is strict in production trust-sensitive config

automation lane cannot be changed after match start
finished normal result cannot be retroactively reclassified as TAS
account without automation capability cannot select bot/TAS lane
integrity_test_account can select integrity_test lane in allowed environments

accepted verdict allows normal downstream routing
review_required blocks or holds public/reward-sensitive effects
rejected blocks trusted downstream effects

anti-farming short-match signal can produce capped, cooldown, review_required, or rejected
EligibilityDecision can allow private stats while blocking leaderboards
```

## Related docs

* [Platform Planning](../!INDEX.md)
* [Account And Identity Systems](../account-and-identity-systems.md)
* [Trust And Eligibility Policy](../../../../domains/platform/trust-and-eligibility-policy.md)
* [Account And Identity Current State](../../../../domains/platform/account-and-identity-current-state.md)
* [Realtime Protocol Architecture](../../../protocol/realtime-protocol-architecture.md)
* [Progression And Rewards](../../gameplay/progression-and-rewards.md)
* [Achievements And Milestones](../gameplay/achievements-and-milestones.md)
* [Match Outcomes And Results](../../gameplay/match-outcomes-and-results.md)
* [Modes And Match Rules](../../gameplay/modes-and-match-rules.md)
* [Leaderboards And Rankings](leaderboards-and-rankings.md)
* [Player Data And Persistence](player-data-and-persistence.md)
* [Multiplayer Session And Lifecycle](multiplayer-session-and-lifecycle.md)

## Open decisions

Implementation-shape decisions remain:

* exact `IntegrityContext` field names
* exact `EligibilityDecision` field names
* exact account flag names
* exact room/session option shape for `automation_lane`
* whether `automation_allowed` is enough or separate `bot_allowed` and `tas_allowed` flags are required
* exact development config names for debug taint suppression
* exact production config names for strict integrity behavior
* exact persistence shape for result category, verdict, tags, and eligibility metadata
* exact anti-farming thresholds
* exact bot/TAS progression policy
* exact bot/TAS public match history policy
* exact bot leaderboard and TAS leaderboard product behavior
* exact review handoff shape
* exact relationship between admin override and eligibility recalculation

These are not open policy questions:

* whether the client proves result facts
* whether bot/TAS runs are normal human results
* whether debug/devtools results silently enter normal competitive lanes
* whether transport choice affects result trust
* whether progression, achievements, leaderboards, and player-data should consume shared eligibility decisions

## Core invariants

```text
The client never proves result facts.

The server owns simulation, scoring, match end, and result classification.

Protocol hardening is an adjacent realtime planning dependency.

Transport choice must not affect result trust.

Automation is allowed only when declared and capability-authorized.

Automation, bot, TAS, load-test, integrity-test, and debug runs use explicit lanes and result categories.

Bot and TAS results are not normal human results.

Debug/devtools taint is optional or suppressible in development but strict for production trust-sensitive flows.

Eligibility is downstream-specific.

One result can be private-stat eligible but leaderboard-ineligible.

Progression, achievements, leaderboards, and player-data consume integrity decisions instead of reimplementing them.

Anti-farming changes reward, public, or competitive eligibility. It does not directly punish accounts.

Bans, suspensions, appeals, and manual enforcement belong to a separate future admin/enforcement domain.
```

## Notes

This document intentionally avoids the old combined “game integrity polciy” ownership model.

Trust policy owns whether an identity and source may create online-trusted facts.

Realtime protocol planning owns how packets and state move over WebSocket, WebRTC, or later transports.

Game integrity owns how server-produced sessions and results are classified for downstream use.

The practical software goal is not to define cheating as a conduct issue. The goal is to ensure every match result carries enough metadata for progression, rewards, achievements, leaderboards, public history, automation lanes, TAS lanes, debug flows, and future review systems to make safe and consistent decisions.

# Abuse And Enforcement Admin

Parent index: [Security And Admin](./!INDEX.md)

## Purpose

This document plans the abuse and enforcement admin domain for Space Rocks.

It defines the future system that receives moderation, trust, integrity, leaderboard, room, report, and support signals, then turns them into cases, automated decisions, enforcement actions, appeals, audit history, and admin/support corrections.

This is a platform-domain planning document. It is not the game-integrity classifier, not the trust-and-eligibility policy, not the realtime protocol hardening plan, not the account-auth policy, and not Discord platform enforcement implementation.

## Overview

Abuse and enforcement admin owns the consequence layer.

The planned flow is:

```text
trust, integrity, moderation, report, support, or admin signal
-> case or direct automated decision
-> moderation/enforcement decision
-> affected object state change
-> audit event
-> optional appeal or admin review
-> optional reversal or restoration
```

The core split is:

```text
Trust And Eligibility Policy
-> decides whether an identity, mode, or fact source can become trusted

Game Integrity Policy
-> classifies gameplay, automation, debug, anti-farming, and result-integrity state

Abuse And Enforcement Admin
-> decides what happens after suspicious, rejected, disputed, abusive, or moderated behavior
```

The domain should support automated moderation authority by default.

Keyword rules, classifiers, and LLM moderation may silently allow, deny, hold, hide, reject, invalidate, warn, or create cases without human approval first. Human review is normally reserved for appeals, exceptional escalation, suspected model failure, or admin intervention.

The system must still log major automated actions well enough to support later appeal, review, reversal, and audit.

## Current status

Active planning.

Current implementation and planning already have related foundations:

* authenticated accounts are required for production multiplayer admission
* trust and eligibility policy defines identity and trusted-fact boundaries
* game integrity policy defines future result classification, integrity verdicts, debug taint, automation lanes, anti-farming signals, and review handoff signals
* account and identity planning requires strict display-name moderation
* matchmaking and room discovery planning identifies room title/name moderation as abuse/enforcement owned
* API product surface planning reserves future admin, support, and moderation surfaces
* social/community planning defines Discord-first social systems
* in-client Discord chat, DMs, and voice are part of the full social target
* official Discord community moderation may create Space Rocks enforcement signals

Current implementation does not yet have:

```text
abuse/enforcement case model
review queues
automated moderation decision records
room_name moderation enforcement
room invalidation for moderated names
appeal request flow
appeal cooldowns
admin enforcement tooling
support correction cases
enforcement audit history
email/website enforcement notice flow
result/reward/leaderboard restoration workflow
```

## Ownership boundary

This document owns planning for:

```text
abuse cases
review queues
automated moderation decisions
keyword moderation
classifier moderation
LLM moderation authority
display_name enforcement
room_name enforcement
room invalidation after name moderation failure
room repair through name correction
warnings
restrictions
suspensions
bans
appeals
appeal cooldowns
case audit history
admin enforcement actions
support correction actions
result invalidation and restoration decisions
leaderboard hide/remove/restore decisions
reward revocation and restoration decisions
report intake routing
Discord/community handoff routing
website/email enforcement communication
official Discord community moderation handoff routing
Discord-powered in-client communication report routing
Space Rocks-side consequences from Discord/community moderation signals
social/communication restrictions
```

This document does not own:

```text
gameplay anti-cheat detection
packet-abuse detection
realtime protocol hardening
transport encryption
trust eligibility rules
account authentication
account provider linking
leaderboard ranking formulas
reward formulas
mode rules
room lifecycle implementation
Discord platform enforcement implementation
Discord ToS enforcement
Discord server tooling implementation details
legal retention/anonymization policy
normal season/ruleset maintenance recalculation
exact HTTP endpoint shapes
exact packet schemas
```

Game integrity owns suspicious result classification, automation lanes, debug taint, anti-farming signals, and eligibility handoff.

Trust and eligibility owns identity/mode/fact-source trust policy.

Account and identity owns account identity, provider identity, auth, account display identity policy, and display-name validation constraints.

Matchmaking and room discovery owns room browser, queue, assignment, and requester-safe discovery projection, but abuse and enforcement admin owns room name moderation consequences.

API product surface owns the exact future internal admin/support/moderation API surface.

Discord platform enforcement remains Discord-owned.
Space Rocks owns moderation responsibility for official Space Rocks Discord community operations.
Space Rocks owns report routing and Space Rocks-side enforcement consequences for Discord-powered communication rendered inside Space Rocks client or website surfaces.

## Signal sources

Abuse and enforcement admin should be able to consume signals from:

```text
trust and eligibility failures
game-integrity verdicts
anti-farming signals
leaderboard/result anomalies
reward/progression abuse signals
display_name moderation
room_name moderation
future profile/public-text moderation
user reports
Discord/community handoff
official Discord server moderation actions
Discord-powered in-client chat reports
Discord-powered in-client DM reports
Discord-powered in-client voice reports
Discord invite abuse
Discord lobby or social abuse
community moderator escalations
Discord community role or membership abuse signals
support/admin manual reports
automation/TAS/debug lane misuse
future payment/shop/RMT abuse if commerce expands
```

Not every signal needs a human-visible case.

Routine moderation may resolve automatically while still writing audit events.

Major or repeated signals should create or update cases so later appeal, review, and enforcement history can be reconstructed.

## Core model

The central object is an enforcement case.

Planned shape:

```text
Case
- case_id
- source_system
- subject_account_id
- public_player_id, where relevant
- room_creator_account_id, where relevant
- affected_object_type
- affected_object_id
- category
- severity
- current_status
- evidence_references[]
- automated_outputs[]
- decision
- enforcement_action
- appeal_state
- audit_events[]
```

Affected object types may include:

```text
account
display_name
room_name
room
match_result
leaderboard_entry
reward_grant
progression_grant
profile_text
report
support_correction
discord_message_context
discord_voice_context
discord_dm_context
discord_invite_context
discord_lobby_context
community_server_member
community_role
```

Useful case statuses:

```text
created
auto_resolved
invalidated
held
actioned
appealed
under_review
reversed
closed
```

Cases should support multiple related issues when an appeal bundles several active enforcement actions into one review request.

## Categories

Planned categories:

```text
cheating_or_gameplay_manipulation
exploit_abuse
farming_or_reward_abuse
leaderboard_manipulation
account_abuse
community_abuse
display_name_violation
room_name_violation
impersonation
harassment_or_unsafe_public_text
in_game_communication_abuse
invite_abuse
voice_or_chat_abuse
spam_or_scam_behavior
automation_lane_misuse
support_correction
stale_or_incorrect_data_correction
false_positive_or_reversal
```

Categories are routing and audit labels. They should not imply every action is punitive.

Support/admin corrections are manual platform actions, but they are not abuse enforcement unless the underlying case is abuse-related.

## Severity

Planned severity shape:

```text
informational
low
medium
high
critical
```

Meanings:

* `informational` records a signal without action.
* `low` supports deny, hide, hold, repairable invalidation, or warning.
* `medium` supports warnings, temporary action restrictions, result removal, room invalidation, or leaderboard hiding.
* `high` supports broader restrictions, temporary suspension, reward/result invalidation, and appeal-ready logging.
* `critical` supports permanent bans, security escalation, severe moderation action, or exceptional admin handling.

Exact thresholds are still open.

## Automated moderation authority

Automated moderation is a first-class authority, not only a recommendation system.

The planned moderation stack is:

```text
keyword and phrase rules
-> classifier analysis
-> LLM moderation
-> report fallback
-> appeal or admin review when needed
```

Keyword rules should handle deterministic blocked terms, phrases, and obvious violations.

Classifiers should produce category, confidence, severity, and routing tags.

LLM moderation may silently decide routine moderation outcomes. It may directly allow, deny, hold, hide, reject, invalidate, warn, or create a case.

LLM output is a system decision. It should record:

```text
model or policy version
input surface
category
confidence
severity
rationale
decision
affected object
timestamp
```

Human review is not the default path. The point of using classifiers and LLMs is to reduce human review needs while preserving appealability, auditability, and correction.

Human review is normally granted through:

```text
appeal
admin escalation
exceptional severity
suspected automated decision failure
policy exception
```

## Public text moderation

The first public text surfaces are:

```text
display_name
room_name
```

Display names and room names should share the same core moderation policy where practical:

```text
keyword and phrase screening
classifier analysis
LLM moderation
report fallback
appeal/review path
audit history for denials, holds, reversals, and admin overrides
```

Provider display names must not bypass moderation.

Future public-facing naming or profile text surfaces should route through the same policy family unless they need stricter domain-specific rules.

## Display-name moderation

Display-name moderation protects persistent public account identity.

Failed display-name moderation may:

```text
reject the proposed display name
hold the change for automated or appealable review
keep the previous accepted display name
create a moderation case
trigger warning or restriction escalation after repeated or severe violations
```

Display-name moderation should not use display names as durable identity.

Account identity remains `account_id`.

## Room-name moderation

Room-name moderation protects public session discovery, invites, room browser visibility, and joinability.

Room-name moderation is a joinability gate, not a cosmetic warning.

If a room name fails moderation after creation or update:

```text
the room becomes non-joinable immediately
the room is removed from public room browser eligibility
invite joins are blocked
matchmaking cannot route players into the room
the room is marked invalid until renamed or closed
repeated or severe violations trigger escalating warning/enforcement actions
```

Room-name enforcement attaches to the room creator by default, not the current room owner.

The room creator and room owner will usually be the same account, but enforcement must not assume ownership always proves authorship. Future ownership transfer, admin reassignment, migration, delegation, or edge-case room recovery should not blame a current owner for a room name they did not create.

Room-name violations should not punish every player in the room by default. Non-creator participants should be unaffected unless evidence shows coordinated abuse, evasion, harassment, or repeated participation in abusive rooms.

## Room repair through name correction

Invalid rooms may be repaired through name correction.

Low- and medium-severity room-name violations should generally allow the creator or authorized room controller to rename the room to an acceptable name.

Successful correction may restore:

```text
room joinability
room browser eligibility
invite joinability
matchmaking eligibility
```

Severe violations may still dissolve the room, keep it invalid, or trigger enforcement review.

Room repair corrects room state. It does not necessarily clear accumulated warning history for repeated moderation abuse.

## Enforcement actions

Possible moderation/content actions:

```text
allow
deny
hold
hide
reject_name
invalidate_room
restore_after_correction
create_case
community_warning
community_timeout
community_kick
community_ban
in_game_communication_restriction
social_feature_restriction
invite_restriction
```

Possible account actions:

```text
no_action
warning
temporary_restriction
temporary_suspension
permanent_ban
public_identity_restriction
room_creation_cooldown
public_room_creation_restriction
matchmaking_or_social_restriction
```

Possible result/reward/leaderboard actions:

```text
hold_result
reject_result
invalidate_match_result
hide_leaderboard_entry
remove_leaderboard_entry
revoke_reward_grant
revoke_progression_grant
restore_result
restore_leaderboard_entry
restore_reward_or_progression_state
```

Possible support/admin correction actions:

```text
correct_stale_entry
fix_mistaken_visibility
correct_display_or_profile_state
apply_privacy_opt_out_correction
annotate_historical_data
reverse_false_positive
```

The system should distinguish punitive/protective enforcement from non-punitive support correction.

## Escalation

Escalation should consider:

```text
severity
repeat count
recency
affected surface
intent evidence
repair behavior
appeal history
prior warnings
```

Suggested room-name escalation path:

```text
name rejected
room invalidated
creator warning
room creation cooldown
public-room creation restriction
broader matchmaking/social restriction
temporary suspension
permanent ban for severe or repeated abuse
```

Suggested general escalation path:

```text
no action
silent allow
silent deny or block
warning
temporary content/action restriction
temporary matchmaking/social/public-room restriction
temporary account suspension
permanent ban
```

Exact thresholds are open.

## Appeals

Appeals exist and should be rate-limited.

A user may appeal multiple related or unrelated enforcement issues in one appeal request.

Appeal requests should have a cooldown. The initial planning assumption is roughly one appeal request per week, with the exact duration still open.

Human review is normally granted through appeal rather than before every major automated action.

Appealable actions should include:

```text
account suspensions
account bans
account restrictions
repeated moderation warnings that create restrictions
leaderboard or result removals
reward or progression revocations
major automated enforcement actions
```

Actions that are not normally appealable as standalone issues:

```text
single rejected room_name
single rejected display_name
transient room invalidation that can be repaired by rename
```

Repeated denials that escalate into restrictions should become appealable.

Appeal review may produce:

```text
uphold
modify
reverse
restore
partial_restore
request_more_information
```

A reversal should not only clear the enforcement state. It must explicitly decide whether to restore affected results, rewards, leaderboard entries, public visibility, room/account capability, or account access.

## Audit history

Audit history is mandatory.

All major automated and admin actions must record enough information to support later appeal, review, debugging, support, and reversal.

Audit events should include:

```text
audit_event_id
actor_id
actor_type
timestamp
source_system
target_account_id
affected_object_type
affected_object_id
previous_state
new_state
category
severity
reason
evidence_reference
automated_output_reference
notes
```

Actor types should distinguish:

```text
user
admin
support
system_keyword_filter
system_classifier
system_llm_moderation
system_integrity
system_trust
```

Automated decisions must not appear to be human admin actions.

Audit history should be append-only in principle. If physical storage later allows compaction, correction, or anonymization, that belongs to legal/product retention planning and must not silently erase enforcement history needed for active appeals or support review.

## Admin and developer tooling

Initial admin tools should mostly reflect developer tooling with safer production constraints.

The intended shape is:

```text
direct state inspection
focused actions
explicit commands
clear output
case linkage
audit event creation
```

Admin tools differ from developer tools by requiring:

```text
permission checks
actor identity
reason fields
audit events
case linkage where applicable
safe confirmation for high-impact actions
restricted access to production-only enforcement actions
```

The first implementation does not need to be a large separate admin product.

Developer-tool-like internal tooling may come first, then evolve into website/admin UI workflows as the platform matures.

## Admin roles and permissions

Minimum planned role shape:

```text
viewer
support_operator
moderation_reviewer
integrity_reviewer
admin
owner_developer
```

Role meanings:

* `viewer` can inspect cases and audit history.
* `support_operator` can create support correction cases and add notes.
* `moderation_reviewer` can review display-name, room-name, report, and public-text cases.
* `integrity_reviewer` can review result, leaderboard, reward, and automation-lane cases.
* `admin` can apply restrictions, suspensions, bans, and reversals.
* `owner_developer` can perform exceptional maintenance or recovery operations.

Exact permissions are open.

High-impact actions should require stronger permissions than notes, views, or low-risk corrections.

## User-facing communication

Most user-facing enforcement communication should run through email and the website.

Email should handle formal notices such as:

```text
suspension notice
ban notice
appeal status
appeal result
major reversal
account restriction notice
```

The website should own richer account and enforcement surfaces:

```text
account status
enforcement notices
appeal request form
appeal status
restriction summary
support-facing explanations
```

The game client should show minimal blocking state only:

```text
room name rejected
room invalid
access denied
matchmaking unavailable
account restricted
```

The game client should not expose:

```text
detection internals
classifier logic
LLM details
anti-cheat details
private evidence
admin notes
sensitive enforcement reasoning
```

## Enforcement propagation

Enforcement decisions may need to affect:

```text
login/access
matchmaking
room creation
room joining
room browser visibility
invite joinability
display identity editing
leaderboard eligibility
reward eligibility
progression eligibility
website account status
email notice delivery
official Discord server access
in-client Discord communication access
invite creation/use
social/lobby participation
website community/profile surfaces
Space Rocks multiplayer access where severity warrants
```

A ban, suspension, restriction, or room invalidation must be explicit about which surfaces it affects.

Avoid vague enforcement states that downstream systems must reinterpret independently.

## Relationship to leaderboards and rankings

Leaderboards and rankings own ranking formulas, boards, categories, visibility policy, and rating lifecycle.

Abuse and enforcement admin owns abuse-related removal, hiding, invalidation, and restoration decisions after suspicious or invalid entries are identified.

The planning split is:

```text
maintenance/lifecycle
-> season resets
-> ruleset/version archival
-> global scoring bug correction
-> privacy opt-out propagation

abuse/enforcement
-> hacking
-> exploitation
-> suspicious result invalidation
-> account enforcement
-> abuse-related removal

admin/support
-> stale entries
-> mistaken visibility state
-> display/profile correction
-> manual support correction
```

Not every manual leaderboard change is abuse enforcement.

## Relationship to rooms and matchmaking

Room lifecycle owns room membership, ready state, start-game validation, game start, cleanup, and room state transitions.

Matchmaking and room discovery own browser/search/queue/assignment behavior.

Abuse and enforcement admin owns room-name moderation policy and the enforcement consequences of failed room-name moderation.

A room with a failed room name is not joinable and should not be a valid matchmaking or room-browser candidate until repaired or closed.

## Relationship to social and Discord

Social and Community Systems owns Discord-first relationships, friends, blocks, mutes, invites, presence, in-client Discord chat/DM/voice product surfaces, report-button product surfaces, and social integration planning.

Discord platform enforcement remains Discord-owned.
Official Space Rocks Discord server moderation is Space Rocks community moderation.
Discord-powered communication rendered inside the Space Rocks client or website can create Space Rocks enforcement signals.
Severe or repeated Discord community abuse can lead to Space Rocks account restrictions, social restrictions, matchmaking restrictions, suspensions, or bans.

Abuse and enforcement admin may consume Discord/community handoff signals, report summaries, or support escalations when those signals affect Space Rocks account, room, result, reward, leaderboard, or platform enforcement.

## Relationship to API product surface

This document owns policy and domain planning.

API product surface owns endpoint and product-surface planning for:

```text
case listing
case filtering
case detail
case notes
case assignment
decision application
appeal request
appeal status
audit history reads
admin action execution
```

Do not duplicate exact endpoint shapes here.

## Implementation sequence

1. Define the case and audit event model.
2. Add internal/manual case creation and notes.
3. Add automated audit records for major moderation and enforcement decisions.
4. Create trust/game-integrity handoff cases for rejected or suspicious results.
5. Add result, leaderboard, reward, and progression hide/remove/restore actions.
6. Add display-name moderation decision and audit support.
7. Add room-name moderation decision support.
8. Add room invalidation, non-joinability, room browser removal, invite block, matchmaking block, and repair-through-rename behavior.
9. Add keyword and phrase moderation rules.
10. Add classifier moderation support.
11. Add LLM moderation authority with logged output and policy/model versioning.
12. Add warning and restriction escalation.
13. Add appeal request flow with cooldown and issue bundling.
14. Add appeal review, reversal, and restoration workflow.
15. Add email and website enforcement notice surfaces.
16. Add report intake and Discord/community handoff routing, including official Discord server moderation handoff, in-client Discord communication reports, and social/invite restriction propagation.
17. Evolve internal developer-tool-like admin tools into safer website/admin workflows.
18. Add broader support correction tooling after the case/audit foundation exists.

## Open decisions

Implementation-shape decisions remain:

* Exact severity thresholds.
* Exact keyword and phrase rules.
* Exact classifier choice.
* Exact LLM moderation provider or model strategy.
* Exact LLM confidence thresholds.
* Exact actions allowed at each automated confidence/severity level.
* Exact appeal cooldown duration.
* Exact actions appealable at launch.
* Exact room-name severe-violation behavior: immediate dissolve, invalidation, or review.
* Exact room repair permissions.
* Exact warning escalation thresholds.
* Exact admin roles and permissions.
* Exact user-facing website enforcement history detail.
* Exact email templates and notice categories.
* Exact audit retention and legal handling.
* Exact report intake shape.
* Exact Discord/community handoff shape.
* Exact in-client Discord communication report context.
* Exact community moderation escalation thresholds.
* Exact server-action-to-account-enforcement mapping.

The major policy questions are decided:

* Abuse and enforcement admin is separate from trust policy and game integrity.
* Automated moderation may be authoritative.
* Human review is not the default path.
* Appeals provide the normal review path.
* Appeals are rate-limited and may bundle multiple issues.
* Room-name moderation blocks joinability.
* Failed room-name moderation invalidates the room until repair or closure.
* Room-name escalation attaches to the room creator by default.
* Low- and medium-severity invalid rooms may be repaired through name correction.
* Major automated actions must be logged for later appeal/review.
* User-facing enforcement communication primarily belongs to email and the website.
* Admin tools should initially resemble developer tooling with production safety, permission, case, and audit additions.

## Related docs

* [Trust And Eligibility Policy](../../../../domains/platform/security-and-admin/trust-and-eligibility-policy.md)
* [Game Integrity Policy](./game-integrity-policy.md)
* [Account And Identity Systems](../account-and-identity-systems.md)
* [Social And Community Systems](../social-and-community-systems.md)
* [Matchmaking And Room Discovery](../matchmaking-and-room-discovery.md)
* [Leaderboards And Rankings](../leaderboards-and-rankings.md)
* [API Product Surface](../../../protocol/api-product-surface.md)

## Notes

This document is the planning home for abuse and enforcement admin policy. Current implementation should remain documented in current-state or service/system docs when it exists.

Use `room_name` consistently for the moderated room naming surface unless the product later standardizes on `room_title`.

Room creator is the default enforcement subject for room-name violations. Room owner may be used for control and repair permissions, but ownership should not be treated as proof of authorship.

LLM moderation is allowed to be silent authority, but silent does not mean unaudited. Major automated decisions must be recorded so appeals, reversals, support review, and admin debugging can work later.

Human review should be preserved as an appeal and escalation capability, not as the default cost of every moderation decision.

Related links should stay limited to existing canonical docs and critical implemented systems. Future systems may be named without links until the docs exist and become canonical.

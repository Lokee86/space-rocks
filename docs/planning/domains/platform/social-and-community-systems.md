# Social And Community Systems

Parent index: [Platform Planning](./!INDEX.md)

## Purpose

This document plans the social and community systems domain for Space Rocks.

It defines the intended Discord-first social suite, how Space Rocks consumes Discord social features, which game-specific consequences remain Space Rocks-owned, how social facts affect matchmaking, profiles, leaderboards, website surfaces, reporting, and abuse enforcement, and how official Discord community moderation connects to Space Rocks account enforcement.

This is a platform-domain planning document. It is not an account-auth plan, not a matchmaking implementation plan, not a leaderboard formula plan, not a website visual design document, and not a replacement Discord social network plan.

## Overview

Space Rocks should use Discord as the primary provider for social and community features where practical.

The intended model is:

```text
Discord social feature
-> Space Rocks social adapter or normalized fact
-> Space Rocks game-policy decision
-> matchmaking, room, profile, leaderboard, website, report, or enforcement behavior
```

Discord should own social and communication primitives where practical:

```text
friends
friend requests
blocks
presence
Rich Presence
game invites
in-client chat
in-client DMs
in-client voice
official community server surfaces
```

Space Rocks should not build replacement systems for Discord-owned social features unless Discord cannot support the required use case.

Space Rocks still owns game-specific interpretation and enforcement:

```text
account_id
public_profile_id
room invite validation
room join eligibility
matchmaking use of social facts
recent-player match history
profile visibility
leaderboard social visibility
website/account/community surfaces
report entry points
Space Rocks enforcement consequences
```

The full plan may be implemented in stages, but this document describes the complete intended social/community shape rather than a minimal V0.

## Current status

Active planning.

Current implementation and planning already include related foundations:

```text
Discord OAuth exists
authenticated accounts are required for production multiplayer admission
account_id is the durable Space Rocks account identity
public_profile_id is planned for public profiles and website lookup
matchmaking planning reserves social and Discord relationship seams
leaderboard planning reserves friends/social filters
abuse and enforcement planning reserves Discord/community handoff routing
website planning reserves future account, profile, ranking, and community surfaces
```

Current implementation does not yet have:

```text
Discord Social SDK integration in the Godot client
in-client Discord friends/relationships
in-client Discord chat, DMs, or voice
Discord-backed game invites
Discord-backed presence/Rich Presence
recent-player social actions
Space Rocks social fact normalization
social-fact-aware matchmaking behavior
website public profile/social settings surfaces
Discord community moderation escalation into Space Rocks enforcement
```

## Ownership boundary

This document owns planning for:

```text
Discord social integration
Discord-backed friends and relationships
Discord-backed blocks and mutes
Discord-backed presence and Rich Presence
Discord-backed game invites
Discord-backed in-client chat, DMs, and voice
recent players
social report entry points
social graph privacy
profile visibility inputs
leaderboard social visibility inputs
website community/profile social requirements
official Discord community handoff expectations
social restrictions and enforcement propagation expectations
```

This document does not own:

```text
account authentication mechanics
OAuth provider implementation details
durable account identity policy
manual signup/login policy
room browser/search/queue implementation
matchmaking assignment rules
room lifecycle and room membership
game-server final join validation
leaderboard formulas
ranking board definitions
website visual design
Discord platform enforcement
full abuse case model
appeal procedure
exact HTTP endpoint shapes
exact packet schemas
native plugin implementation details
```

Account and Identity Systems owns account identity, provider identity, OAuth policy, account display identity, and account lifecycle.

Matchmaking and Room Discovery owns room browser, queueing, assignment, grouped matchmaking, room fit, and room-entry handoff.

Multiplayer Session And Lifecycle owns room membership, ready state, room lifecycle, active match session behavior, return-to-lobby, timeout, and cleanup.

Leaderboards And Rankings owns board definitions, ranking formulas, ranking surfaces, public ranking identity, and leaderboard lifecycle.

Abuse And Enforcement Admin owns moderation cases, automated enforcement, warnings, restrictions, suspensions, bans, appeals, audit history, and Space Rocks-side consequences.

Website And Web Presence owns website product presentation, page structure, account portal presentation, public profile pages, and public web experience.

## Discord-first social model

Discord is the preferred provider for:

```text
friends
friend requests
relationship state
blocks
presence
Rich Presence
game invites
in-client chat
in-client DMs
in-client voice
community server entry points
community roles or membership facts where useful
```

Space Rocks should consume Discord facts through a narrow adapter rather than scattering raw Discord API behavior through game systems.

Useful normalized facts may include:

```text
discord_user_id linked to account_id
is_discord_friend
is_blocked
is_muted
presence_state
can_invite
can_receive_invite
can_communicate
can_join_social_context
```

Exact field names are implementation details.

The invariant is that downstream systems consume Space Rocks-normalized facts, not raw Discord SDK objects.

## Discord plugin and bridge policy

The existing Godot Discord Social SDK plugin is the preferred first implementation path.

If that plugin is not viable, Space Rocks may fork it, patch it, use another bridge, or build a native bridge around the official Discord SDK.

The social plan does not depend on the plugin being guaranteed viable. If Discord integration becomes impossible or impractical at implementation time, the plan can pivot then.

Plugin viability is not a product-planning blocker.

## Account and profile identity

Discord identity does not replace Space Rocks account identity.

The identity split is:

```text
account_id
-> internal durable Space Rocks account identity

public_profile_id
-> public Space Rocks profile and website lookup identity

display_name
-> moderated Space Rocks presentation identity

discord_user_id
-> linked provider/social identity
```

Discord profiles probably should not be used as Space Rocks public profiles.

Space Rocks public profiles should remain Space Rocks-owned and should use `public_profile_id`.

Discord identity may appear as linked-provider or community context only where product settings and privacy rules allow it. It should not become the public Space Rocks profile model.

## Friends and relationships

Discord should provide friend and relationship facts where practical.

Space Rocks consumes those facts for:

```text
friend-facing room discovery where supported
friend-filtered leaderboard views
friend/profile visibility rules
friend invite eligibility
presence visibility where allowed
website signed-in friend filters where supported
```

Friend status should not bypass:

```text
account restrictions
room restrictions
block policy
private visibility settings
matchmaking validation
game-server final join validation
```

Friend-specific surfaces should fail safely when relationship facts are unavailable.

## Blocks and mutes

Blocks and mutes are Discord-backed where practical, but Space Rocks still decides their game consequences.

Blocks should always block communication where technically possible.

Blocked users should not be able to use Space Rocks-rendered Discord surfaces for direct communication, including:

```text
chat
DMs
voice
direct social invites
```

Blocks should affect automatic matchmaking co-placement where practical, but Space Rocks does not need to guarantee impossible global avoidance in every edge case.

Blocks may also suppress richer social/profile/presence views where supported.

Mute is communication-focused.

Mute should suppress communication or presentation surfaces, but mute does not by itself imply room, match, or queue exclusion.

Safe failure behavior matters. Matchmaking, room joining, and invite failure messages should not expose private social relationship details.

## Presence and Rich Presence

Discord presence and Rich Presence are part of the planned social suite.

Exact Rich Presence fields are a gametime decision.

Presence may eventually show safe public activity such as:

```text
main menu
in lobby
in match
mode category
party size
joinability where safe
season/event/campaign context where public
```

Presence and Rich Presence must not expose:

```text
private room names
internal room IDs
account IDs
session IDs
invite tokens
matchmaking ratings
enforcement state
hidden or private player state
non-public social relationship details
debug-only state
```

Presence should respect Space Rocks privacy rules where Space Rocks controls or filters the displayed surface.

## Invites

Discord should carry or present social invites where practical.

Space Rocks validates the game consequence.

The planned invite flow is:

```text
Discord invite
-> carries or resolves to opaque Space Rocks join token
-> API server validates account, invite, social restrictions, room state, and eligibility
-> game server performs final room join validation
```

Invite joins must not bypass:

```text
authentication
account bans
social restrictions
room invalidation
room capacity
private-room policy
block policy
matchmaking eligibility
game-server final validation
```

Room invites are social-owned as a product surface.

Matchmaking owns whether an invite target can become a valid room assignment or grouped entry.

The game server owns final room/session admission.

## Grouped play, parties, and matchmaking handoff

Grouped play is matchmaking-owned.

Discord may provide social context, communication, invites, or lobby-like UI support, but Space Rocks matchmaking owns:

```text
grouped matchmaking requests
party or group fit
queue participation
room assignment
room-entry handoff
eligibility checks
```

The game server owns final room join validation.

Space Rocks should not build a general custom social party system if Discord can provide the social layer and matchmaking only needs a temporary grouped request.

If Discord lobby features are useful for communication or social context, they may be integrated without making Discord authoritative for Space Rocks matchmaking.

## In-client chat, DMs, and voice

Discord-powered in-client chat, DMs, and voice are part of the full planned social suite.

These are not custom Space Rocks communication systems. They are Discord-backed communication surfaces rendered inside the Space Rocks client where practical.

Because they are rendered inside Space Rocks, they are also Space Rocks product surfaces.

They require:

```text
report entry points
mute/block behavior
Space Rocks moderation routing
Discord/community moderation handoff
Space Rocks enforcement handoff
safe error and restriction states
```

Discord platform enforcement remains Discord-owned.

Space Rocks owns its own community standards and Space Rocks-side consequences when Discord-powered communication affects Space Rocks accounts, access, matchmaking, social features, rooms, profiles, leaderboards, or website surfaces.

## Recent players

Recent players are Space Rocks-owned because Discord does not know Space Rocks match history.

Recent-player records should be built from Space Rocks room or match participation.

Recent players should support:

```text
safe profile open
invite where allowed
block through Discord integration where available
mute where applicable
report
match context for report handoff
```

Recent-player surfaces should expose only requester-safe public identity and relevant match context.

Recent-player records should not expose:

```text
account_id
session IDs
internal room IDs
private matchmaking state
hidden enforcement state
raw Discord identifiers unless intentionally exposed
```

Recent-player history is a game-local social utility, not a replacement friend system.

## Reports and moderation handoff

Social and community systems own report-button product surfaces for social and community contexts.

Report entry points should exist on relevant Space Rocks-rendered surfaces, including:

```text
recent players
player profile
room member list
post-match result screen
leaderboard entry
room browser or room name surface
website public profile
website leaderboard entry
Discord-powered in-client chat
Discord-powered in-client DM
Discord-powered in-client voice context
Discord-powered invite or lobby context
```

Abuse And Enforcement Admin owns case creation, automated moderation, enforcement decisions, appeals, and audit history.

Social/community surfaces provide the report path and context.

Reports should include enough context for enforcement without exposing sensitive internals to users.

## Official Discord community moderation

The official Space Rocks Discord server is part of the Space Rocks community surface.

Discord platform enforcement remains Discord-owned, but Space Rocks is responsible for moderation of official Space Rocks Discord community spaces.

Official Discord server moderation actions may become Space Rocks enforcement signals.

Depending on severity, context, and evidence, Discord community moderation may lead to:

```text
Space Rocks account warning
social feature restriction
in-game communication restriction
invite restriction
matchmaking/social restriction
temporary account suspension
account ban
```

This should be reflected in Abuse And Enforcement Admin.

The abuse/enforcement plan should consume signals such as:

```text
official Discord server moderation actions
Discord-powered in-client chat reports
Discord-powered in-client DM reports
Discord-powered in-client voice reports
Discord invite abuse
Discord lobby or social abuse
community moderator escalations
Discord community role or membership abuse signals
```

The system should remain automation-first where practical. Human review is primarily for appeals, exceptional escalation, suspected automated decision failure, or admin intervention.

## Website and community surfaces

The website supports the Space Rocks side of community. It does not replace Discord community discussion or mirror Discord profiles.

Planned website/community surfaces include:

```text
public Space Rocks profiles
public leaderboard pages
friend-filtered leaderboard views when signed in
season, event, and campaign pages
shareable match-result pages
invite landing pages
account and social settings portal
enforcement, appeal, and support portal
official Discord community hub
community rules and code of conduct
report entry points from public surfaces
```

The website should not add:

```text
custom forums
custom comments
custom DMs
custom social feed
custom guild or clan pages
Discord profile mirrors
public raw Discord identity pages
```

Website public profiles should use Space Rocks `public_profile_id`, not Discord profile identity.

Discord identity may be displayed only as optional linked-provider or community context where allowed.

## Leaderboard and profile visibility

Leaderboards consume social relationship and visibility facts.

Social/community owns relationship facts and social visibility inputs.

Leaderboards own ranking formulas, board definitions, board lifecycle, ranking views, and ranking catalog source of truth.

Social inputs may support:

```text
friends leaderboard filter
profile visibility
blocked/private visibility handling
public board opt-outs
website ranking/profile surfaces
report entry points
```

Local profile rankings do not use Discord social filters.

Online authenticated-account rankings may use social filters where the required durable facts and privacy rules exist.

## Matchmaking and room discovery handoff

Matchmaking consumes normalized social facts.

It should not depend directly on raw Discord API behavior.

Possible social facts consumed by matchmaking include:

```text
friend relationship exists
blocked relationship exists
muted relationship exists
invite eligibility exists
grouped play request exists
social restriction exists
```

Social owns relationship, invite, and communication facts.

Matchmaking owns browser/search/queue/assignment behavior.

Room lifecycle owns room membership and room state.

The game server owns final room join validation.

Social relationship failures should produce safe failure states and should not expose private relationship or moderation details.

## Public text and profile safety

Space Rocks display names, room names, public profile fields, website-visible profile content, and other public text surfaces must use Space Rocks moderation rules.

Discord provider display names must not bypass Space Rocks moderation.

Discord profile identity should not be treated as automatically safe Space Rocks public identity.

Where Discord-powered communication is rendered inside Space Rocks, moderation and report routing must be planned as a Space Rocks product responsibility, even though Discord platform enforcement also applies.

## Failure and degraded behavior

Core online gameplay should not depend on live Discord social features once account authentication and admission are satisfied.

Convenience features may fail soft:

```text
Rich Presence disabled
friend filters hidden
social panel unavailable
Discord invite creation unavailable
in-client chat or voice unavailable
```

Safety and privacy features should fail closed where missing social data could expose users, bypass restrictions, or route blocked users into direct communication.

Examples:

```text
direct communication with blocked users should not be allowed when block state cannot be safely checked
private social/profile views should not be exposed when visibility cannot be verified
invite acceptance should not bypass Space Rocks validation when Discord state is unavailable
```

Exact degraded-mode behavior belongs to implementation and product design, but the safety invariant should remain stable.

## Implementation sequence

1. Graduate this document from stub to canonical platform planning.
2. Update the platform index to point to the graduated document path.
3. Update Abuse And Enforcement Admin to reflect official Discord community moderation responsibility and in-client Discord communication moderation responsibility.
4. Define the Discord social adapter boundary so downstream systems consume normalized facts.
5. Evaluate the existing Godot Discord Social SDK plugin as the preferred first implementation path.
6. Add or bridge Discord social SDK support through the selected implementation path.
7. Add Discord presence and Rich Presence support with safe-field restrictions.
8. Add Discord relationship read support for friends, blocks, and mutes where practical.
9. Add Space Rocks interpretation rules for communication blocking, invite eligibility, and safe visibility.
10. Add Discord-backed game invite handoff using opaque Space Rocks invite or join tokens.
11. Add recent-player history and social actions.
12. Add report entry points for social, recent-player, profile, match-result, and Discord-powered communication surfaces.
13. Add in-client Discord chat, DMs, and voice with report, mute/block, and enforcement handoff behavior.
14. Add website public profile, social settings, leaderboard/social filters, invite landing, report, appeal, and community hub surfaces.
15. Add official Discord server moderation handoff signals into Space Rocks abuse/enforcement.
16. Expand matchmaking and leaderboard consumers to use normalized social facts where relevant.
17. Keep Discord profile mirroring, custom forums, custom comments, custom social feeds, and custom social-network features out of scope unless a later plan explicitly changes direction.

## Open decisions

Exact implementation details remain open, including:

```text
exact Discord plugin, fork, or bridge path
exact normalized social fact names
exact Rich Presence field set
exact in-client chat/voice/DM UI
exact report context captured from Discord-powered communication
exact recent-player retention limits
exact website page routes
exact social settings names
exact degraded-mode UX
exact automation thresholds for Discord/community enforcement signals
```

These are not open product-direction questions:

```text
whether social is Discord-first
whether Space Rocks should avoid replacement Discord social systems
whether in-client Discord communication belongs in the full plan
whether Discord server moderation can affect Space Rocks account enforcement
whether Space Rocks owns game-specific interpretation and enforcement consequences
whether Discord profiles replace Space Rocks public profiles
```

## Related docs

* [Platform Planning](./!INDEX.md)
* [Account And Identity Systems](account-and-identity-systems.md)
* [Matchmaking And Room Discovery](matchmaking-and-room-discovery.md)
* [Multiplayer Session And Lifecycle](multiplayer-session-and-lifecycle.md)
* [Leaderboards And Rankings](leaderboards-and-rankings.md)
* [Abuse And Enforcement Admin](security-and-admin/abuse-and-enforcement-admin.md)
* [API Product Surface](../../protocol/api-product-surface.md)
* [Website And Web Presence](../web/website-and-web-presence.md)
* [Modes And Match Rules](../gameplay/modes-and-match-rules.md)
* [Player Experience Systems](../gameplay/player-experience-systems.md)

## Notes

The current planning direction is stable: Discord-first social, no custom Discord replacement systems, Space Rocks-owned game consequences, in-client Discord communication in the full target plan, and official Discord community moderation connected to Space Rocks enforcement where appropriate.

The full social suite may still be implemented in stages. Staged implementation should not be documented as if the excluded stages are no longer part of the target plan.

The phrase “Discord moderation” should be used carefully. Discord platform enforcement remains Discord-owned, but Space Rocks is responsible for official Space Rocks Discord community moderation and for Space Rocks-side consequences from Discord-powered social and communication surfaces.

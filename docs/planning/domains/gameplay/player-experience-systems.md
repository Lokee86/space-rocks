# Player Experience Systems
Parent index: [Gameplay Planning](./!INDEX.md)

## Purpose

This doc maps the player-facing experience loops for Space Rocks.

It coordinates how the player moves between identity/profile selection, pregame preparation, room setup, lobby, loadout selection, match runtime, post-match results, progression, rewards, and review surfaces.

This doc is an umbrella map. It does not replace the narrower system plans that own mode rules, inventory, loadouts, rewards, achievements, commerce, or persistence.

## Ownership Boundary

This doc owns the broad player-facing flow and the handoffs between related systems.

It covers:

```text
profile / identity loop
pregame preparation loop
inventory / hangar / shop access points
room setup and session entry flow
single-player create flow
multiplayer create flow
multiplayer join flow
lobby presentation and pre-start coordination expectations
loadout display and selection points
match runtime handoff
post-match result, progression, and reward presentation
next available player actions
cross-system handoff vocabulary
```

This doc does not own:

```text
exact UI layout
scene hierarchy
mode behavior
objective behavior
content catalog schemas
inventory schema
shop pricing
reward formulas
achievement definitions
database schema
packet or API field names
```

Detailed ownership belongs to the related system docs.

## Core PX Model

Player experience is not one linear lifecycle. It is a set of connected loops around the match flow.

```text
Profile / Identity Loop
-> Preparation / Hangar / Shop Loop
-> Room / Session Loop
-> Build / Readiness Loop
-> Match Runtime Loop
-> Post-Match / Rewards Loop
-> Review / Progression Loop
```

The main player-facing flow is:

```text
choose who I am
-> prepare what I can use
-> choose or join what I am playing
-> select a valid build/loadout
-> ready/start
-> play the match
-> see results
-> receive progression/rewards
-> choose the next action
```

## Handoff Vocabulary

Use existing system vocabulary where possible.

Planned handoff concepts:

```text
SelectedPlayerContext
PlayerDataSnapshot
RoomContentConfig
ResolvedMatchRules
HangarInventory
BuildEligibility
LoadoutSelection
ResolvedPlayerBuild
EndOfMatchFlow
MatchSummary
MatchSummaryDispatcher
GrantAward
```

`SelectedPlayerContext` is the selected identity/profile context used by the player-facing flow.

`PlayerDataSnapshot` is the loaded or safely synthesized player-facing state needed by pregame, hangar, shop, loadout, progression, and review surfaces.

`RoomContentConfig` is the create/configure input for a new room or single-player run.

`ResolvedMatchRules` is the authoritative playable resolution consumed after room setup is validated and resolved.

## Profile And Identity Loop

Single-player identity is selected from the pregame menu.

```text
pregame_menu
-> Guest or Local Profile selection
-> selected single-player identity
-> profile readout / local pilot flow
```

Multiplayer identity is selected through login/auth.

```text
login/auth flow
-> Authenticated Account identity
-> multiplayer pregame / create / join
```

Rules:

```text
Profile selection only happens in single-player pregame_menu.
Multiplayer identity comes from login/auth.
Identity cannot change inside any room.
Local Profile state never imports into multiplayer.
Guest behaves like a normal profile-shaped identity for gameplay flow, but uses transient player-data storage.
```

## Preparation, Inventory, Hangar, And Shop Loop

The pregame menu is the hub for preparation surfaces.

Pregame-accessible surfaces:

```text
profile readout
local pilot selection for single-player
RoomContentConfig setup
inventory / hangar
loadout creation and editing
shop / provisioner
wallet display
progression and review surfaces as they are added
```

Inventory, hangar, loadout creation/editing, shop, and provisioner interaction are not active-match surfaces.

Loadouts are not changeable mid-match for now. Respawn loadout changes may be considered later, but the baseline rule is:

```text
active match
-> no loadout changes
-> respawn restores from existing ResolvedPlayerBuild
```

## Room And Session Loop

Room setup starts from `RoomContentConfig`.

```text
RoomContentConfig
-> server validation / resolution
-> ResolvedMatchRules
```

`RoomContentConfig` is the configured setup for a new room or single-player run. It may include mode refs, mode options, objective refs, mission refs, challenge refs, match-level refs, encounter refs, spawn refs, arena refs, and content modifiers as owned by the narrower docs.

`RoomContentConfig` is not gameplay policy. It composes refs and selected options. The owning systems validate and resolve those refs.

`ResolvedMatchRules` is the authoritative playable rule object used by match runtime, lobby presentation, build filtering, results, progression eligibility, and related systems.

## Single-Player Create Flow

Single-player uses the same room-configuration concept as multiplayer create, but does not enter a lobby.

```text
SelectedPlayerContext
-> PlayerDataSnapshot
-> player configures RoomContentConfig in pregame_menu
-> server validates / resolves RoomContentConfig
-> ResolvedMatchRules
-> loadout selection / validation
-> ResolvedPlayerBuild
-> match starts directly
```

Single-player replay immediately starts a new game with the same match config.

Normal single-player post-match flow remains the current flow, with progression and reward presentation added.

## Multiplayer Create Flow

Multiplayer create starts from the pregame menu and creates a resolved room.

```text
SelectedPlayerContext
-> PlayerDataSnapshot
-> creator configures RoomContentConfig
-> server validates / resolves RoomContentConfig
-> room stores resolved rules / resolved room summary
-> creator enters lobby
-> loadout display / existing-loadout selection
-> ready / owner start
```

The creator configures room setup before room creation.

After room creation, lobby surfaces should present resolved setup, not raw configuration internals.

## Multiplayer Join Flow

Multiplayer join enters an existing resolved room.

```text
SelectedPlayerContext
-> PlayerDataSnapshot
-> player joins existing room
-> player receives ResolvedMatchRules or a resolved room summary
-> loadout display / existing-loadout selection
-> ready
```

Joiners do not interact with `RoomContentConfig`.

By the time a player is joining, the room has already been created and resolved.

## Lobby And Pre-Start Coordination

The lobby already owns create/join/leave/ready/start flow.

Current lobby behavior:

```text
create room
join room
leave room
owner-gated start
ready state
return-to-lobby path
```

Planned lobby upgrades are presentation and pre-start coordination upgrades, not a replacement lobby system.

Planned lobby presentation:

```text
resolved match / mode rules summary
objective summary
room content summary
player slots
display names beside player / team slots
owner marker
local player marker
ready status per player
empty slot display
loadout display
blocked / invalid loadout indicator
start availability feedback
countdown state
```

Planned lobby interactions:

```text
select from existing valid loadouts
select team when the resolved mode/team policy uses teams
select player color when player_color_policy allows it
ready / unready
owner start
leave room
```

Loadout creation and editing remain pregame/hangar actions. The lobby may select among existing valid loadouts, but does not become the full hangar editor.

Teams are real mode/match functionality. Team selection belongs to the resolved mode/team policy and lobby presentation.

Player color selection uses existing color vocabulary:

```text
player_color_policy
Player.player_hue
```

Known policy values include:

```text
local_selected
auto_distinct
player_id_assigned
```

Color selection should only be available when the resolved `player_color_policy` allows player choice.

## Countdown

Countdown starts when the owner clicks Start.

```text
owner clicks Start
-> countdown begins
-> any player unreadying can cancel countdown until the final 1 second
-> final 1 second locks the start
-> match starts
```

Countdown state should be visible in the lobby.

During the locked final second, the lobby should not allow changes that would invalidate the start.

## Build, Loadout, And Readiness Loop

`BuildEligibility` is a small filtering seam, not a major player-experience system.

It filters available player builds against the resolved room rules.

```text
HangarInventory / saved loadouts
+ ResolvedMatchRules
-> BuildEligibility
-> allowed / blocked loadout options
-> LoadoutSelection
-> ResolvedPlayerBuild
```

It answers:

```text
Can this build/loadout be used here?
If not, why?
What valid alternatives remain?
```

Build filtering may consider resolved mode rules, team rules, ship restrictions, weapon restrictions, module restrictions, hardwired equipment policy, and other rule restrictions defined by the owning systems.

`LoadoutSelection` is the player choice.

`ResolvedPlayerBuild` is the immutable match-start build setup consumed by runtime.

## Match Runtime Loop

Match runtime consumes resolved setup.

```text
ResolvedMatchRules
+ ResolvedPlayerBuild per player
+ resolved content / encounter / spawn runtime refs
-> runtime match
```

Runtime state should not mutate the resolved setup objects.

Runtime-owned state includes:

```text
current health
current shield
current ammo
cooldowns
pickup overwrites
temporary softpoint weapons
active buffs / debuffs
death and respawn state
```

Resolved setup remains the source for match start and respawn baseline behavior.

Mid-match loadout changes are not supported for now.

## Match End And Results Loop

Match end is coordinated by the match outcome system.

```text
match-end condition reached
-> EndOfMatchFlow
-> MatchSummary
-> MatchSummaryDispatcher
-> persistence / progression / achievement / presentation slices
```

The player-facing result surface should use presentation-safe result data.

Current match result behavior remains valid and should be upgraded rather than replaced.

Planned post-match presentation:

```text
Match Results
-> XP / standard progression dialog
-> normal return flow
```

Single-player normal continue returns to `pregame_menu` after progression/reward presentation.

Replay starts a new match immediately with the same match config.

Return to lobby keeps the existing implemented behavior.

## Progression, Rewards, And Notifications

Standard progression uses a post-match dialog.

```text
XP / rank / standard progression
-> post-match progression dialog
-> return flow
```

Item rewards should create visible inventory or hangar feedback.

```text
new item reward
-> inventory / hangar flash or highlight
-> new item highlighted where it appears
```

Currency rewards update the wallet balance. They do not require a blocking reward flow.

```text
currency increase
-> wallet balance updates
-> optional small animation on or above the currency display
```

Achievements use persistent notification presentation.

```text
achievement completion
-> persistent CanvasLayer notification
-> can appear whenever completion is received
```

Achievement notifications are not limited to post-match. They may appear during active gameplay, menus, or future loading screens if an achievement completion is received there.

## Review And Next-Action Surfaces

PX review surfaces include current and future player-facing summaries.

```text
profile readout
local pilot selector
match results
XP / progression dialog
achievement / milestone notifications
inventory / hangar
loadout creation and editing
shop / provisioner
wallet display
leaderboards later
social / community later
match history later
```

Post-match and pregame next actions include:

```text
Replay same match config
Return to lobby
Return to pregame_menu
Change RoomContentConfig from pregame_menu
Change / create / edit loadouts from pregame_menu
Open inventory / hangar from pregame_menu
Open shop / provisioner from pregame_menu
View profile / progression / results
Return to main menu
```

## Cross-System Handoff Map

High-level handoff:

```text
Profile / Login
-> SelectedPlayerContext
-> PlayerDataSnapshot
-> Pregame Preparation
-> RoomContentConfig
-> ResolvedMatchRules
-> BuildEligibility
-> LoadoutSelection
-> ResolvedPlayerBuild
-> Runtime Match
-> EndOfMatchFlow
-> MatchSummary
-> MatchSummaryDispatcher
-> player-data / progression / achievements / presentation
-> Next Action
```

Single-player create:

```text
pregame_menu
-> RoomContentConfig
-> ResolvedMatchRules
-> ResolvedPlayerBuild
-> match start
```

Multiplayer create:

```text
pregame_menu
-> RoomContentConfig
-> ResolvedMatchRules / resolved room summary
-> lobby
-> ready / countdown / start
```

Multiplayer join:

```text
join existing room
-> ResolvedMatchRules / resolved room summary
-> lobby
-> ready / countdown / start
```

Post-match:

```text
EndOfMatchFlow
-> MatchSummary
-> presentation result
-> progression / rewards / achievements
-> replay / lobby / pregame / menu
```

## Implementation Planning

Recommended implementation direction:

```text
1. Promote this doc out of stubs after completion.
2. Keep current menu, local pilot, lobby, match result, replay, and return-to-lobby behavior intact.
3. Map current single-player play through the RoomContentConfig path.
4. Map multiplayer create through the RoomContentConfig path.
5. Make join flow consume resolved room summary / ResolvedMatchRules-derived data.
6. Add resolved match/mode/objective summary presentation to lobby.
7. Add display names, owner/local markers, ready state, and player/team slot presentation to lobby.
8. Add lobby loadout display.
9. Add existing-loadout selection in lobby.
10. Add blocked/invalid loadout display where resolved rules disqualify a selected loadout.
11. Add player color selection only when player_color_policy allows it.
12. Add team selection for modes that use teams.
13. Add start countdown with unready-cancel behavior until the final 1 second.
14. Add post-match XP/progression dialog.
15. Add item reward highlight behavior in inventory/hangar surfaces.
16. Add optional currency-display animation for currency increases.
17. Add achievement CanvasLayer notification presentation.
```

Early slices should preserve current behavior while routing new presentation through the planned seams.

## Testing Direction

Important future checks:

```text
single-player configures RoomContentConfig and starts directly
multiplayer create configures RoomContentConfig and enters lobby
multiplayer join receives resolved rules or resolved room summary
joiners do not need RoomContentConfig
lobby displays resolved match/mode/objective summary
lobby displays player display names beside slots
lobby displays owner and local player markers
lobby displays ready state
lobby displays selected loadout
lobby can select existing valid loadouts
lobby shows blocked loadout reasons when relevant
team selection appears only for modes that use teams
player color selection appears only when player_color_policy allows it
owner start begins countdown
unready cancels countdown until the final 1 second
final 1 second locks start
loadouts cannot change mid-match
respawn restores from ResolvedPlayerBuild
Match Results still display
single-player Replay starts same match config immediately
return to lobby keeps existing behavior
XP/progression dialog appears after match results
item rewards highlight inventory/hangar surfaces
currency increases can animate near currency display
achievement notification can appear whenever achievement completion is received
```

## Related Docs

* [Planning](../../!INDEX.md)
* [Modes And Match Rules](modes-and-match-rules.md)
* [Levels, Missions, And Content Structure](levels-missions-and-content-structure.md)
* [Player Build And Loadouts](player-build-and-loadouts.md)
* [Inventory And Hangar](inventory-and-hangar.md)
* [Match Outcomes And Results](match-outcomes-and-results.md)
* [Progression And Rewards](progression-and-rewards.md)
* [Shop, Commerce, And Economy](shop-commerce-and-economy.md)
* [Account And Identity Systems](../platform/account-and-identity-systems.md)
* [Player Data And Persistence](../../services/player-data/!INDEX.md)
* [Multiplayer Session And Lifecycle](../platform/multiplayer-session-and-lifecycle.md)
* [Matchmaking And Room Discovery](../platform/matchmaking-and-room-discovery.md)

## Open Gametime Decisions

```text
Exact SelectedPlayerContext fields.
Exact PlayerDataSnapshot shape.
Exact resolved room summary shape.
Exact lobby layout.
Exact loadout selector presentation.
Exact blocked loadout reason vocabulary.
Exact team policy fields.
Exact player_color_policy behavior for selectable room colors.
Exact countdown duration.
Exact final-second lock presentation.
Exact XP/progression dialog layout.
Exact inventory/hangar reward highlight behavior.
Exact currency animation presentation.
Exact achievement notification layout and queueing behavior.
Whether future respawn loadout changes are worth adding.
```

## Core Invariants

```text
Player Experience Systems maps player-facing loops and handoffs; it does not own detailed mechanics.

Single-player and multiplayer create both configure RoomContentConfig.

Single-player starts directly after setup.

Multiplayer create enters lobby after room setup.

Multiplayer join consumes ResolvedMatchRules or resolved room summary.

Joiners do not interact with RoomContentConfig.

RoomContentConfig is a create/configure input.

ResolvedMatchRules is the authoritative playable resolution.

Inventory, hangar, loadout creation/editing, shop, and provisioner live in pregame_menu.

Lobby may select from existing valid loadouts but does not become the full hangar/editor.

Loadouts are not changeable mid-match for now.

Teams are real mode/match functionality.

Player color selection depends on player_color_policy.

Countdown starts when owner clicks Start.

Countdown can be canceled by anyone unreadying until the final 1 second.

The final 1 second locks the start.

XP/progression uses a post-match dialog.

Item rewards highlight inventory/hangar surfaces.

Currency increases may animate near the currency display.

Achievements use persistent CanvasLayer notifications and are not limited to post-match.

Replay starts a new game with the same match config.

Return to lobby keeps existing implemented behavior.
```

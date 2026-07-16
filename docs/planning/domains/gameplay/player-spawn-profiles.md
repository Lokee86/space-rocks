# Player Spawn Profiles
Parent index: [Gameplay Planning](./!INDEX.md)

## Purpose

This doc is the authoritative P4 planning owner for player-spawn profile selection, placement, safety, and spawn presentation handoffs.

It defines how a selected profile chooses a preferred location for a player-spawn reason, applies reusable safety search, incorporates team and world context, and returns an authoritative spawn outcome. It defines planning boundaries without claiming that the implementation already exists.

## Ownership Boundary

This doc owns planning for:

```text
player-spawn profile identifiers and selection
spawn-reason-specific placement policy
initial-spawn and respawn location selection
mid-match join and recovery/revival placement
checkpoint/campaign and admin/devtool placement policy
preferred-location sources and ordering
reusable spawn safety search
world-wrap-aware spawn regions and safety checks
team-aware player placement
simultaneous position reservation
spawn fallback and retry behavior
spawn stationary-state handoff
spawn protection profile seam
spawn-placement telemetry requirements
on-screen spawn-indicator policy
```

This doc does not own:

```text
respawn eligibility, lives, death, or elimination policy
team assignment or team membership authority
encounter, enemy, asteroid, or other non-player spawning
room membership, admission, match lifecycle, or reconnect execution
mode objectives or match-end policy
combat damage formulas or collision detection
ship loadouts or equipment restoration
client UI layout or input ownership
packet or storage schemas
```

Modes and match rules select the player-spawn profile as part of resolved rules. The player-spawn profile applies the selected placement and safety policy. Owning systems provide the participant, spawn reason, team facts, world state, and eligibility context rather than redefining player placement or safety.

## Settled Product Model

The initial profile is:

```text
basic_safe_spawn_v1
```

The first slice preserves the existing origin-area grid baseline and the current outward safety search. This profile planning does not replace that behavior with a new placement algorithm. It separates preferred-location selection from the reusable safety search so later profiles can vary location policy without duplicating safety behavior.

A profile may define different preferred-location selection for initial spawn and respawn. Respawn is not permanently tied to the player's original initial-spawn anchor. The profile receives the current spawn context and may select a new preferred location using the reason-specific policy and current world state.

Supported spawn reasons are:

```text
initial spawn
respawn
mid-match join
recovery/revival
checkpoint/campaign transition
admin/devtool spawn
```

A spawn reason is authoritative context for profile policy. It does not establish eligibility: lives, death, elimination, lifecycle, campaign, or devtool owners decide whether the request is permitted.

## Preferred-Location Selection

Location sources may include:

```text
fixed points
generated formations
near teammates
away from enemies
near objectives
checkpoint locations
random valid areas
```

The profile may combine or order these sources by spawn reason, mode, team context, and available world facts. The first slice does not include player-selected spawn locations. Player-selected placement may be added later only as an explicit mode/profile capability with authoritative validation.

The existing origin-area grid remains the baseline for the initial profile. The current outward search remains the fallback search behavior after a preferred location has been selected. A preferred location is therefore an input to safety evaluation, not a guarantee that the player will be placed there.

Authored or scripted campaign points are supported. Normal safety search applies to those points unless an explicit scripted override bypasses safety. Such an override must be deliberate, visible in the resolved spawn context, and owned by the campaign/scripted content seam rather than inferred from the point's origin.

## Safety Search And Placement

Safety avoids generic dangerous objects. Exact dangerous-object categories may expand as gameplay adds hazards, enemies, projectiles, or other threats. The safety contract should remain category-based rather than assuming one fixed object list.

Other-player overlap is a separate presentation/placement concern. A profile may allow it even when safety rejects dangerous-object overlap. Profiles may later choose stricter separation or a player-overlap resolution policy without changing generic danger evaluation.

Safety search and spawn regions respect toroidal world wrapping. Distances, neighborhood checks, search expansion, and region boundaries must use wrap-aware spatial rules rather than treating the world edge as a hard boundary.

The profile reserves simultaneous positions before assignment is finalized. Assignment order has no gameplay significance. Reservation prevents simultaneous placements from accidentally selecting the same location when the selected profile requires distinct positions, while keeping the order of requests from changing gameplay outcomes.

If no safe point is found, the profile restarts from the preferred origin/current anchor and retries until successful while re-evaluating world state. It must not return an unsafe point merely because the first search was exhausted, and it must not assume that a previously rejected point remains rejected after the world changes.

## Team-Aware Placement

Team placement is mode-selected through the player-spawn profile. The default expectation is to group teammates and separate opponents. The profile consumes authoritative team membership and relationship facts from the team system; it does not assign teams or redefine team relationships.

Team-aware preferences affect preferred-location selection and placement constraints. They do not bypass generic danger safety. A selected mode may choose a profile that ignores team grouping, but that choice must be explicit in resolved rules rather than inferred from team structure.

## Spawn State And Presentation

Ships spawn stationary and control starts immediately. The spawn outcome must establish the stationary initial runtime state, after which authoritative gameplay accepts normal control without a product-level warmup or forced launch delay.

There is no spawn protection in the baseline. Safety placement is sufficient for the initial balance model. Spawn protection remains an optional profile-owned future seam because it is high balance risk. If enabled by a future profile, that profile owns its duration, protected interactions, cancellation conditions, and any movement or attack restrictions; lives and respawn rules must not embed those mechanics.

Existing on-screen spawn indicators remain visible to everyone for now. Indicator visibility, timing, and presentation may evolve later, but the first profile contract does not make indicators private to the spawning player or team.

## Profile Selection And Configuration

The mode selects the player-spawn profile. Raw spawn tuning is not exposed at game creation initially. Future modes may expose selected, validated spawn options, but clients must not submit arbitrary profile internals or replace the mode's resolved profile.

The resolved rules carry the selected profile ID and any explicitly resolved options needed by the runtime. The profile remains the owner of interpreting those options; consumers must not duplicate profile-specific placement policy.

## Spawn-Reason Handoffs

Initial spawn uses the profile's initial-spawn preferred-location policy and the preserved origin-area grid baseline, followed by reusable safety search.

Respawn uses respawn location logic first. It may select a location different from the player's original initial-spawn anchor. This placement request is made only after the lives/death/respawn owner has established respawn eligibility and trigger policy.

Mid-match join uses respawn location logic first, then the baseline initial-spawn fallback. Lifecycle and mode rules decide whether the join is admitted and activated; the profile decides placement after receiving the resulting spawn reason.

Recovery/revival uses the profile policy for that reason and current world state. Recovery does not silently become a new participant or bypass the owning lives/lifecycle transition.

Checkpoint/campaign transition may provide an authored or scripted point. Normal safety search applies unless an explicit scripted override bypasses it.

Admin/devtool spawn is an explicit administrative request. Devtools may select the reason and request context, but the profile remains the placement and safety authority unless an explicit scripted/admin override is part of the resolved contract.

## Telemetry Baseline

The baseline telemetry record includes:

```text
profile ID
spawn reason
preferred position
final position
whether safety moved the player
retry count
```

Telemetry records the authoritative result of a placement attempt. It should distinguish a preferred location from the final safe location and make retry behavior observable without requiring clients or devtools to reconstruct the search.

## System Handoffs

```text
resolved mode rules
-> player_spawn_profile_id and resolved profile options
-> eligible participant plus spawn reason
-> team, world, objective, encounter, and current-player facts
-> preferred-location selection
-> reusable wrap-aware safety search
-> simultaneous position reservation
-> final spawn position and stationary runtime state
-> shared spawn indicator presentation
-> spawn-placement telemetry
```

### Modes And Match Rules

Modes select the player-spawn profile and any future exposed options. They do not own the placement algorithm, dangerous-object checks, reservations, or retry loop. The mode may select different policy for initial spawn, respawn, team placement, objectives, or campaign transitions through the profile contract.

### Lives, Death, Elimination, And Respawn

Lives/death/respawn owns death transitions, respawn eligibility, trigger, delay, and restoration policy. It supplies respawn context and the selected profile reference, then consumes the authoritative spawn outcome. It does not embed preferred-location or safety mechanics.

### Teams And Team Rules

Teams own assignment, membership, and relationship facts. Player-spawn profiles consume those facts to group teammates and separate opponents by default. Team rules do not embed spawn coordinates, placement algorithms, or safety fallback behavior.

### Multiplayer Session And Lifecycle

Lifecycle owns room membership, admission, activation, connection state, removal, and reconnect execution. It determines when a mid-match join or recovery transition may become active, while the profile determines the resulting location.

### Objectives And Campaign Content

Objective or campaign systems may provide objective-adjacent or checkpoint/scripted location context. The profile applies normal safety search unless an explicit scripted override bypasses it. Objective and campaign systems do not become alternate safety authorities.

### Encounter Spawning

[Encounter Spawn Profiles](encounter-spawn-profiles.md) owns enemies, asteroids, hazards, waves, and all other non-player spawn scheduling and policy. Player-spawn profiles are a separate seam and must not reuse encounter safety or placement policy as an implicit player-placement authority.

### Runtime Ship And Player Experience

Runtime ship owns the live avatar state and receives the final stationary spawn state. Player experience renders shared spawn indicators and later presentation feedback, but it does not choose, validate, reserve, or commit the position.

## Implementation Direction

The first implementation slice should preserve existing placement behavior while establishing the profile seam:

```text
1. Define the player-spawn profile identifier and basic_safe_spawn_v1 contract.
2. Preserve the origin-area grid baseline and current outward safety search.
3. Separate reason-specific preferred-location selection from reusable safety search.
4. Pass initial-spawn, respawn, mid-match-join, recovery, checkpoint, and admin/devtool reasons explicitly.
5. Apply wrap-aware regions and safety checks.
6. Add team-aware preferred placement using authoritative team facts.
7. Reserve simultaneous positions without making assignment order gameplay-significant.
8. Retry from the preferred origin/current anchor while re-evaluating world state until successful.
9. Return stationary spawn state with immediate control availability.
10. Preserve the optional profile-owned spawn-protection seam without enabling baseline protection.
11. Emit the baseline placement telemetry fields.
```

Implementation should keep profile policy in a focused player-spawning owner. Modes, lives, teams, lifecycle, encounter, campaign, runtime, devtools, and presentation should route facts or execute their own policy rather than becoming alternate player-spawn authorities. This document describes the planning direction only; it does not claim that any of these steps are implemented.

## Testing Direction

Important future checks:

```text
basic_safe_spawn_v1 is the initial profile identifier
origin-area grid behavior remains the baseline
current outward safety search remains behaviorally preserved
preferred-location selection is separate from reusable safety search
initial spawn and respawn may use different location-selection policy
respawn is not permanently tied to the initial-spawn anchor
all six supported spawn reasons reach explicit profile policy
player-selected spawn is unavailable in the first slice
fixed, formation, teammate, enemy-distance, objective, checkpoint, and random sources are representable
normal safety rejects generic dangerous-object overlap
exact dangerous-object categories can expand without changing the profile seam
other-player overlap is evaluated separately and may be allowed by a profile
team placement groups teammates and separates opponents by default
simultaneous positions are reserved
assignment order does not change gameplay outcomes
failed searches restart from the preferred origin/current anchor
retry behavior re-evaluates current world state
spawn regions and safety checks respect toroidal wrapping
ships spawn stationary and controls start immediately
baseline placement provides no spawn protection
spawn protection remains an optional profile-owned future seam
spawn indicators remain visible to everyone
raw spawn tuning is not exposed at game creation
mid-match join tries respawn logic before initial-spawn fallback
authored/scripted campaign points use normal safety unless explicitly overridden
telemetry records profile, reason, preferred position, final position, safety movement, and retries
```

## Related Docs

- [Gameplay Planning](./!INDEX.md)
- [Modes And Match Rules](modes-and-match-rules.md)
- [Lives, Death, Elimination, And Respawn](lives-death-elimination-and-respawn.md)
- [Teams And Team Rules](teams-and-team-rules.md)
- [Enemies, Bosses, And Encounters](enemies-bosses-and-encounters.md)
- [Encounter Spawn Profiles](encounter-spawn-profiles.md)
- [Objectives And Objective Runtime](objectives-and-objective-runtime.md)
- [Levels, Missions, And Content Structure](levels-missions-and-content-structure.md)
- [Player Experience Systems](player-experience-systems.md)
- [Multiplayer Session And Lifecycle](../platform/multiplayer-session-and-lifecycle.md)
- [Devtools And Telemetry](../../devtools/devtools-and-telemetry.md)

## Remaining Implementation-Level Decisions

- Exact profile type, field, and package names.
- Exact profile registry and resolved-rule representation for `player_spawn_profile_id`.
- Exact preferred-location source ordering and weighting for each spawn reason.
- Exact origin-area grid dimensions, spacing, and current outward-search tuning.
- Exact dangerous-object categories and per-category safety constraints.
- Exact other-player overlap and separation contract for each profile.
- Exact team grouping and opponent-separation distance rules.
- Exact wrap-aware region and distance APIs used by the profile owner.
- Exact simultaneous reservation lifetime, conflict resolution, and release timing.
- Exact retry scheduling, world-state snapshot boundary, and termination safeguards.
- Exact fixed-point, formation, objective, checkpoint, and random-area data shapes.
- Exact scripted-override capability, authorization, and observability requirements.
- Exact stationary runtime-state handoff and control activation event shape.
- Exact future spawn-protection policy, tuning, and balance controls.
- Exact shared spawn-indicator event, timing, and presentation projection.
- Exact telemetry event name, field types, sampling, and retention.
- Exact mode option validation and future selected-option exposure.
- Exact admin/devtool authorization and override boundaries.
- Exact package boundaries and packet/storage shapes chosen at implementation time.

There are no remaining product-level player-spawn decisions blocking P4 system planning.

## Core Invariants

```text
Player Spawn Profiles is the authoritative player-placement and safety planning owner.

basic_safe_spawn_v1 is the initial player-spawn profile.

The origin-area grid baseline and current outward safety search are preserved.

Preferred-location selection is separate from reusable safety search.

Initial spawn and respawn may use different preferred-location policy.

Respawn is not permanently tied to the original initial-spawn anchor.

Initial spawn, respawn, mid-match join, recovery/revival, checkpoint/campaign transition, and admin/devtool spawn are supported reasons.

Player-selected spawn is not available in the first slice.

Safety avoids generic dangerous objects; other-player overlap is a separate concern.

Team placement is mode-selected, with teammates grouped and opponents separated by default.

Simultaneous positions are reserved, and assignment order has no gameplay significance.

Failure to find a safe point retries from the preferred origin/current anchor while re-evaluating world state.

Spawn regions and safety search respect toroidal world wrapping.

Ships spawn stationary and control starts immediately.

Safety placement is sufficient for the baseline; spawn protection is not enabled.

Spawn protection remains an optional profile-owned future seam because it is high balance risk.

Existing on-screen spawn indicators remain visible to everyone for now.

Modes select profiles; raw spawn tuning is not exposed at game creation initially.

Mid-match join uses respawn placement logic first, then baseline initial-spawn fallback.

Authored/scripted points use normal safety unless an explicit scripted override bypasses it.

Telemetry records profile ID, reason, preferred position, final position, safety movement, and retry count.

This doc does not own lives/respawn eligibility, team assignment, encounter spawning, room lifecycle, or match-end policy.
```

# Deferred Player Experience Systems

## Purpose

This doc parks player-experience systems that may become important later, but do not need full planning yet.

These systems are not part of the immediate gameplay/progression foundation. They should remain visible so they are not forgotten, but they should not distract from the core player experience architecture.

Most of these systems are upgrades, afterthoughts, or highly dependant on the outcome of foundational systems.

## Ownership Boundary

This doc owns deferred system stubs only.

It does not define implementation plans, UI layouts, data contracts, reward formulas, persistence schemas, or gameplay rules.

If one of these areas becomes near-term work, it should be promoted into its own focused planning doc or into the existing owner doc that best fits.

## Deferred Systems

### Onboarding And Tutorial

Future home for:

```text
first-launch flow
first guest/profile decision guidance
first match guidance
control teaching
tutorial mission or guided arcade run
first-time menu guidance
```

Likely related docs:

* [Player Experience Systems](player-experience-systems.md)
* [Levels, Missions, And Content Structure](levels-missions-and-content-structure.md)
* [Modes And Match Rules](modes-and-match-rules.md)

Promotion trigger:

```text
when first-run player guidance becomes necessary beyond basic menu defaults
```

### Player Settings, Controls, And Accessibility

Future home for:

```text
input remapping
gamepad/mouse/keyboard preferences
audio settings
video/display settings
HUD scale
readability and visibility options
accessibility options
local settings persistence
```

Likely related docs:

* [Player Experience Systems](player-experience-systems.md)
* [Player Data And Persistence](../platform/player-data-and-persistence.md)

Promotion trigger:

```text
when player-controlled settings need shared ownership, persistence, or more than basic client-local behavior
```

### Player Profile, Records, And Match History

Future home for:

```text
profile summary presentation
personal stats display
recent match history
personal records
high scores
achievement summary surfaces
progression summary surfaces
match result browsing
```

Likely related docs:

* [Player Experience Systems](player-experience-systems.md)
* [Match Outcomes And Results](match-outcomes-and-results.md)
* [Progression And Rewards](progression-and-rewards.md)
* [Achievements And Milestones](achievements-and-milestones.md)
* [Player Data And Persistence](../platform/player-data-and-persistence.md)

Promotion trigger:

```text
when player-facing stats/history screens become more than current profile readout
```

### Reward Reveal And Notifications

Future home for:

```text
post-match reward reveal
achievement completion popups
milestone completion notices
unlock notifications
currency gain presentation
rare reward presentation
claimable reward flow if needed
reward inbox if needed
```

Likely related docs:

* [Progression And Rewards](progression-and-rewards.md)
* [Achievements And Milestones](achievements-and-milestones.md)
* [Match Outcomes And Results](match-outcomes-and-results.md)
* [Shop, Commerce, And Economy](shop-commerce-and-economy.md)

Promotion trigger:

```text
when rewards need a unified player-facing reveal, notification, or claim surface
```

### Live Events And Seasons

Future home for:

```text
time-limited events
season identifiers
event-active challenges
event missions
event reward tracks
limited-time shop offers
event availability windows
event progression gates
```

Likely related docs:

* [Progression And Rewards](progression-and-rewards.md)
* [Achievements And Milestones](achievements-and-milestones.md)
* [Levels, Missions, And Content Structure](levels-missions-and-content-structure.md)
* [Shop, Commerce, And Economy](shop-commerce-and-economy.md)

Promotion trigger:

```text
when time-limited content becomes an actual product or gameplay requirement
```

### Cosmetics And Customization

Future home for:

```text
ship colors
ship skins
profile cosmetics
badges
titles
insignia display
cosmetic ownership
cosmetic selection
cosmetic application rules
```

Likely related docs:

* [Inventory And Hangar](inventory-and-hangar.md)
* [Progression And Rewards](progression-and-rewards.md)
* [Shop, Commerce, And Economy](shop-commerce-and-economy.md)
* [Player Build And Loadouts](player-build-and-loadouts.md)

Promotion trigger:

```text
when cosmetics become more than simple unlock targets or static player/profile display fields
```

## Notes

Consumables and useable items are not listed as a separate deferred system here for now. They should stay under inventory, shop, progression, and player-build planning unless they become active gameplay items with their own pre-match selection, runtime activation, spending, and mode-restriction rules.

## Core Invariants

```text
Deferred systems should stay visible but low-priority.

Deferred systems should not expand the immediate implementation scope.

Existing owner docs remain authoritative until a deferred system is promoted.

Promotion should happen only when the system needs its own ownership boundary.

This doc is a parking lot, not an implementation plan.
```

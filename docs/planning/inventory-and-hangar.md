# Inventory And Hangar

## Purpose

This doc plans the ownership and hangar seam for player-held ships, equipment, and acquisition state.

## Ownership Boundary

This doc owns planning for `HangarInventory`, `OwnedShip`, owned weapons and modules, hardwired modules, acquisition, unlocks, drops, and interaction with `BuildEligibility`.

It should focus on what the player owns and how that owned state feeds selection and validation.

## Current Inputs

- `HangarInventory`
- `OwnedShip`
- owned weapon state
- owned module state
- hardwired module state
- acquisition inputs
- unlock inputs
- drop inputs
- `BuildEligibility`

## Planned Outputs

- a clear inventory ownership model
- the handoff between owned content and selectable build options
- planning for how acquisitions and unlocks affect loadout availability

## Related Docs

- [Systems Plan Index](systems-plan-index.md)
- [Player Experience Systems](player-experience-systems.md)
- [Player Build And Loadouts](player-build-and-loadouts.md)
- [Progression And Rewards](progression-and-rewards.md)

## Open Planning Questions

- Which owned items are persistent versus run-scoped?
- Which acquisition paths should update inventory directly versus through rewards?
- Which unlock state should `BuildEligibility` read without owning the inventory model itself?

# Progression And Rewards

## Purpose

This doc plans the durable progression and reward seam for player advancement and reward grants.

## Ownership Boundary

This doc owns planning for XP and rank, currency, unlocks, reward grants, ship parts, rare drops, and grant routing for local profile versus authenticated account storage.

It should also cover idempotent grant writes so reward events stay safe to replay.

## Current Inputs

- match outcome and result data
- progression eligibility inputs
- XP and rank inputs
- currency inputs
- unlock inputs
- reward grant inputs
- ship part inputs
- rare drop inputs
- local profile routing inputs
- authenticated account routing inputs
- idempotent grant write inputs

## Planned Outputs

- progression award planning boundaries
- reward grant routing expectations
- durable update ownership between local profile and authenticated account paths

## Related Docs

- [Systems Plan Index](systems-plan-index.md)
- [Player Experience Systems](player-experience-systems.md)
- [Match Outcomes And Results](match-outcomes-and-results.md)
- [Player Data And Persistence](player-data-and-persistence.md)
- [Anti-Cheat And Trust Policy](anti-cheat-and-trust-policy.md)
- [API Product Surface](api-product-surface.md)
- [Achievements And Milestones](achievements-and-milestones.md)
- [Inventory And Hangar](inventory-and-hangar.md)

## Open Planning Questions

- Which rewards are immediate versus queued for later grant handling?
- Which award types apply to local profile only, authenticated account only, or both?
- Which idempotency key is the stable source of truth for reward replay safety?

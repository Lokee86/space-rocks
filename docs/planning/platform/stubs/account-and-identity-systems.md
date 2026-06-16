# Account And Identity Systems

## Purpose

This doc plans the identity and account seam across guest, local profile, and authenticated account paths.

## Ownership Boundary

This doc owns planning for guest identity, local profile identity, authenticated account identity, OAuth expansion, account linking, and local-to-online migration questions.

It should stay focused on account and identity policy, not realtime gameplay or UI layout.

## Current Inputs

- guest identity inputs
- local profile identity inputs
- authenticated account identity inputs
- OAuth expansion inputs
- account linking inputs
- local-to-online migration inputs

## Planned Outputs

- identity and account state boundaries
- routing expectations between guest, local profile, and authenticated account paths
- questions that must be answered before migration or linking policy is finalized

## Related Docs

- [Systems Plan Index](systems-plan-index.md)
- [Player Experience Systems](player-experience-systems.md)
- [Player Data And Persistence](player-data-and-persistence.md)
- [Social And Community Systems](social-and-community-systems.md)
- [API Product Surface](api-product-surface.md)
- [Progression And Rewards](progression-and-rewards.md)

## Open Planning Questions

- Which identity transitions should be allowed without account creation?
- Which parts of local profile state should migrate into online accounts later?
- Which account linking flow should be treated as the long-term default?

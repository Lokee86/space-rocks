# Player Experience Systems

## Purpose

This is the umbrella planning map for the player-facing flow from identity and profile selection through match setup, objective play, outcomes, and the next available options.

It connects:

- identity/profile
- mode selection
- eligible build options
- match start
- objective play
- match outcome
- rewards/progression
- next available options

## Ownership Boundary

This doc coordinates the broad player experience sequence across related system plans.

It does not define final UI layouts, menu screens, or concrete runtime policy. Those details belong in the narrower system docs linked here.

## Current Inputs

- player identity and profile state
- mode availability and room rules
- build eligibility and loadout selection
- trusted match state and outcome data
- progression and reward grant inputs

## Planned Outputs

- an ordered map of the player-facing system flow
- clear references to the owning planning docs for each step
- shared vocabulary for the handoff between systems

## Related Docs

- [Systems Plan Index](systems-plan-index.md)
- [Account And Identity Systems](account-and-identity-systems.md)
- [Player Data And Persistence](player-data-and-persistence.md)
- [Leaderboards And Rankings](leaderboards-and-rankings.md)
- [Modes And Match Rules](modes-and-match-rules.md)
- [Inventory And Hangar](inventory-and-hangar.md)
- [Match Outcomes And Results](match-outcomes-and-results.md)
- [Progression And Rewards](progression-and-rewards.md)
- [Achievements And Milestones](achievements-and-milestones.md)
- [Player Build And Loadouts](player-build-and-loadouts.md)

## Open Planning Questions

- Which system should own the first authoritative handoff after profile selection?
- Which match-state inputs are required before outcome and reward planning can stay decoupled?
- Which parts of the player-facing flow should remain shared across local profile and authenticated account paths?

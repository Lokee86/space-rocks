# Leaderboards And Rankings

## Purpose

This doc plans the leaderboard and ranking seam for online player comparison and persistence.

## Ownership Boundary

This doc owns planning for online leaderboard eligibility, mode and result dependency, authenticated account requirement, anti-cheat and trust dependency, rankings, and API persistence.

It should remain separate from realtime simulation and from the core progression grant path.

## Current Inputs

- leaderboard eligibility inputs
- mode dependency inputs
- result dependency inputs
- authenticated account requirement inputs
- anti-cheat and trust inputs
- ranking inputs
- API persistence inputs

## Planned Outputs

- leaderboard eligibility boundaries
- ranking ownership boundaries
- the dependency set needed before persistence is trusted online

## Related Docs

- [Systems Plan Index](systems-plan-index.md)
- [Account And Identity Systems](account-and-identity-systems.md)
- [Anti-Cheat And Trust Policy](anti-cheat-and-trust-policy.md)
- [Match Outcomes And Results](match-outcomes-and-results.md)
- [Progression And Rewards](progression-and-rewards.md)

## Open Planning Questions

- Which modes are eligible for online leaderboards?
- Which result facts are required before a ranking write is valid?
- Which trust checks are required before API persistence is allowed?

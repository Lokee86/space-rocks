# Anti-Cheat And Trust Policy
Parent index: [Platform Planning](../!README.md)

## Purpose

This doc plans the trust-policy seam for online progression, rankings, and abuse resistance.

## Ownership Boundary

This doc owns planning for progression eligibility, leaderboard eligibility, devtools and debug exclusion, local profile trust limits, online-authoritative rules, and future anti-farming policy.

It should define trust boundaries, not the enforcement implementation.

## Current Inputs

- progression eligibility inputs
- leaderboard eligibility inputs
- devtools/debug exclusion inputs
- local profile trust-limit inputs
- online-authoritative inputs
- anti-farming inputs

## Planned Outputs

- trust-policy boundaries
- eligibility expectations for online-facing systems
- a shared vocabulary for trusted versus untrusted player facts

## Related Docs

- [Systems Plan Index](systems-plan-index.md)
- [Account And Identity Systems](account-and-identity-systems.md)
- [Leaderboards And Rankings](leaderboards-and-rankings.md)
- [Progression And Rewards](progression-and-rewards.md)
- [Shop, Commerce, And Economy](shop-commerce-and-economy.md)
- [Match Outcomes And Results](match-outcomes-and-results.md)

## Open Planning Questions

- Which local-profile facts are never trusted for online systems?
- Which devtools or debug routes must always be excluded from trust-sensitive flows?
- Which anti-farming rules should be planned now versus deferred until a stronger online model exists?

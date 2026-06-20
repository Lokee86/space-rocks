# API Product Surface
Parent index: [Platform Planning](../!INDEX.md)

## Purpose

This doc plans the API-owned account and profile surface outside the realtime game-server.

## Ownership Boundary

This doc owns planning for API-owned profile and account surfaces, account product concerns, online persistence, account-owned match history, unlocks, currency, and leaderboards.

It should stay separated from realtime simulation and from game-server transport concerns.

## Current Inputs

- profile surface inputs
- account surface inputs
- online persistence inputs
- account-owned match history inputs
- unlock inputs
- currency inputs
- leaderboard inputs
- account product inputs

## Planned Outputs

- API product ownership boundaries
- the split between account surfaces and realtime gameplay
- persistence expectations for account-owned data

## Related Docs

- [Systems Plan Index](systems-plan-index.md)
- [Account And Identity Systems](../account-and-identity-systems.md)
- [Social And Community Systems](social-and-community-systems.md)
- [Website And Web Presence](website-and-web-presence.md)
- [Shop, Commerce, And Economy](shop-commerce-and-economy.md)
- [Player Data And Persistence](player-data-and-persistence.md)
- [Progression And Rewards](progression-and-rewards.md)
- [Leaderboards And Rankings](leaderboards-and-rankings.md)

## Open Planning Questions

- Which surfaces should be API-owned instead of client-local?
- Which account data is durable enough to support online product features?
- Which match-history facts belong in the API product surface rather than in gameplay results?

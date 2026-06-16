# Social And Community Systems

## Purpose

This doc owns planning for friends, blocks, mutes, invites, parties, recent players, presence, profile visibility, account display identity, social graph privacy, future chat boundaries, and guilds or clans as a deferred possible system.

## Ownership Boundary

This doc owns how players relate to other players.

`account-and-identity-systems.md` owns who the player or account is.

`matchmaking-and-room-discovery.md` owns finding and joining playable rooms.

`leaderboards-and-rankings.md` owns competitive comparison.

`api-product-surface.md` owns backend API product and account endpoints.

V0 social should stay small:

- display name and callsign relationship
- recent players
- block list
- mute list
- friend invites later
- party invites later
- presence later
- chat deferred

Chat would require moderation and reporting planning before implementation.

## Current Inputs

- friend relationship inputs
- block list inputs
- mute list inputs
- invite inputs
- party inputs
- recent player inputs
- presence inputs
- profile visibility inputs
- display identity inputs
- social graph privacy inputs
- future chat boundary inputs
- deferred guild and clan inputs

## Planned Outputs

- social relationship planning boundaries
- V0 social scope boundaries
- the split between player identity and player relationships
- deferred future-scope questions for chat and guilds

## Related Docs

- [Systems Plan Index](systems-plan-index.md)
- [Account And Identity Systems](account-and-identity-systems.md)
- [Matchmaking And Room Discovery](matchmaking-and-room-discovery.md)
- [API Product Surface](api-product-surface.md)
- [Player Experience Systems](player-experience-systems.md)

## Open Planning Questions

- Which social features should ship with the smallest useful V0?
- Which relationship data should remain visible without requiring chat or parties?
- Which future social systems should be split into dedicated owner docs first?

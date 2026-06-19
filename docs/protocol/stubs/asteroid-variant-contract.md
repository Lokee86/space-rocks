# Asteroid Variant Contract
Parent index: [Protocol](../!README.md)

## Purpose

This stub is incomplete and non-canonical.
It points to the cross-service asteroid variant id contract boundary.

## Overview

This stub tracks the contract for asteroid variant ids exchanged between the game server and the client.

The server sends asteroid variant ids in asteroid state. The client consumes those ids as presentation selectors for texture and scene-level variant presentation.

## Ownership boundary

The intended ownership boundary is:

* game server owns authoritative asteroid variant assignment and state export
* client owns presentation-side consumption of the variant id
* shared data owns the source catalog used to keep ids stable across services

This contract should remain focused on the variant id itself, not on spawning behavior or client scene details.

## Compatibility expectations

Asteroid variant ids should remain stable enough for client and server to agree on meaning across generated outputs.

At a stub level, compatibility means:

* the server exports variant ids in asteroid state
* the client can consume known ids as presentation selectors
* generated client and server outputs stay aligned with the shared variant data source
* unknown or stale ids should be treated as a mismatch to investigate, not as a new protocol shape

## Conceptual links

* [Client asteroid variant presentation](../../services/client/world-sync/asteroid-variant-presentation.md)
* [Asteroid variants data](../../data/stubs/asteroid-variants-data.md)
* [Server asteroid spawning and variants](../../services/game-server/simulation/world/stubs/asteroid-spawning-and-variants.md)

## Notes

This is a scaffold only.
It does not define packet schema details or runtime spawning mechanics.

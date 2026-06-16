# Shop, Commerce, And Economy

## Purpose

This doc owns planning for soft currency, premium currency if ever used, earned currency, purchased currency, the shop catalog, prices, offers, purchase receipts, refund handling, entitlements, owned item grants, cosmetic purchases, ship or weapon or module unlock purchases, battle pass or season pass as deferred possible systems, DLC-like packs as deferred possible systems, real-money transaction boundaries, and platform or store policy questions.

## Ownership Boundary

This doc owns pricing, catalog, commerce, purchases, receipts, refunds, and entitlement grants.

`progression-and-rewards.md` owns gameplay reward grants and earned currency sources.

`inventory-and-hangar.md` owns durable owned items and unlock state.

`api-product-surface.md` owns backend API and product endpoints for account commerce data.

`anti-cheat-and-trust-policy.md` owns abuse, fraud, dupes, and third-party real-money trading concerns.

`deployment-and-packaging.md` owns store and platform packaging constraints.

Realtime gameplay should consume resolved ownership and build eligibility and should not care whether ownership came from reward, achievement, unlock, purchase, DLC, or admin grant.

V0 should not include:

- real-money transactions
- premium currency
- paid loot boxes
- battle pass
- player-to-player trading
- transferable items
- shop UI

## Current Inputs

- soft currency inputs
- premium currency inputs
- earned currency inputs
- purchased currency inputs
- shop catalog inputs
- price inputs
- offer inputs
- receipt inputs
- refund inputs
- entitlement inputs
- owned item grant inputs
- cosmetic purchase inputs
- ship unlock purchase inputs
- weapon unlock purchase inputs
- module unlock purchase inputs
- deferred battle pass inputs
- deferred season pass inputs
- deferred DLC pack inputs
- real-money transaction boundary inputs
- platform and store policy inputs

## Planned Outputs

- commerce ownership boundaries
- catalog and pricing planning boundaries
- entitlement and refund planning boundaries
- deferred future commerce-system questions

## Related Docs

- [Systems Plan Index](systems-plan-index.md)
- [Progression And Rewards](progression-and-rewards.md)
- [Inventory And Hangar](inventory-and-hangar.md)
- [API Product Surface](api-product-surface.md)
- [Anti-Cheat And Trust Policy](anti-cheat-and-trust-policy.md)
- [Deployment And Packaging](deployment-and-packaging.md)
- [Player Experience Systems](player-experience-systems.md)

## Open Planning Questions

- Which commerce flows should remain purely earned rather than purchased?
- Which platform or store policies must be settled before any real-money features are considered?
- Which future commerce systems should be split into dedicated owner docs first?

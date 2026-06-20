# Shop, Commerce, And Economy
Parent index: [Gameplay Planning](./!INDEX.md)

## Purpose

This doc plans the economy and commerce seam for Space Rocks.

It owns currencies, wallet policy, shop catalog offers, pricing, purchases, receipts, refunds, sellback, entitlements, and future real-money transaction boundaries.

The goal is to keep value movement explicit and route all ownership changes through the same grant path used by progression and rewards. Runtime gameplay should consume resolved ownership and build eligibility, not commerce internals.

## Ownership Boundary

This doc owns:

```text
currency definitions
wallet policy
transaction-backed balance policy
shop catalog offer policy
pricing and cost policy
purchase validation flow
purchase receipts
refund policy
sellback / resell policy
stack cap and prorated purchase policy
future exceptional currency boundary
future RMT boundary
entitlement boundary
admin/devtools commerce mutation lane
guest/local/online commerce separation
```

This doc does not own:

```text
XP reward formulas
earned-currency reward amounts
achievement definitions
milestone definitions
inventory persistence details
owned ship/weapon/module instance models
runtime pickup effects
loadout eligibility
shop UI layout
physical database schema
API route names
payment-provider implementation details
exact legal refund policy by jurisdiction
```

Progression and rewards owns earned-currency grant policy.

Inventory and hangar owns owned item state and grant application into owned ships, weapons, modules, and stackable inventory.

Player-data owns identity routing, physical persistence, atomic writes, and storage-specific receipt/transaction mechanics.

The SSoT/data pipeline owns static shop catalog and offer definitions.

The client owns display and presentation only.

V0 should not include:

```text
real-money transactions
premium currency
paid loot boxes
battle pass
season pass
player-to-player trading
transferable items
```

## Core Architecture

Commerce validates purchases and produces grant awards.

Commerce does not directly create owned ships, owned weapons, owned modules, titles, unlocks, or stackable inventory.

Core purchase flow:

```text
ShopOffer
-> purchase validation
-> wallet transaction
-> purchase receipt
-> GrantAward source_type shop_purchase
-> player-data route
-> inventory/progression grant application
-> updated player commerce state
```

Core entitlement flow:

```text
Entitlement Source
-> entitlement validation
-> entitlement record
-> GrantAward source_type entitlement
-> player-data route
-> inventory/progression grant application
```

Purchases, entitlements, admin grants, migration grants, and future RMT grants should all converge on the same durable grant path after their owning validation step.

## Currency And Wallet Policy

V0 uses one earned soft currency:

```text
currency.orebits
```

Rules:

```text
currency amounts are integer values
currency refs are stable content refs
normal purchases cannot create negative balances
wallet mutations are transaction-backed
client state is display-only
server/player-data validation is authoritative
```

Progression owns how Orebits are earned.

Commerce owns how Orebits are spent.

A future version may add a second currency. That currency should not be treated as premium currency by default. The intent is an exceptional or rare currency class, not necessarily an RMT-only currency.

Possible uses for a future exceptional currency:

```text
rare or exotic unlocks
high-end modules or ships
hardwire install/removal fees
event or milestone rewards
special shop offers
rare cosmetic purchases
late-game currency sinks
```

Rules for future additional currencies:

```text
V0 still uses only currency.orebits
the second currency is deferred
the second currency uses the same wallet and transaction model
the second currency has an explicit currency_ref
the second currency does not imply RMT support
if RMT ever grants it, that grant must flow through verified RMT/entitlement handling
```

Wallet balances are durable player state.

Balance changes should go through logged transactions. Non-RMT transaction logs may be pruned or compacted by scheduled jobs. RMT records, if ever added, are permanent audit-grade records.

## Catalog Source Of Truth

Static shop catalog and offer definitions belong to the project SSoT/data pipeline.

They should not be authored as player-data database rows.

The intended split is:

```text
SSoT structured data
-> generated game-server authoritative offer data
-> generated client display data
-> possible generated refs for persistence/API validation
```

Game-server generated catalog data is authoritative for purchase validation.

Client generated catalog data is display-only.

Player-data persistence stores player-specific commerce facts, not static catalog definitions.

Player-specific commerce state includes:

```text
wallet balances
currency transactions
purchase receipts
refund/reversal state
sellback records
owned items
entitlement records
future RMT audit records
```

A database-owned catalog may be useful later only if the project needs live-ops behavior such as remotely edited offers, rotating store inventory without deploys, platform-specific overrides, experiments, or admin-managed sale events.

That is not the V0 architecture.

## Shop Offers

A `ShopOffer` is a static catalog-defined purchase option.

It packages cost, requirements, visibility, purchase policy, and grants.

Likely offer fields:

```text
offer_id
catalog_version
offer_scope
costs[]
grant_bundle[]
requirements[]
purchase_policy
visible_if_unavailable
max_quantity optional
metadata optional
```

`visible_if_unavailable` is a single boolean.

```text
true
-> show the offer even when locked, unaffordable, or otherwise unavailable

false
-> hide the offer unless currently available
```

Normal progression-visible items should usually be visible when unavailable so the player can see requirements.

Special, event, secret, or conditional offers may remain hidden until available.

Likely offer scopes:

```text
shared
local_only
online_only
event_later
rmt_later
```

Likely purchase policies:

```text
one_time
repeatable
limited_count
stackable
capped_prorated_stack
```

Likely offer requirements:

```text
unlocked content
level requirement
achievement requirement
milestone requirement
mode completion requirement
owned item requirement
not already owned
event active
account type
platform entitlement
```

Requirements should stay generic enough to avoid hardcoding shop behavior to only unlock checks.

## Unlocks, Access, And Ownership

Unlocks do not equal ownership.

Unlocks allow access, visibility, purchase eligibility, or future acquisition paths.

Ownership comes from grants.

Examples:

```text
unlock weapon.railgun
-> weapon.railgun can appear as purchasable or selectable according to rules

purchase offer.weapon.railgun.purchase
-> emits inventory_item weapon.railgun
-> inventory creates OwnedWeapon
```

```text
unlock ship.scout
-> ship.scout can appear in the shop or acquisition path

purchase offer.ship.scout.purchase
-> emits inventory_item ship.scout
-> inventory creates OwnedShip
```

Commerce validates purchase eligibility.

Inventory applies ownership grants.

BuildEligibility consumes normalized inventory/access state later.

## Purchase Flow

V0 does not need a quote flow.

The catalog is authoritative game data generated from the SSoT. Purchases validate against the current authoritative offer state at execution time.

Purchase flow:

```text
Client requests offer purchase
-> player identity route is resolved
-> authoritative offer data is loaded
-> offer scope and availability are validated
-> requirements are validated
-> purchase policy and limits are validated
-> wallet balance is validated
-> receipt is created
-> wallet transaction is applied
-> GrantAward source_type shop_purchase is emitted
-> grants are applied through player-data/inventory/progression seams
-> updated wallet, receipt, and grant result are returned
```

The client never decides price, eligibility, purchase result, grant result, ownership state, or wallet state.

Purchases should be idempotent so duplicate request or receipt replay cannot double-spend or double-grant.

## Receipts And Retention

Every completed purchase creates a receipt.

Receipts should preserve enough information to explain:

```text
which offer was purchased
which catalog version was used
what it cost
what grants were produced
whether it was refunded, reversed, or sold back later
```

Record retention policy:

```text
RMT records
-> permanent audit-grade records if RMT ever exists

Non-RMT economy records
-> DB logged
-> pruned or compacted by scheduled jobs

Purchase receipts / entitlement / refund records
-> durable enough to explain ownership and reversals
```

Exact receipt schema, retention periods, pruning windows, compaction shape, and scheduled job names are implementation decisions.

## Refunds, Sellback, And Reversals

Refund and sellback are separate concepts.

### Refund

A refund reverses a recent eligible purchase.

Currency refunds have a very short availability window, likely only while the shop remains open.

After the refund window closes, the item cannot be refunded through normal currency refund policy.

Consumed items cannot be refunded.

Rules:

```text
refunds are short-window purchase reversals
refunds create explicit reversal records
refunds should not silently erase purchase history
consumed items cannot be refunded
consumed RMT consumables cannot be refunded through normal game policy if they ever exist
```

If future platform or legal RMT requirements force reversal after consumption, the system records an explicit correction rather than pretending the purchase never happened.

### Sellback / Resell

Sellback is a later sale of owned inventory back to the shop.

Sellback is not a refund.

Any owned item can be sold back unless item state blocks it.

Consumed items cannot be sold because they no longer exist in inventory.

Expected sellback return is likely 30-50% of purchase or base value. Exact value is a gametime economy decision.

Sellback creates an item removal/reversal and a currency transaction.

## Stack Caps And Repair Credit

Stack caps are hard caps.

If a stackable purchase would exceed the max, the purchase may clamp to the remaining capacity and prorate the cost.

General partial purchases should be avoided.

Prorated capped stack purchases are the only planned partial-like exception.

Rules:

```text
stackable max is a hard max
purchase may clamp quantity to remaining capacity
purchase price may prorate to actual quantity granted
no unrelated partial purchase behavior
consumed stackables cannot be refunded
```

Wallet balances normally cannot go negative.

Ship repair credit may allow a negative Orebits balance if that policy is enabled.

Repair credit is a narrow exception.

Rules:

```text
normal purchases require sufficient spendable balance
repair transactions may allow balance below zero
negative balance naturally blocks normal purchases
future Orebits earnings pay the wallet back toward zero
all ships may be repairable on credit if repair-credit policy is enabled
```

Exact repair costs and whether repair credit is enabled are gametime decisions.

## RMT And Entitlement Boundary

V0 has no real-money transactions.

V0 also has no:

```text
premium currency
paid loot boxes
battle pass
season pass
paid consumables
```

Future RMT, if added, must remain separated from normal soft-currency purchases but still resolve through the same grant path after verification.

Future RMT rules:

```text
authenticated online accounts only
never guest
never local-profile-only
verified by platform/store/payment provider
records are permanent audit-grade records
refunds follow platform/legal requirements
grants still flow through GrantAward
RMT does not create a second ownership path
```

Future RMT flow:

```text
platform/store purchase
-> verified external receipt
-> RMT receipt or entitlement record
-> GrantAward source_type entitlement or shop_purchase
-> normal player-data and inventory grant application
```

Entitlements represent ownership or access granted by an external or non-normal-shop source.

Possible entitlement sources:

```text
DLC or founder packs
platform grants
promotional grants
external product ownership
account migration grants
admin grants
future RMT products
```

Entitlement records should preserve enough information to prove why access or ownership exists and whether the entitlement was later reversed.

## Guest, Local, And Online Commerce

Guest, Local Profile, and Authenticated Account commerce can use the same logical contracts with different storage routes.

```text
Guest
-> transient commerce state only
-> no RMT
-> no permanent external entitlement
-> may copy to Local Profile only through explicit profile creation behavior
```

```text
Local Profile
-> local wallet
-> local shop
-> local receipts
-> local owned state
-> no import into multiplayer or online account commerce
```

```text
Authenticated Account
-> online wallet
-> online shop
-> online receipts
-> online entitlements
-> online source of truth
```

Local and online shops may exist at the same time.

Local commerce state stays local.

Online commerce state stays online.

There is no Local Profile import into multiplayer or online commerce.

## Admin And Devtools

Admin and devtools economy mutations should share the same controlled mutation lane, with different source-tracking tags.

They should not pretend to be normal purchases or rewards.

Likely source tags:

```text
admin_grant
admin_adjustment
devtools_test_grant
devtools_reset
devtools_purchase_simulation
```

Admin and devtools actions may create wallet transactions, grants, resets, corrections, or test purchase simulations.

They should remain distinguishable from player-earned or player-purchased value.

## Implementation Planning

Recommended implementation sequence:

```text
1. Define currency refs and wallet contract.
2. Add transaction-backed Orebits wallet.
3. Define shop catalog SSoT shape.
4. Generate authoritative game-server offer data and client display data.
5. Define ShopOffer model.
6. Define purchase receipt behavior.
7. Add purchase validation path.
8. Wire purchases into GrantAward source_type shop_purchase.
9. Apply purchase grants through inventory/hangar.
10. Add refund eligibility/reversal seam.
11. Add sellback/resell seam.
12. Add stack cap and prorated purchase handling.
13. Add repair-credit/negative-wallet exception if repair needs it.
14. Add entitlement record seam.
15. Keep RMT provider integration deferred but bounded.
```

First useful slice:

```text
currency.orebits wallet
+ one static SSoT shop offer
+ purchase receipt
+ wallet spend
+ GrantAward source_type shop_purchase
+ inventory grant application
```

This proves the commerce path without needing shop UI, RMT, premium currency, live catalog editing, or platform payment integration.

## Testing Direction

Important future tests:

```text
purchase succeeds with sufficient Orebits
purchase fails with insufficient Orebits
purchase fails when requirements are unmet
hidden unavailable offers are omitted
visible unavailable offers show blocked reasons
duplicate one-time purchases are rejected
repeatable purchases work
limited-count purchases enforce limit
capped stack purchase clamps and prorates
general partial purchases are not allowed
consumed items cannot be refunded
sellback is not treated as refund
sellback removes owned item and creates currency transaction
negative repair wallet blocks normal purchases
guest commerce remains transient
local commerce does not import to online account
online and local shops can coexist
client catalog is display-only
purchase uses authoritative game-server catalog data
duplicate receipt/idempotency replay does not double-spend or double-grant
admin/devtools mutations use explicit source tags
```

## Related Docs

* [Progression And Rewards](progression-and-rewards.md)
* [Inventory And Hangar](inventory-and-hangar.md)
* [Achievements And Milestones](achievements-and-milestones.md)
* [Player Build And Loadouts](player-build-and-loadouts.md)
* [Player Data And Persistence](../platform/player-data-and-persistence.md)
* [API Product Surface](../platform/api-product-surface.md)
* [Anti-Cheat And Trust Policy](../platform/anti-cheat-and-trust-policy.md)
* [Deployment And Packaging](../technical/deployment-and-packaging.md)
* [Data Sync And Ssot Pipeline](../technical/data-sync-and-ssot-pipeline.md)

## Open Gametime Decisions

```text
exact Orebits prices
exact Orebits earning rates
exact starter wallet balance
exact sellback percentage
exact currency refund window
exact repair-credit behavior
whether ship repair credit is enabled
exact future exceptional currency name and use
exact catalog file format inside the SSoT pipeline
exact generated data shape
exact purchase API route names
exact physical database schema
exact transaction pruning and compaction schedule
exact entitlement provider vocabulary
exact future RMT provider/SKU mappings
exact future platform refund handling
exact first shop UI layout
```

## Core Invariants

```text
Commerce owns catalog, pricing, wallet policy, purchases, receipts, refunds, sellback, entitlements, and RMT boundaries.
Static catalog data comes from the SSoT/data pipeline, not DB rows.
Generated game-server catalog data is authoritative.
Generated client catalog data is display-only.
Client is never authoritative for prices, purchases, wallet, or ownership.
Purchases produce GrantAwards.
Inventory ownership comes from grants.
Unlocks do not equal ownership.
Wallet mutations are transaction-backed.
Normal purchases cannot create negative balances.
Repair credit is the only planned negative-wallet exception.
Stack caps are hard caps.
Capped stack purchases may clamp and prorate.
Consumed items cannot be refunded.
Refunds are short-window reversals.
Sellback is a separate economy action.
RMT is not part of V0.
RMT, if ever added, uses permanent records and the same grant path.
Guest commerce is transient.
Local commerce and online commerce can coexist.
Local Profile commerce never imports into multiplayer or online account commerce.
Admin/devtools mutations use explicit source-tracking tags.
Runtime gameplay consumes resolved ownership and build eligibility, not commerce internals.
```

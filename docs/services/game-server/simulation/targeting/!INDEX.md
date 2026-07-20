---
author: brian
created: "2026-07-19"
document_id: 019f7d55-fb2c-7647-9062-6bfc664aa9da
document_type: general
policy_exempt: false
summary: Targeting documentation for the game server lives here.
---
# Game Server Simulation Targeting

Parent index: [Game Server Simulation](../!INDEX.md)

Targeting documentation for the game server lives here.

## Ownership

This folder owns canonical target state, target selection, and target status read model documentation for the game server simulation.

## Does Not Belong

- Room membership or lifecycle rules.
- WebSocket transport details.
- External integration internals.
- Process startup or shutdown.
- Combat damage or scoring rules.

## Direct Files
<!-- doc-ledger:files:start -->

- [canonical-target-state.md](canonical-target-state.md) - Canonical target state documentation.
- [target-selection-and-status.md](target-selection-and-status.md) - Target selection and target status handling documentation.
<!-- doc-ledger:files:end -->
## Stub Files
<!-- doc-ledger:stubs:start -->
<!-- doc-ledger:stubs:end -->
## Direct Folders
<!-- doc-ledger:folders:start -->
<!-- doc-ledger:folders:end -->
## Related Docs

- [Game Server Simulation](../!INDEX.md)
- [Game Server](../../!INDEX.md)
- [Services index](../../../!INDEX.md)

## Notes

This boundary stays on target ownership and target read models rather than combat outcomes.
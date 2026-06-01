# Client Mouse Input Reference

Mouse input should route through semantic actions, not scattered raw click checks across gameplay scripts.

Related references:

- [Canonical targeting reference](../server/targeting.md)
- [Devtools toggles and player-target controls](../devtools/toggles.md)

## Semantic Actions

Current semantic mouse/input concepts:

- select target
- deselect target
- spawn/place entity
- cancel pending mouse action

Raw mouse button checks should stay at input binding/translation edges, then map into these semantic actions.

## Priority Rules

Pending placement/action has priority over generic target behavior.

- Pending placement/action owns left click until it is completed or canceled.
- Right click or Escape cancels pending action before generic deselect behavior runs.

## Target Selection and Server Confirmation

Target selection is requested from client input, then confirmed by canonical server-driven gameplay state.

- Client selection actions send selection/deselection intent.
- Canonical target identity is read from authoritative gameplay state updates (`target_kind` + `target_id`), not from local click assumptions.

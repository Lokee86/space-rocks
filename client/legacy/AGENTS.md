# Legacy Quarantine

`client/legacy` contains quarantined code only.

Do not add new feature behavior under `client/legacy`.

Only active API or facade layers outside `client/legacy` may depend on legacy code.

Every child quarantine folder must include `API.md`.

Before editing compatibility code for a quarantine folder, read that folder's `API.md`.

If `API.md` is missing, stop and report that the API location is undocumented.

## Black-box Rule

Files under `client/legacy` are not normal active source files.

Do not add behavior, features, policy, or new design decisions under `client/legacy`.

Legacy code may only be touched for:

- path or import repairs after file moves
- compatibility methods required by an active API
- behavior-preserving fixes that keep the legacy implementation working behind its API

If a task seems to require new logic in legacy, stop and propose or modify the active API instead.

Allowed edits:

- import or path fixes required by file moves
- behavior-preserving compatibility edits required by active APIs
- mechanical changes requested by active APIs

Forbidden edits:

- new gameplay rules
- new camera behavior
- new render-anchor policy
- new HUD, debug, or targeting behavior
- new feature work

## Direct Import Rule

Active code must not import files from `client/legacy` directly unless the quarantine folder's `API.md` explicitly allows that active file or folder.

For `player_render`, only `client/scripts/world/player_render/*.gd` may import `res://legacy/player_render/`.

## Review Checklist

- Am I editing legacy code?
- Did I read the child folder `API.md`?
- Is this edit behavior-preserving?
- Could this behavior belong in the active API instead?
- Am I adding a direct legacy import outside the allowed API folder?

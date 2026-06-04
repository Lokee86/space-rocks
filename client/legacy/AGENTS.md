# Legacy Quarantine

`client/legacy` contains quarantined code only.

Do not add new feature behavior under `client/legacy`.

Only active API or facade layers outside `client/legacy` may depend on legacy code.

Every child quarantine folder must include `API.md`.

Before editing compatibility code for a quarantine folder, read that folder's `API.md`.

If `API.md` is missing, stop and report that the API location is undocumented.

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

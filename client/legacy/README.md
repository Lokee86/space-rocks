# Legacy Client Quarantine

`client/legacy` is a quarantine area for required legacy client code.

New feature code should use active APIs under `client/scripts` instead of importing legacy code directly.

## Rules

- do not add new responsibilities here
- do not build new features here
- do not use this folder as a normal extension point
- keep edits behavior-preserving unless explicitly requested
- prefer adding or changing API files outside this folder

## API References

- each child folder must include `API.md`
- `API.md` points to the active replacement API

## Current Quarantines

- `player_render/`: tangled legacy player/render sync code

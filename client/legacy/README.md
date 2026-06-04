# Legacy Client Quarantine

`client/legacy` is a quarantine area for required legacy client code.

New feature code should use active APIs under `client/scripts` instead of importing legacy code directly.

## Purpose

`client/legacy` exists to contain code that is still required but should not be used as an active extension point.

## How to Use This Folder

- New features go outside `client/legacy`.
- New architecture seams go outside `client/legacy`.
- New behavior should be added to active API or facade layers.
- Legacy folders are allowed to remain messy internally because the goal is containment, not cleanup.

## Required `API.md` File

Every quarantine folder must contain `API.md` pointing to the active replacement API.

## Rules

- do not add new responsibilities here
- do not build new features here
- do not use this folder as a normal extension point
- keep edits behavior-preserving unless explicitly requested
- prefer adding or changing API files outside this folder

## Current Quarantines

- `player_render/`: tangled legacy player/render sync code that combines:
  - player lifecycle
  - player visual placement
  - view target state
  - interpolation support
  - legacy visual anchor math
  - server-to-visual mapping

`player_render` is intentionally ignored by normal feature work.

# Camera / View-Origin Planning

This document is future planning only. It does not describe the current runtime architecture, and it should not be used as a shortcut for a small refactor.

## Current Constraint

- `Camera2D` currently stays under `Player`.
- Spectate currently works by using the local `Player` as the camera carrier.
- The local player is moved and hidden to follow the selected remote player.

## Failed Refactor Lesson

An independent `CameraAnchor` can move the camera correctly.

That alone is not enough.

Rendered gameplay does not follow, because much of presentation is still local-player-relative.

The result is that the camera sees the background while gameplay remains centered around the local-player presentation origin.

## Future Required Refactor

The long-term fix is to introduce an active view origin.

- Normal origin: local player
- Spectate origin: selected remote player
- Future origins may include free camera or cinematic targets

Players, bullets, asteroids, hitboxes, offscreen indicators, target positions, and the background reference must all render relative to the same active origin.

## Warning

- Do not move `Camera2D` out from under `Player` as a small refactor.
- Do not replace the current carrier model until active view origin is implemented.

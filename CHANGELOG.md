# Changelog

All notable player-facing changes to Space Rocks are recorded here.

## [0.1.2] - Unreleased

### Added

- Server-owned player bots with lobby add/remove controls and normal gameplay, death, respawn, scoring, and team participation.
- Team game support across room creation, lobby assignment, gameplay state, HUD presentation, and match results.
- Player-controlled camera zoom with world-space background scaling.

### Changed

- Reworked realtime ship, asteroid, bullet, and lifecycle lanes to remain responsive under multiplayer load.
- Refined the multiplayer lobby, team selectors, player rows, action buttons, and match-results layout to match the transmission-screen interface.
- Devtools commands now immediately return authoritative debug status after applying a state change.
- Moved client log maintenance out of the startup-critical path.

### Fixed

- Bots now remain active when a completed multiplayer game returns to the lobby and starts another match.
- Team IDs and shader-derived team colours now remain consistent from lobby assignment through gameplay and results.
- Free-for-all player hues no longer flicker or shift when players die and respawn.
- Final player ships finish despawning after game over instead of reappearing behind the results window.
- Camera zoom input is ignored after game over.
- Match-result team swatches remain square and use the same base colour and hue-shift rules as player ships.
- Weapon cooldown sweeps now advance smoothly from local time while packet updates correct drift without repeatedly snapping the display.
- Concurrent local clients now use separate active log segments, preventing one client from disabling another client’s structured logging.
- Remote-player, spectator, asteroid lifecycle, and realtime cadence behavior is more stable under load.
- Saved authentication survives restarts, test runs, and temporary provider outages.
- Asteroid spawning remains bounded and clear of active player camera views.

## [0.1.1] - 2026-07-24

- Initial packaged alpha release with Windows and macOS installers and native smoke-tested release archives.

# Local Packaged Alpha Manual Smoke

Run this short presentation check against the exact artifact produced by the automated native release gate.

- Launch the packaged client normally outside the Godot editor.
- Confirm the main menu renders correctly and normal input works.
- Create or select a local profile through the visible UI.
- Start and play a single-player match using ordinary player controls.
- Confirm the visible results screen presents the completed match correctly.
- Quit and relaunch the same package.
- Confirm the visible profile and statistics presentation matches the persisted data.

Record the artifact manifest commit/version and the result of each item before promotion. Secure credentials, native package contents, loopback-only server startup, process cleanup, single-player session start/end, resolved match data, local profile selection, and statistics persistence are already enforced by the automated gate; this checklist covers only presentation and human interaction that would otherwise require brittle UI automation.

# Account And Identity Systems

Parent index: [Platform Planning](./!README.md)

## Purpose

This document plans future account and identity policy that is not yet current implementation.

It owns policy for authenticated accounts, OAuth provider expansion, manual signup and login, email verification, account recovery, provider linking and unlinking, account display identity, display-name moderation, account deletion and deactivation, token/session upgrades, preferences transfer, and development-only auth bypass behavior.

Current implemented account and identity behavior belongs to [Account And Identity Current State](../../domains/platform/account-and-identity-current-state.md). This document should not duplicate the current login, WebSocket auth, admission, or player-data routing flows except where current behavior constrains future planning.

## Overview

Space Rocks has three identity states:

* Guest
* Local Profile
* Authenticated Account

Guest and Local Profile remain local-only identities.

Authenticated Account is the only production identity that may enter online multiplayer.

The future account surface should support:

* Discord OAuth as the current provider.
* Google OAuth as the second provider.
* Manual signup and login.
* Manual signup with email verification.
* Account recovery.
* Provider linking and unlinking.
* Strict account display-name moderation.
* Account deletion and deactivation.
* A stronger future token/session model.
* Development-only auth bypass for local/dev environments.
* Preferences/settings transfer only, with strict trust limits.

The account surface must not support:

* Steam login.
* Guest to Authenticated Account migration.
* Local Profile to Authenticated Account migration.
* Any local-to-multiplayer migration.
* Account merge.
* Automatic sync between Local Profile and Authenticated Account.
* Production auth bypass.

Local Profile is not a Rails/API cache.

Authenticated Account is not a synced Local Profile.

## Current status

Active planning.

The current production identity model is partially implemented and documented in the current-state domain doc. This plan defines the remaining account and identity policy required before the account surface is complete.

Current implementation facts:

* Discord OAuth exists.
* Manual API auth exists at the backend level, but manual login/signup is not fully enabled as a product/client surface.
* Current user tokens are opaque bearer tokens.
* Production multiplayer create/join requires Authenticated Account identity.
* Missing auth verifier returns `auth_unavailable`.
* Guest and Local Profile are local-only identities.

Planned changes from current implementation:

* Add Google OAuth as the second provider.
* Add/enable manual login and signup.
* Require email verification for manual signup.
* Add account recovery.
* Upgrade opaque bearer tokens to JWTs or a better selected token model.
* Add provider linking and unlinking policy.
* Add deletion/deactivation policy.
* Add strict display-name moderation.
* Add build-flagged/env-gated development-only auth bypass.

## Decisions made

### Production admission

* Production multiplayer create/join requires Authenticated Account identity.
* Missing auth verifier returns `auth_unavailable`.
* Live deployed servers must never allow auth bypass.
* Development-only auth bypass is planned, but it must be build-flagged, environment-gated, and unavailable on live deployed servers.

### Providers

* Discord remains the current provider.
* Google is the planned second OAuth provider.
* Steam is explicitly excluded.
* Provider display names are not durable identity.
* Provider display names may seed Space Rocks display names only after moderation/screening.

### Manual auth

* Manual login/signup must be added and enabled.
* Manual signup creates an unverified account.
* Provider signup does not create an unverified account.
* Manual signup requires email verification.
* Unverified manual accounts must not enter online multiplayer.

### Account identity

* `account_id` is canonical authenticated-account identity.
* Rails `user_id` remains internal.
* Account display name is presentation identity.
* Account display name is not durable identity.
* Display names do not need to be globally unique initially.
* Internal identity and writes must use `account_id`, not display name.

### Token/session policy

* Current opaque bearer tokens are temporary.
* Future auth tokens should be upgraded to JWTs or a better selected model.
* The future token/session model must support expiry, revocation, logout invalidation, password-reset invalidation, account-recovery invalidation, account-disabled/deleted invalidation, and service-to-service verification.
* The game-server must continue verifying account identity through the API auth boundary rather than reading Rails auth tables directly.

### Provider linking and unlinking

* One Authenticated Account may have multiple linked provider identities.
* Provider linking requires proof of the current account session and proof of the new provider identity.
* Provider unlinking is planned.
* The final usable login method must not be unlinkable unless account recovery or another login method exists.
* If a provider identity is already linked to an account, sign-in with that provider resolves to the already-linked account.
* If another signed-in account tries to link an already-linked provider identity, linking rejects and routes the user to the already-linked account login or recovery path.
* Provider conflicts must not transfer data.
* Provider conflicts must not merge accounts.

### Account recovery

* Account recovery is planned.
* Password reset is part of account recovery.
* Recovery must support session/token revocation after account recovery or password reset.
* Recovery must account for users who lose provider access.
* Recovery must support manual account flows once manual signup/login is enabled.

### Account deletion and deactivation

* Account deletion/deactivation is planned.
* Disabled accounts cannot authenticate.
* Deleted or pending-deletion accounts cannot authenticate.
* Disabled/deleted accounts cannot enter online multiplayer.
* Disabled/deleted accounts cannot create online-trusted facts.
* Deletion/deactivation must not create migration, transfer, or merge behavior.
* Exact retention, anonymization, erasure timing, audit history, and legal handling are later legal/product decisions.

### Migration, transfer, and merge

* Guest to Authenticated Account migration will never happen.
* Local Profile to Authenticated Account migration will never happen.
* Local-to-multiplayer migration of any kind will never happen.
* Account merge will never happen.
* Provider linking must not imply account merge.
* Account recovery must not imply account merge.
* Provider conflict handling must not imply account merge.
* Automatic sync between Local Profile and Authenticated Account is explicitly excluded.
* The only local/online transfer surface that may ever exist is player preferences/settings.
* Preferences/settings transfer should prefer online to offline export.
* Offline to online preference import, if ever supported, requires screening and trust verification.

Never import or transfer these from local/offline into online-trusted account state:

* currency
* inventory
* unlocks
* progression
* achievements
* leaderboard scores
* ranked results
* trusted match history
* commerce entitlements
* competitive challenge completions
* anti-cheat-sensitive facts

### Trust

* Authenticated Account is required for online-trusted facts.
* Guest and Local Profile facts are not online-trusted.
* Local facts do not become trusted because they are copied, imported, linked, or associated with an account.
* Preferences/settings are the only possible local/online transfer category, and offline to online transfer requires screening/trust verification.
* Trust-sensitive imported data policy belongs to Anti-Cheat And Trust Policy, but this document owns the identity boundary.

### Display-name moderation

Display-name moderation should be added immediately and strictly enforced.

Launch ideal:

1. banned-word list enforcement
2. secondary classifier analysis
3. report button fallback
4. LLM review queue
5. human review or appeal when ambiguous

Moderation applies to account display identity and any future public-facing player naming surface.

Provider display names must not bypass moderation.

## Open decisions

Only implementation-shape decisions remain:

* Exact account display-name validation rules.
* Exact display-name moderation thresholds.
* Exact secondary classifier choice.
* Exact LLM review queue shape.
* Exact human review and appeal flow.
* Exact JWT-or-better token model.
* Exact token expiry and refresh behavior.
* Exact session revocation rules for password reset, recovery, account disable, and account deletion.
* Exact manual signup UX.
* Exact email verification UX.
* Exact resend-verification behavior.
* Exact account recovery UX.
* Exact account deletion/deactivation UX.
* Exact retention, anonymization, erasure timing, audit history, and legal handling.
* Exact preferences/settings export shape.
* Exact offline to online preferences/settings screening rules if import is ever supported.
* Exact build flag and environment variable names for development-only auth bypass.

The major policy questions are decided.

These should not remain open:

* Whether Google is the second provider.
* Whether Steam is supported.
* Whether manual signup/login exists.
* Whether manual signup requires email verification.
* Whether account recovery is planned.
* Whether Local Profile can migrate into Authenticated Account.
* Whether Guest can migrate into Authenticated Account.
* Whether account merge exists.
* Whether production auth bypass is allowed.

## Expected ownership

Account and identity planning owns:

* account identity policy
* provider identity policy
* manual auth policy
* email verification policy
* account recovery policy
* provider linking and unlinking policy
* account lifecycle states
* account deletion/deactivation identity effects
* display identity policy
* identity transition exclusions
* account merge exclusion
* local/online preference transfer boundary
* development-only auth bypass policy

API Product Surface owns:

* API/account product endpoints
* account settings surfaces
* manual signup/login endpoints and UI-facing backend behavior
* email verification endpoints
* password reset endpoints
* account recovery endpoints
* account deletion/deactivation endpoints
* account-facing deletion/recovery UX support
* legal/product retention and anonymization implementation once decided

Player Data And Persistence owns:

* player-data contracts
* storage routes
* schema parity
* persistence behavior
* preferences/settings persistence shape
* any export/import storage mechanics for preferences/settings

Multiplayer Session And Lifecycle owns:

* WebSocket sessions
* room participation lifecycle
* multiplayer create/join admission execution
* reconnect behavior
* active room/session behavior

Social And Community Systems owns:

* player relationships
* profile visibility
* public player-facing display surfaces
* friends
* blocks
* mutes
* invites
* report-button integration with social/public identity surfaces

Anti-Cheat And Trust Policy owns:

* online trust policy
* imported-fact trust limits
* leaderboard eligibility support
* abuse-resistance policy
* anti-farming policy
* debug/devtools exclusion from online-trusted facts

Devtools planning owns:

* development-only auth bypass controls
* build/runtime gates
* local/dev-only behavior
* ensuring dev bypass cannot be enabled in production deployments

## Implementation sequence

1. Define the account display identity policy and validation rules.
2. Add strict display-name moderation policy before public display-name editing expands.
3. Add manual signup/login product flow.
4. Add email verification for manual signup.
5. Block unverified manual accounts from online multiplayer.
6. Plan and implement password reset and account recovery.
7. Upgrade opaque bearer tokens to JWTs or a better selected token/session model.
8. Define token expiry, refresh, logout, revocation, password-reset invalidation, recovery invalidation, and account-disabled/deleted invalidation behavior.
9. Add Google OAuth as the second provider.
10. Define provider linking with current-session proof and new-provider proof.
11. Define provider conflict handling so already-linked providers reject linking and route to the existing linked account login/recovery path.
12. Define provider unlinking rules after recovery and alternate-login requirements are in place.
13. Add account deletion/deactivation identity behavior.
14. Add legal/product-specific retention and anonymization details when available.
15. Define preferences/settings export behavior, preferring online to offline.
16. If offline to online preferences/settings import is ever supported, define screening and trust verification first.
17. Define development-only auth bypass as build-flagged and environment-gated.
18. Ensure development-only auth bypass cannot exist on live deployed servers and cannot create online-trusted facts.
19. Keep account merge and local-to-online migration explicitly unsupported.

## Related docs

* [Account And Identity Current State](../../domains/platform/account-and-identity-current-state.md)
* [Player Data And Persistence](stubs/player-data-and-persistence.md)
* [Multiplayer Session And Lifecycle](stubs/multiplayer-session-and-lifecycle.md)
* [API Product Surface](stubs/api-product-surface.md)
* [Social And Community Systems](stubs/social-and-community-systems.md)
* [Anti-Cheat And Trust Policy](stubs/anti-cheat-and-trust-policy.md)

## Notes

This document is the planning home for future account and identity policy. Current implemented auth/session/admission behavior belongs in the domain current-state doc.

Manual signup creates an unverified account. Provider signup does not.

A possible development-only auth bypass is not “unauthorized access.” It is a controlled local/dev capability, must be build-flagged and environment-gated, and must not be available on live deployed servers.

Local Profile data and Guest data must never become online multiplayer identity, online-trusted progression, or account migration material.

The only possible local/online transfer category is preferences/settings, and the preferred direction is online to offline export. Offline to online import is optional future work and requires strict screening/trust verification before it can exist.

Account merge is permanently excluded. Provider linking, recovery, deletion, and conflict handling must never imply account merge.

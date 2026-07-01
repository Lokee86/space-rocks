# Observability Logging And Diagnostics

Parent index: [Technical Planning](./!INDEX.md)

## Purpose

This doc plans product-wide observability, logging, diagnostics, and log aggregation for Space Rocks.

It defines what each service should log, at what level, with which shared fields, how logs are grouped across services, how sensitive data and noise are controlled, how local and hosted diagnostics are gathered, and how aggregated logs support bug reports, operational readiness, admin review, audit-grade records, and future incident response.

## Overview

This doc keeps product logging, diagnostics, aggregation, and redaction policy aligned so failures stay diagnosable without turning logs into gameplay or analytics noise.

## Current status

Active planning.

## Ownership Boundary

This doc owns planning for:

* product log standards,
* log levels,
* stable event names,
* canonical log fields,
* product log aggregation,
* error aggregation as a log-aggregation use case,
* diagnostic bundles and copy diagnostics,
* audit trigger fields and audit-grade aggregated records,
* observability SSoT consumption,
* redaction and forbidden fields,
* local single-player diagnostic aggregation,
* launch-shaped observability expectations.

Service docs own exact logging implementation. Domain docs own the meaning of domain events, such as enforcement actions, reward grants, purchases, integrity flags, and admin corrections.

## Does Not Belong

* Exact logging implementation code.
* Gameplay behavior.
* Auth behavior.
* Persistence behavior.
* Reward formulas.
* Enforcement policy.
* Exact log aggregation backend choice.
* Exact incident platform implementation.
* Player analytics or marketing analytics.
* Current implementation authority.

## Existing Inputs

Space Rocks already has partial logging foundations.

The game-server has structured/category logging through a Go `slog` wrapper. Current game-server categories include:

* `server`
* `network`
* `rooms`
* `game`

The client has a logging helper with levels, categories, and helper methods. Current client categories include:

* `shell`
* `lobby`
* `network`
* `game`
* `world_sync`
* `hud`
* `input`
* `packets`

Network packet observability is planned separately in [Network Observability And Packet Budget](network-observability-and-packet-budget.md). Operational readiness already depends on copy diagnostics, bug reports, telemetry/logging readiness, health checks, and future incident-platform support.

## Product Observability Model

Space Rocks should use one product-wide observability model across:

* Godot client,
* bundled local single-player server,
* hosted game-server,
* API-server,
* player-data service,
* website,
* devtools/admin tools,
* future workers or jobs,
* future product log aggregation service.

The model is:

```text
product services emit structured logs/events
-> product log aggregator collects and correlates them
-> errors, diagnostics, incidents, bug reports, and audit-grade records are consumers or tiered outputs of the aggregated event stream
```

Logging is observational. Logs, metrics, telemetry, diagnostics, and aggregation must not change gameplay state, persistence behavior, auth behavior, packet routing, reward behavior, or enforcement behavior by themselves.

## Product Log Aggregation Service

Space Rocks should plan a future product log aggregation service.

The log aggregation service should collect, normalize, group, search, and retain product events from all services.

It should support:

* cross-service event grouping,
* error aggregation,
* bug report attachment,
* diagnostic bundle generation,
* pending/retry investigation,
* admin and support review,
* operational incident diagnosis,
* audit-grade persistence when triggered,
* future dashboards and alerting.

Error aggregation is a use case of log aggregation. A separate error aggregation service is not required.

Audit-grade records are also handled by the log aggregation system when triggered by audit-worthy events. A separate audit aggregation service is not required.

Aggregator failure must not break gameplay. It should degrade diagnostics and reporting. For audit-grade events, failed submission should use pending/retry where practical.

## Local Single-Player Diagnostics

Local packaged single-player must participate in observability.

Local packaged single-player should aggregate client and bundled-server diagnostics locally. It must not require hosted services for diagnostic collection.

Expected behavior:

| Situation                 | Behavior                                                                    |
| ------------------------- | --------------------------------------------------------------------------- |
| Offline single-player     | Aggregate locally; expose copy diagnostics or saved diagnostic report.      |
| Online services available | May offer upload or auto-upload depending beta/production reporting policy. |
| Upload fails              | Preserve local diagnostics and avoid creating a second failure loop.        |

The bundled local server may own the local aggregation endpoint or module because it already exists in single-player and can collect server-side context.

## Bug Reports And Copy Diagnostics

Bug reporting is the user/tester-facing surface of diagnostics.

Bug reports should attach or reference an aggregated diagnostic bundle when available. Copy diagnostics is the fallback when automatic upload is unavailable, disabled, or fails.

Copy diagnostics should be safe to paste unless explicitly marked otherwise.

Bug-report and diagnostic bundles should include enough service, build, session, environment, and failure context to debug the issue while avoiding secrets, tokens, raw private profile data, payment data, and unnecessary raw payloads.

## Observability SSoT

Space Rocks should maintain a product observability source of truth.

The observability SSoT should define:

* log/event schema,
* schema version,
* canonical log levels,
* canonical event names,
* canonical field names,
* audit trigger fields,
* audit types,
* diagnostic bundle shape,
* redaction and forbidden-field rules,
* correlation ID rules,
* local/dev-only event rules,
* aggregation eligibility,
* retention tier metadata where practical.

Possible future shape:

```text
shared/contracts/observability/
  events.toml
  fields.toml
  audit_types.toml
  diagnostic_bundle.toml
  retention_tiers.toml
```

Exact file names and format are open.

## Observability SSoT Consumption

The observability SSoT should be machine-consumed.

Developers and agents should not manually copy shared event names, field names, audit types, or diagnostic schema rules from prose into service code.

The SSoT should generate:

| Target              | Generated Output                                                    |
| ------------------- | ------------------------------------------------------------------- |
| Go game-server      | Event constants, field constants, audit type constants, validators. |
| GDScript client     | Event constants, field constants, diagnostic bundle constants.      |
| Ruby API-server     | Event constants, field constants, audit type constants.             |
| Player-data service | Event, field, audit, and storage-related constants.                 |
| Log aggregator      | Validation schema, grouping fields, redaction rules, audit rules.   |
| Documentation       | Human-readable event and field catalog.                             |
| Verification gates  | Drift checks for generated observability outputs.                   |

This should be another data-pipeline consumption point. `data-sync` or an equivalent generation path should eventually produce observability constants, validators, and reference docs from the SSoT.

Release-shaped builds should fail if generated observability constants, validators, schemas, or documentation are stale relative to the SSoT.

## Stable Event Names

Product logs should use stable event names in addition to human-readable messages.

Stable event names allow aggregation, search, grouping, dashboards, alerting, and audit promotion without parsing prose.

Examples:

```text
service_started
service_startup_failed
configuration_invalid
dependency_unavailable
auth_admission_failed
compatibility_blocked
client_connected
client_disconnected
room_created
match_started
match_ended
match_result_write_failed
match_result_retry_scheduled
player_data_unavailable
local_profile_migration_failed
packet_size_warning
runtime_slow_tick
environment_degraded
maintenance_started
admin_action_applied
bug_report_created
```

Human-readable messages may still exist, but product tooling should key off stable event names.

## Canonical Fields

Services may implement logging differently, but shared concepts should use canonical field names.

Core fields:

```text
timestamp
level
event
service
environment
build_version
schema_version
category
message
request_id
session_id
room_id
match_id
player_id
account_id
route
packet_type
error_code
failure_mode
duration_ms
degraded_state
idempotency_key
diagnostic_report_id
audit_event_id
```

Not every event needs every field. The requirement is that shared concepts use the same names across services.

## Correlation IDs

Any flow that crosses a service boundary should create or preserve a correlation ID.

Examples:

| Flow                     | Correlation Fields                         |
| ------------------------ | ------------------------------------------ |
| API request              | `request_id`                               |
| Gameplay session         | `session_id`                               |
| Room flow                | `room_id`                                  |
| Match result write       | `match_id`, `player_id`, `idempotency_key` |
| Bug report               | `diagnostic_report_id`                     |
| Audit-worthy action      | `audit_event_id`                           |
| Production investigation | `environment`, `build_version`, `service`  |

This is what lets separate client, game-server, API-server, and player-data logs become one diagnosable failure chain.

## Log Levels

Log levels should be used by operational meaning.

| Level          | Use                                                                                |
| -------------- | ---------------------------------------------------------------------------------- |
| Debug          | Temporary or gated subsystem investigation.                                        |
| Info           | Important normal lifecycle events; low volume.                                     |
| Warn           | Recoverable issue, suspicious condition, retry, degradation, or threshold warning. |
| Error          | Failed operation requiring attention, but service can continue.                    |
| Critical/Fatal | Service cannot start, cannot operate safely, or must block admission/shutdown.     |

Debug logs are temporary by default. They may remain only when category-gated, off by default, low-volume, safe, and useful for recurring diagnosis.

## Logging Policy

Permanent logs should describe important events, not every operation.

Log:

* state transitions,
* failures,
* degraded conditions,
* service boundary crossings,
* state-changing admin/devtools actions,
* threshold crossings,
* audit-worthy actions,
* diagnostic bundle and bug-report events.

Do not log by default:

* every function call,
* every tick,
* every frame,
* every entity update,
* every input packet,
* every successful lane packet,
* every collision candidate,
* every normal position update,
* full packet payloads,
* raw profile blobs,
* duplicate same-failure logs without correlation.

High-frequency samples should usually be metrics or devtools telemetry, not normal logs. Logs should capture threshold crossings, summaries, lifecycle events, and failures.

## Required Logging And Observability Points

### Service Lifecycle

Every service should log:

| Event                                        | Level          |
| -------------------------------------------- | -------------- |
| service starting                             | Info           |
| service started/ready                        | Info           |
| service stopping                             | Info           |
| service stopped                              | Info           |
| startup failed                               | Error/Critical |
| required config missing or invalid           | Critical       |
| unsafe production config detected            | Critical       |
| dependency initialization failed             | Error/Critical |
| service entered degraded state               | Warn           |
| service entered admission-blocked state      | Warn           |
| service entered offline or maintenance state | Warn/Info      |
| service recovered from degraded state        | Info           |

### Configuration And Environment

Log:

| Event                                          | Level                           |
| ---------------------------------------------- | ------------------------------- |
| environment selected                           | Info                            |
| build version loaded                           | Info                            |
| protocol/API/data contract version loaded      | Info                            |
| feature flag or devtools mode loaded           | Info/Warn depending environment |
| production client devtools capability detected | Error/Critical                  |
| production server admin/devtools path enabled  | Warn with gate context          |
| invalid environment variable                   | Error/Critical                  |
| missing secret or credential                   | Critical                        |
| forbidden SQLite use in multiplayer build      | Critical                        |
| local packaged server using local-only bind    | Info                            |
| local packaged server failed local-only bind   | Error/Critical                  |

Never log raw secret values.

### Client Startup And Shell

Log:

| Event                                     | Level      |
| ----------------------------------------- | ---------- |
| client startup                            | Info       |
| client build/environment selected         | Info       |
| main scene or session boot failed         | Error      |
| config load failed                        | Error      |
| local settings/profile config read failed | Warn/Error |
| packaged runtime mismatch                 | Error      |
| local bundled server launch requested     | Info       |
| local bundled server launch failed        | Error      |
| local bundled server connected            | Info       |
| local bundled server exited unexpectedly  | Error      |
| copy diagnostics generated                | Info       |
| bug report submitted or uploaded          | Info       |
| bug report upload failed                  | Warn       |

Noisy UI branch tracing should stay debug-only or temporary.

### Client Network And Session Connection

Log:

| Event                                    | Level      |
| ---------------------------------------- | ---------- |
| connection attempt started               | Info/Debug |
| connection succeeded                     | Info       |
| connection failed                        | Warn/Error |
| connection lost unexpectedly             | Warn/Error |
| expected disconnect                      | Debug      |
| reconnect attempt scheduled              | Info/Warn  |
| reconnect succeeded                      | Info       |
| reconnect failed                         | Warn/Error |
| packet decode failed                     | Warn       |
| packet route unknown                     | Warn       |
| compatibility block received             | Warn       |
| auth/admission block received            | Warn       |
| server unavailable received              | Warn       |
| player-data unavailable warning received | Warn       |

Do not log every inbound or outbound packet by default.

### Auth, Account, And Admission

Log:

| Event                                     | Level                             |
| ----------------------------------------- | --------------------------------- |
| auth flow started                         | Info                              |
| auth callback received                    | Info                              |
| auth succeeded                            | Info                              |
| auth failed                               | Warn/Error                        |
| auth provider unavailable                 | Error                             |
| account identity loaded                   | Info/Debug                        |
| admission check started                   | Debug                             |
| admission passed                          | Debug/Info in hosted environments |
| admission failed                          | Warn                              |
| incompatible client blocked               | Warn                              |
| banned/suspended account blocked          | Warn                              |
| debug-tainted or bot/TAS account admitted | Info/Warn                         |
| token verification failed                 | Warn/Error                        |

Never log tokens, OAuth codes, raw auth headers, or secrets.

### API-Server HTTP Requests

Log:

| Event                           | Level                          |
| ------------------------------- | ------------------------------ |
| request started                 | Debug or structured access log |
| request completed               | Debug or structured access log |
| request failed                  | Warn/Error                     |
| validation failed               | Warn                           |
| unauthorized or forbidden       | Warn                           |
| rate limit triggered            | Warn                           |
| upstream dependency unavailable | Error                          |
| player-data request failed      | Error                          |
| idempotency conflict            | Warn                           |
| request timeout                 | Warn/Error                     |
| route not found                 | Debug/Warn depending volume    |

Production request logging should use structured access logs with sampling/rate controls where needed, not hand-written info spam.

### Player-Data Service

Log:

| Event                                   | Level          |
| --------------------------------------- | -------------- |
| store initialized                       | Info           |
| store initialization failed             | Critical       |
| storage backend selected                | Info           |
| invalid storage backend for environment | Critical       |
| profile read failed                     | Error          |
| profile write failed                    | Error          |
| local profile migration started         | Info           |
| local profile backup created            | Info           |
| local profile backup failed             | Error          |
| local profile migration succeeded       | Info           |
| local profile migration failed          | Error          |
| match result write started              | Debug          |
| match result write succeeded            | Info/Debug     |
| match result write failed               | Error          |
| result marked pending                   | Warn           |
| result retry scheduled                  | Warn           |
| result retry succeeded                  | Info           |
| duplicate or idempotent result ignored  | Info/Warn      |
| reward grant failed                     | Error          |
| currency/item/stat update failed        | Error          |
| database unavailable                    | Error/Critical |
| database migration failed               | Critical       |

Do not log raw private profile blobs.

### Game-Server Networking

Log:

| Event                                       | Level                          |
| ------------------------------------------- | ------------------------------ |
| WebSocket upgrade failed                    | Warn                           |
| WebSocket connected                         | Info/Debug                     |
| WebSocket disconnected expected             | Debug                          |
| WebSocket disconnected unexpected           | Warn                           |
| read failed expected close                  | Debug                          |
| read failed unexpected                      | Warn                           |
| write failed                                | Warn/Error                     |
| packet envelope decode failed               | Warn                           |
| packet decode failed                        | Warn                           |
| unknown packet type                         | Warn                           |
| packet route failed                         | Warn/Error                     |
| outbound encode failed                      | Error                          |
| gameplay packet too large                   | Warn/Error depending threshold |
| gameplay write too slow                     | Warn                           |
| devtools packet rejected                    | Warn/Error                     |
| production client devtools packet attempted | Error                          |

Do not log successful lane packet writes by default.

### Game-Server Room Lifecycle

Log:

| Event                            | Level                        |
| -------------------------------- | ---------------------------- |
| room created                     | Info                         |
| room creation failed             | Error                        |
| room join succeeded              | Debug/Info                   |
| room join failed                 | Warn                         |
| room full                        | Debug/Warn depending context |
| room title/name rejected         | Warn                         |
| room invalidated by moderation   | Warn                         |
| room ready state changed         | Debug                        |
| room starting countdown started  | Info                         |
| room starting countdown canceled | Info                         |
| room match started               | Info                         |
| room match ended                 | Info                         |
| room cleanup scheduled           | Debug                        |
| room cleanup skipped             | Debug/Warn                   |
| room cleaned up                  | Info/Debug                   |
| owner transferred                | Info                         |
| player kicked                    | Info/Warn                    |
| player banned from room          | Warn                         |
| no-action timeout removed player | Info                         |

Room logs should focus on lifecycle, not every small lobby state mutation.

### Game Simulation And Match Lifecycle

Log:

| Event                                  | Level                      |
| -------------------------------------- | -------------------------- |
| match started                          | Info                       |
| match ended                            | Info                       |
| game over detected                     | Info                       |
| player added to simulation             | Debug                      |
| player removed from simulation         | Debug                      |
| player died                            | Debug/Info                 |
| player game over                       | Info                       |
| respawn requested                      | Debug                      |
| respawn blocked                        | Warn/Debug depending cause |
| score awarded                          | Debug                      |
| pickup collected                       | Debug, not permanent Info  |
| collision shape missing                | Warn                       |
| invalid entity state detected          | Warn/Error                 |
| simulation panic/recover               | Error/Critical             |
| slow tick threshold crossed            | Warn                       |
| entity count warning threshold crossed | Warn                       |

Do not log every tick, movement update, collision candidate, asteroid spawn candidate, or bullet update by default.

### Match Results, Rewards, And Progression

Log:

| Event                                          | Level                        |
| ---------------------------------------------- | ---------------------------- |
| match result report started                    | Info                         |
| match result report skipped                    | Warn/Info depending expected |
| match result report failed                     | Error                        |
| match result report succeeded                  | Info                         |
| reward calculation failed                      | Error                        |
| reward grant failed                            | Error                        |
| reward grant succeeded                         | Info/Debug                   |
| result marked pending                          | Warn                         |
| retry scheduled                                | Warn                         |
| retry exhausted                                | Error                        |
| idempotent duplicate suppressed                | Info/Warn                    |
| leaderboard update failed                      | Error                        |
| leaderboard update skipped due to bot/TAS flag | Info                         |
| bot/TAS run flagged                            | Info/Warn                    |
| abuse/integrity flag attached to result        | Warn                         |

Reward-bearing flows need stronger logs because they affect player trust, support, eligibility, and recovery.

### Local Profiles And Migrations

Log:

| Event                      | Level      |
| -------------------------- | ---------- |
| local profile store opened | Info       |
| local profile store failed | Error      |
| profile create failed      | Error      |
| profile load failed        | Error      |
| profile save failed        | Error      |
| profile delete failed      | Warn/Error |
| migration needed           | Info       |
| migration backup started   | Info       |
| migration backup succeeded | Info       |
| migration backup failed    | Error      |
| migration started          | Info       |
| migration succeeded        | Info       |
| migration failed           | Error      |
| repair/export offered      | Info       |
| repair/export failed       | Error      |

No raw profile blobs.

### Devtools, Admin Tools, And State-Changing Actions

Log any state-changing devtools/admin action.

| Event                                   | Level          |
| --------------------------------------- | -------------- |
| devtools enabled in allowed environment | Info           |
| devtools disabled by build policy       | Info           |
| devtools command received               | Debug          |
| devtools command applied to gameplay    | Info           |
| devtools command rejected               | Warn           |
| admin action requested                  | Info           |
| admin action applied                    | Info/Warn      |
| admin action failed                     | Error          |
| account flag changed                    | Warn           |
| room force-closed                       | Warn           |
| reward/currency correction applied      | Warn           |
| moderation action applied               | Warn           |
| ban/kick/mute/rename action applied     | Warn           |
| audit-grade record creation failed      | Error/Critical |

These should include audit trigger fields when they affect accounts, rewards, moderation, rankings, rooms, or production state.

### Abuse, Integrity, And Trust Signals

Log:

| Event                           | Level      |
| ------------------------------- | ---------- |
| suspicious packet rejected      | Warn       |
| impossible state/input detected | Warn       |
| debug-tainted run marked        | Info/Warn  |
| bot/TAS flag applied            | Info/Warn  |
| integrity check failed          | Warn/Error |
| account trust status changed    | Warn       |
| moderation classifier failed    | Error      |
| moderation action applied       | Warn       |
| appeal submitted                | Info       |
| appeal rate-limited             | Warn       |
| enforcement action failed       | Error      |

Do not log sensitive moderation text unnecessarily. Evidence storage and review policy belong in security/admin docs.

### Website And Web Presence

Log:

| Event                        | Level      |
| ---------------------------- | ---------- |
| website service started      | Info       |
| website health check failed  | Error      |
| page/render failure          | Error      |
| direct purchase flow started | Info       |
| direct purchase flow failed  | Error      |
| signup/follow action failed  | Error      |
| Steam key claim started      | Info       |
| Steam key claim failed       | Error      |
| API dependency unavailable   | Error      |
| invalid webhook or signature | Warn/Error |
| payment provider unavailable | Error      |

Website analytics are separate. Do not mix marketing analytics into operational logs by accident.

### Commerce, Purchase, Keys, And Entitlements

When these exist, log carefully.

| Event                            | Level |
| -------------------------------- | ----- |
| purchase initiated               | Info  |
| purchase authorized              | Info  |
| purchase failed                  | Error |
| entitlement grant started        | Info  |
| entitlement grant succeeded      | Info  |
| entitlement grant failed         | Error |
| Steam key claim started          | Info  |
| Steam key claim succeeded        | Info  |
| Steam key claim failed           | Error |
| duplicate entitlement suppressed | Warn  |
| refund/reversal received         | Warn  |
| payment webhook invalid          | Error |
| payment provider unavailable     | Error |

Never log payment card data, payment secrets, or raw provider payloads unless explicitly redacted and gated.

### Product Log Aggregation Service

The aggregator should log its own lifecycle and failures.

| Event                               | Level          |
| ----------------------------------- | -------------- |
| aggregation service started         | Info           |
| aggregation service unavailable     | Error          |
| event accepted                      | Debug          |
| event rejected invalid schema       | Warn           |
| event rejected unsafe fields        | Warn/Error     |
| event grouped into diagnostic chain | Debug/Info     |
| audit-grade record created          | Info           |
| audit-grade record creation failed  | Error/Critical |
| aggregation storage failed          | Error/Critical |
| duplicate event suppressed          | Debug          |
| retention policy applied            | Info           |
| bug-report bundle created           | Info           |
| bug-report upload failed            | Warn/Error     |

Accepted events should not be Info by default or the aggregator will become noisy.

### Bug Reports And Copy Diagnostics

Log:

| Event                                   | Level          |
| --------------------------------------- | -------------- |
| diagnostic bundle created               | Info           |
| copy diagnostics generated              | Info           |
| report upload started                   | Info           |
| report upload succeeded                 | Info           |
| report upload failed                    | Warn           |
| report rejected for unsafe fields       | Warn           |
| report attached to aggregated log group | Info           |
| report redaction failed                 | Error/Critical |

Copy diagnostics must be safe to paste unless explicitly marked otherwise.

### Health, Readiness, And Operations

Log:

| Event                                      | Level      |
| ------------------------------------------ | ---------- |
| liveness check failed internally           | Error      |
| readiness check failed                     | Warn/Error |
| dependency health failed                   | Error      |
| admission health blocked                   | Warn       |
| maintenance mode entered                   | Warn/Info  |
| maintenance mode exited                    | Info       |
| environment degraded                       | Warn       |
| environment recovered                      | Info       |
| telemetry/logging unavailable pre-launch   | Error      |
| telemetry/logging unavailable after launch | Warn/Error |
| incident opened                            | Warn/Error |
| incident resolved                          | Info       |

Do not log every successful health check at Info.

### Runtime Performance And Scale

Log only thresholds and summaries, not every sample.

| Event                                  | Level                          |
| -------------------------------------- | ------------------------------ |
| slow tick threshold crossed            | Warn                           |
| frame pressure threshold crossed       | Warn                           |
| packet size warning threshold crossed  | Warn                           |
| packet danger threshold crossed        | Error or release-blocking gate |
| entity count warning threshold crossed | Warn                           |
| memory growth warning                  | Warn                           |
| load scenario started                  | Info                           |
| load scenario completed                | Info                           |
| load scenario failed                   | Error                          |
| soak run detected degradation          | Warn/Error                     |

Raw high-frequency measurements should be metrics/devtools, not normal logs.

## Audit-Grade Records

The log aggregator should create audit-grade records when triggered by audit-worthy structured events.

Audit-worthy events should include:

```text
audit_required: true
audit_type: <type>
```

Recommended audit fields:

```text
audit_required
audit_type
actor_id
actor_type
target_type
target_id
action
reason_code
case_id
transaction_id
match_id
result_id
account_id
```

`audit_required` means the aggregator must validate, preserve, and promote the event into an audit-grade record. It is clearer than `auditable`.

Aggregator behavior:

```text
audit_required=true
-> validate required audit fields
-> redact or reject unsafe fields
-> group with related logs
-> persist audit-grade aggregated record
-> attach or return audit_event_id where applicable
```

Domains decide which actions are audit-worthy and what domain-specific payload is required. The aggregator owns shared collection, validation, correlation, retention tiering, and audit-grade persistence.

## Audit Boundary

Observability owns:

* audit trigger fields,
* audit reference fields,
* audit-grade aggregation behavior,
* shared event schema,
* redaction rules,
* correlation with logs.

Domain docs own:

* which actions are audit-worthy,
* what domain-specific payload is required,
* policy consequences,
* appeals,
* reversals,
* restoration rules.

Audit-grade events may include:

* moderation/enforcement actions,
* account restrictions, bans, and suspensions,
* appeals and appeal decisions,
* bot/TAS flags,
* integrity eligibility decisions,
* reward/currency/item grants,
* reward/currency/item corrections,
* leaderboard removals and restorations,
* purchases, refunds, and entitlements,
* admin overrides,
* room moderation actions,
* competitive result adjudication.

## Retention And Durability Tiers

The aggregator should support different retention and durability tiers.

| Tier              | Use                                                            |
| ----------------- | -------------------------------------------------------------- |
| Ephemeral/Dev     | Local debugging.                                               |
| Operational       | Production diagnosis.                                          |
| Diagnostic Report | Bug report and copy diagnostics.                               |
| Audit-Grade       | Enforcement, economy, admin, disputes, purchases, eligibility. |

Exact retention durations are open.

## Privacy And Redaction

Logs, diagnostics, and bug reports must not include:

* secrets,
* bearer tokens,
* OAuth codes,
* client secrets,
* raw auth headers,
* raw private profile data,
* payment data,
* unnecessary raw packet dumps.

The aggregator should reject or redact unsafe events where practical.

Debug-only payload dumps must be explicitly gated and should not be part of normal release-shaped logging.

## Noise Control

Product logging should avoid flooding itself into uselessness.

Required controls:

* no per-tick logs by default,
* no per-frame logs by default,
* no per-function-call logs by default,
* no per-entity logs by default,
* no full packet dumps by default,
* rate-limit repeated warnings,
* summarize repeated failures where practical,
* support category-level log controls,
* keep debug logs disabled by default in production,
* remove or promote temporary debug logs after investigation.

## Metrics, Telemetry, And Logs

Metrics and telemetry are not the same as logs.

| Signal                                | Preferred Surface          |
| ------------------------------------- | -------------------------- |
| every tick duration                   | metrics/devtools           |
| slow tick threshold crossed           | log                        |
| every frame-time sample               | metrics/devtools           |
| frame pressure threshold crossed      | log                        |
| every packet size                     | metrics/devtools           |
| packet-size warning threshold crossed | log                        |
| every successful lane packet         | metrics/devtools if needed |
| failed packet encode/write            | log                        |

Logs should capture meaningful events, failures, summaries, and thresholds. Metrics and telemetry should carry high-frequency numeric pressure.

## Launch-Shaped Expectations

| Stage                  | Observability Expectation                                                                                                          |
| ---------------------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| Local Development      | Local logs; shared fields where practical.                                                                                         |
| Local Packaged Beta    | Client plus bundled-server diagnostics; copy diagnostics.                                                                          |
| Dev-Hosted Multiplayer | Cross-service logs manually reconcilable by shared IDs.                                                                            |
| Hosted Staging         | Central aggregation should exist or be scaffolded enough to validate grouping.                                                     |
| Hosted Production      | Aggregated logs support incident diagnosis, bug reports, admin review, audit-grade records, recovery, and release-readiness gates. |

## Verification Expectations

Release-shaped builds should verify that:

* generated observability constants are current,
* generated aggregator validation schema is current,
* generated observability docs are current,
* audit-required events include required audit fields,
* forbidden fields are rejected or redacted,
* copy diagnostics excludes unsafe fields,
* important cross-service flows preserve correlation IDs,
* log aggregation failure does not break gameplay,
* production client devtools logging capability follows build policy.

## Implementation sequence

1. Keep the existing structured logging helpers and category sets aligned with the shared observability model.
2. Define the product observability source of truth and its generated consumers.
3. Make local packaged single-player participate in local diagnostic aggregation.
4. Keep bug reports and copy diagnostics attached to the aggregated event stream when possible.
5. Preserve redaction, audit-trigger, and retention rules before broadening incident tooling.
6. Leave service-specific logging implementation in the owning service docs.

## Related docs

* [Planning](../../!INDEX.md)
* [Development Roadmap](../../development-roadmap.md)
* [Network Observability And Packet Budget](network-observability-and-packet-budget.md)
* [Operational Readiness And Failure Modes](operational-readiness-and-failure-modes.md)
* [Verification And Quality Gates](verification-and-quality-gates.md)
* [Runtime Performance And Scale Budget](runtime-performance-and-scale-budget.md)
* [Compatibility Versioning And Migrations](compatibility-versioning-and-migrations.md)
* [Build Release And Environment Matrix](build-release-and-environment-matrix.md)
* [Devtools And Telemetry](../../devtools/devtools-and-telemetry.md)
* [Data Sync And SSoT Pipeline](../../../data/data-sync-and-ssot-pipeline.md)
* [Game Server Logging And Diagnostics](../../../services/game-server/observability/logging-and-diagnostics.md)
* [Client Logging](../../../services/client/client-logging.md)
* [Abuse And Enforcement Admin](../platform/security-and-admin/abuse-and-enforcement-admin.md)
* [Game Integrity Policy](../platform/security-and-admin/game-integrity-policy.md)

## Open decisions

* What exact files define the observability SSoT?
* Which generator owns observability constants and validators?
* Which events are required before local packaged beta?
* Which events are required before hosted production?
* What backend stores aggregated logs?
* What backend stores audit-grade aggregated records?
* What retention durations apply to each tier?
* Which production logs are sampled or rate-limited?
* Which single-player diagnostic uploads require explicit consent?
* Which audit types are required at launch versus later?
* Does the aggregator return `audit_event_id` synchronously or attach it asynchronously?
* What admin/incident dashboard shape comes first?
* What exact bug-report upload policy applies to beta and production builds?

## Notes

Preserve the detailed logging inventory and shared-field catalog; this doc should stay focused on policy, aggregation, diagnostics, and redaction.

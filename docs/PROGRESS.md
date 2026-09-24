# Commercial WCS Progress

Updated: 2026-09-24

## Repository recovery status

The GitHub repository was recreated on 2026-09-24. Acceptance is based only on code and tests present in the recreated repository; historical implementation claims are context, not evidence.

## Verified on main

PR #2 restored the first control safety kernel with CI PASS:

- atomic all-or-none resource reservation;
- deterministic wait-for/deadlock detection;
- trusted device-position confirmation as the only physical-resource release path;
- Go test/vet plus frontend build CI.

## Current PR: domain + lifecycle recovery

`feature/rebuild-domain-lifecycle` restores a typed Go production domain boundary for Device, Task and Alarm and begins safe lifecycle recovery:

- explicit device AUTO/MANUAL/MAINTENANCE modes and trusted-position state;
- task deadlines and bounded retry policy;
- UNKNOWN MOVE transitions to UNKNOWN and is never eligible for blind retry;
- running cancellation is a request, not an immediate physical-state rewrite;
- cancellation/UNKNOWN recovery cannot enter compensation until a trusted non-empty device position is available;
- regression tests cover deadline expiry, retry budget, UNKNOWN MOVE no-replay, and trusted-position compensation gates.

This PR intentionally does not claim failover, compensation completion, segmented MOVE, adapter/protocol support, persistence, or hardware SAT yet.

## Commercial benchmark delta

Fresh public benchmark review on 2026-09-24 confirms the priority. GALAXIS RCS 3.0 combines planning, simulation, virtual commissioning, control, scheduling and O&M with 2D/3D digital twin and space-time coordination. BlueSword IMHS-WCS/3D-SCADA combines heterogeneous equipment control, material-position visibility, fault localization and virtual simulation. Damon WCS reports millisecond scheduling, heterogeneous equipment integration and large project deployment, while its newer systems combine shuttles and AMRs with swarm/path scheduling. Quicktron exposes WES/LES/RCS integration with upstream WMS/ERP/MES and robot traffic/path control.

The rebuilt repository remains far behind these commercial delivery baselines. The current domain/lifecycle recovery stays P0 because higher-level SCADA, integration and simulation are unsafe if command outcomes and physical recovery authority are ambiguous.

Priority after this PR is verified:

1. complete lifecycle recovery/failover semantics and segmented MOVE with node confirmation/lookahead/incremental trusted release/reroute;
2. adapter SDK plus Modbus TCP, OPC UA and VDA5050 boundaries;
3. WMS/WES idempotent ingress and durable event delivery;
4. alarm acknowledgement/recovery, operator modes and SSE/WebSocket;
5. RBAC/audit/config versioning;
6. SCADA/material tracking and deterministic twin/FAT/SAT;
7. PostgreSQL/edge snapshot, observability, HA, deployment and backup/restore.

## Acceptance

Commercial delivery status: **NOT READY**.

Only merged code with successful CI is accepted software evidence. Physical-site SAT remains required for hardware-dependent safety claims. Unknown MOVE results must never be blindly replayed, and occupied physical resources must not be released without trusted position confirmation.

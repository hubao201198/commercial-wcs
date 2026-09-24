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

PR #3 restored the production domain/lifecycle boundary with CI PASS:

- Device / Task / Alarm models and AUTO/MANUAL/MAINTENANCE device modes;
- deadline and bounded retry semantics;
- UNKNOWN MOVE is non-retryable;
- cancellation is a request rather than a physical-state rewrite;
- compensation/recovery requires trusted non-empty physical position.

## Current PR: segmented MOVE safety

`feature/rebuild-segmented-move` restores the next P0 control primitive:

- MOVE is dispatched one physical segment at a time and cannot advance until destination-node confirmation;
- two-segment destination lookahead is exposed for atomic reservation by the existing ReservationManager;
- only trusted confirmation of the exact commanded destination advances the path and returns release authority for the previous trusted node;
- UNKNOWN segment outcome stops progression, exposes no further lookahead, does not change confirmed position and cannot be blindly redispatched;
- malformed paths, untrusted confirmations and mismatched-node confirmations fail closed;
- deterministic regression tests cover lookahead, confirmation gating, UNKNOWN no-replay and trusted incremental release authority.

This PR does not yet claim dynamic reroute, failover takeover, adapter/protocol support, persistence or hardware SAT.

## Commercial benchmark delta

Fresh public benchmark review on 2026-09-24 continues to support the control-first recovery order. GALAXIS RCS 3.0 combines planning, simulation, virtual commissioning, control, scheduling and O&M with 2D/3D digital twin and space-time coordination. BlueSword IMHS-WCS/3D-SCADA combines heterogeneous equipment control, material-position visibility, fault localization and virtual simulation. Damon combines heterogeneous conveyor/shuttle/AMR equipment with high-throughput scheduling and large-project delivery. Quicktron exposes WES/LES/RCS integration with upstream WMS/ERP/MES plus robot traffic/path control.

The rebuilt repository is still materially behind those delivery baselines. Safe segmented physical progression is P0 because protocol adapters, SCADA and upstream orchestration must not be allowed to infer physical arrival from command dispatch alone.

Priority after this PR is verified:

1. dynamic reroute plus lifecycle recovery/failover authority integrated with segmented MOVE;
2. adapter SDK plus Modbus TCP, OPC UA and VDA5050 boundaries;
3. WMS/WES idempotent ingress and durable event delivery;
4. alarm acknowledgement/recovery, operator modes and SSE/WebSocket;
5. RBAC/audit/config versioning;
6. SCADA/material tracking and deterministic twin/FAT/SAT;
7. PostgreSQL/edge snapshot, observability, HA, deployment and backup/restore.

## Acceptance

Commercial delivery status: **NOT READY**.

Only merged code with successful CI is accepted software evidence. Physical-site SAT remains required for hardware-dependent safety claims. Unknown MOVE results must never be blindly replayed, and occupied physical resources must not be released without trusted position confirmation.

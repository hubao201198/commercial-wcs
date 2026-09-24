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

PR #4 restored segmented MOVE safety with CI PASS:

- one physical segment is dispatched at a time and exact trusted destination confirmation gates progression;
- two-segment destination lookahead supports atomic reservation;
- only trusted confirmation returns previous-node physical release authority;
- UNKNOWN segment outcomes expose no further lookahead, do not advance position and cannot be blindly redispatched.

## Current PR: safe dynamic reroute

`feature/rebuild-safe-reroute` adds a fail-closed reroute boundary:

- reroute is allowed only while the segmented MOVE is quiescent in READY state;
- an IN_FLIGHT command cannot be reinterpreted as a new route;
- an UNKNOWN MOVE cannot be rerouted around, because physical outcome must be reconciled first;
- replacement route origin must exactly equal the current trusted confirmed node;
- deterministic regressions cover successful reroute after trusted confirmation, in-flight refusal, UNKNOWN refusal and origin mismatch.

This PR does not yet claim failover takeover, compensation completion, adapter/protocol support, persistence or hardware SAT.

## Commercial benchmark delta

Fresh public benchmark review on 2026-09-24 continues to support the control-first recovery order. GALAXIS RCS 3.0 combines planning, simulation, virtual commissioning, control, scheduling and O&M, including time-slot reservation and spatial conflict coordination. BlueSword IMHS-WCS/3D-SCADA emphasizes heterogeneous equipment control, material-position visibility and fault localization. Damon continues broad shuttle/AMR equipment coverage and publishes cloud-edge-device swarm/path scheduling. Quicktron exposes WES/LES/RCS integration with upstream WMS/ERP/MES and robot traffic/path control.

The rebuilt repository remains materially behind these commercial baselines. Safe reroute is P0 because mature schedulers dynamically reassign work, but commercial-wcs must never turn dynamic optimization into permission to reinterpret an unresolved physical command.

Priority after this PR is verified:

1. lifecycle recovery/failover authority integrated with segmented MOVE, including explicit takeover fencing;
2. adapter SDK plus Modbus TCP, OPC UA and VDA5050 boundaries;
3. WMS/WES idempotent ingress and durable event delivery;
4. alarm acknowledgement/recovery, operator modes and SSE/WebSocket;
5. RBAC/audit/config versioning;
6. SCADA/material tracking and deterministic twin/FAT/SAT;
7. PostgreSQL/edge snapshot, observability, HA, deployment and backup/restore.

## Acceptance

Commercial delivery status: **NOT READY**.

Only merged code with successful CI is accepted software evidence. Physical-site SAT remains required for hardware-dependent safety claims. Unknown MOVE results must never be blindly replayed, rerouted around, or used to release occupied physical resources without trusted position reconciliation.

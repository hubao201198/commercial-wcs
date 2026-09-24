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

PR #5 restored fail-closed dynamic reroute with CI PASS:

- reroute is allowed only from quiescent READY state at the exact trusted confirmed node;
- IN_FLIGHT and UNKNOWN commands cannot be reinterpreted as a replacement route;
- deterministic regressions cover successful reroute and refusal boundaries.

## Current PR: failover takeover fencing

`feature/rebuild-failover-fencing` adds explicit command-authority generations:

- every physical dispatch is authorized against the current monotonically increasing control generation;
- controller promotion advances the generation so stale controllers are fenced from future MOVE dispatch;
- promotion is allowed only from quiescent READY state with a trusted confirmed node;
- IN_FLIGHT and UNKNOWN physical outcomes refuse takeover until physical reconciliation;
- failed takeover does not advance generation or mutate trusted position;
- deterministic regressions cover stale-controller fencing, successful promotion, IN_FLIGHT refusal and UNKNOWN refusal.

This PR does not yet claim durable lease persistence/consensus, compensation completion, adapter/protocol support or hardware SAT.

## Commercial benchmark delta

Fresh public benchmark review on 2026-09-24 continues to support the control-first recovery order. GALAXIS RCS 3.0 integrates planning, simulation, virtual commissioning, control, scheduling and O&M with spatial conflict avoidance, time-slot reservation and dynamic reassignment. BlueSword IMHS-WCS/3D-SCADA combines heterogeneous equipment control, online/automatic/manual operations, material-position visibility and component-level fault localization. Damon publishes cloud-edge-device four-way shuttle/AMR swarm scheduling with conflict-free route generation. Quicktron exposes WES/LES/RCS integration with WMS/ERP/MES and robot path/traffic control.

The rebuilt repository remains materially behind these commercial baselines. Explicit takeover fencing remains P0: HA cannot be considered safe if an old controller can continue issuing physical MOVE commands after a replacement controller is promoted.

Priority after this PR is verified:

1. durable failover authority plus compensation completion integrated with task lifecycle;
2. adapter SDK plus Modbus TCP, OPC UA and VDA5050 boundaries;
3. WMS/WES idempotent ingress and durable event delivery;
4. alarm acknowledgement/recovery, operator modes and SSE/WebSocket;
5. RBAC/audit/config versioning;
6. SCADA/material tracking and deterministic twin/FAT/SAT;
7. PostgreSQL/edge snapshot, observability, HA, deployment and backup/restore.

## Acceptance

Commercial delivery status: **NOT READY**.

Only merged code with successful CI is accepted software evidence. Physical-site SAT remains required for hardware-dependent safety claims. Unknown MOVE results must never be blindly replayed, rerouted around, or used to release occupied physical resources without trusted position reconciliation.

# Commercial WCS Progress

Updated: 2026-09-24

## Repository recovery status

The GitHub repository was recreated on 2026-09-24. Acceptance is based only on code and tests present in the recreated repository; historical implementation claims are context, not evidence.

## Verified on main

PR #2 restored atomic resource reservation, deterministic wait-for/deadlock detection, trusted-position physical-resource release, and CI.

PR #3 restored Device / Task / Alarm models, AUTO/MANUAL/MAINTENANCE modes, deadline/bounded retry, UNKNOWN MOVE no-replay, cancellation-as-request, and trusted-position compensation/recovery gates.

PR #4 restored segmented MOVE safety: one physical segment at a time, exact trusted destination confirmation, two-segment lookahead, trusted release authority, and UNKNOWN fail-closed behavior.

PR #5 restored fail-closed dynamic reroute: only quiescent READY state at the exact trusted confirmed node may reroute; IN_FLIGHT and UNKNOWN cannot be reinterpreted as a replacement route.

PR #6 restored explicit failover takeover fencing with CI PASS: monotonic command-authority generations, stale-controller dispatch rejection, and takeover refusal for IN_FLIGHT/UNKNOWN physical execution.

## Current PR: durable failover authority boundary

`feature/durable-failover-authority` adds an AuthorityStore CAS contract and a restorable DurableControlLease:

- controller restart restores the last authority generation rather than resetting fencing state;
- promotion persists generation advancement with compare-and-swap before exposing new authority;
- competing controllers restored from the same generation cannot both promote successfully;
- every authorization checks both the controller's cached generation and current store generation, so authority advanced elsewhere fences stale dispatch;
- physical promotion remains gated by quiescent READY state and trusted confirmed position.

The included MemoryAuthorityStore is deterministic test/simulation evidence for the CAS contract, not a claim of production durability. A PostgreSQL/consensus-backed AuthorityStore and multi-process/host failover tests remain required before HA is accepted for commercial delivery.

## Commercial benchmark delta

Fresh public benchmark review on 2026-09-24 continues to support the control-first recovery order. BlueSword Pro-WCS publicly combines task/path coordination, 3D-SCADA component-level diagnosis and 99.9% availability, while VirtuSync covers simulation and virtual commissioning. Damon publishes cloud-edge-device shuttle/AMR coordination and conflict-free route generation. Quicktron exposes WES/LES/RCS integration with upstream WMS/ERP/MES plus traffic control, multi-robot collaboration, operations and simulation capabilities. GALAXIS remains a benchmark for layered WCS/RCS and large-scale robot coordination.

The rebuilt repository remains materially behind these commercial baselines. Durable command authority is a prerequisite for credible HA, but the current PR intentionally stops at a storage/consensus abstraction plus deterministic CAS tests rather than pretending an in-memory store is production durability.

Priority after this PR is verified:

1. PostgreSQL-backed authority CAS and crash/restart integration test, plus compensation completion;
2. adapter SDK plus Modbus TCP, OPC UA and VDA5050 boundaries;
3. WMS/WES idempotent ingress and durable event delivery;
4. alarm acknowledgement/recovery, operator modes and SSE/WebSocket;
5. RBAC/audit/config versioning;
6. SCADA/material tracking and deterministic twin/FAT/SAT;
7. edge snapshot, observability, HA deployment and backup/restore.

## Acceptance

Commercial delivery status: **NOT READY**.

Only merged code with successful CI is accepted software evidence. Physical-site SAT remains required for hardware-dependent safety claims. Unknown MOVE results must never be blindly replayed, rerouted around, or used to release occupied physical resources without trusted position reconciliation.

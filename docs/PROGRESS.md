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

PR #7 added the AuthorityStore CAS contract and restorable DurableControlLease with CI PASS: restart restores generation, competing promotions cannot both win, and stale cached controllers are fenced against the current store generation. MemoryAuthorityStore remains test/simulation only; production PostgreSQL/consensus durability is still required.

## Current PR: compensation completion safety

`feature/compensation-completion` closes the task-lifecycle gap after compensation begins:

- adds an explicit CANCELLED terminal state rather than treating a compensation command as task completion;
- CompleteCompensation is legal only from COMPENSATING;
- cancellation completion requires a trusted, non-empty post-compensation physical position;
- failed completion attempts do not mutate lifecycle state;
- deterministic regressions cover untrusted/empty position refusal, valid completion, and wrong-state refusal.

This does not claim that the compensated position is automatically safe for every device type. Device-specific safe-zone/interlock policy and physical SAT remain required before releasing protected resources or handing a device back to AUTO.

## Commercial benchmark delta

Fresh public benchmark review on 2026-09-24 continues to support the control-first recovery order. GALAXIS RCS 3.0 integrates planning, simulation, virtual commissioning, control, scheduling and O&M with 2D/3D digital twin and spatial-temporal coordination. BlueSword Pro-WCS combines task/path coordination, real-time equipment monitoring, 3D-SCADA component-level diagnosis and 99.9% availability, while VirtuSync covers simulation and virtual commissioning. Damon publishes cloud-edge-device shuttle/AMR coordination with swarm scheduling and conflict-free route generation. Quicktron exposes WES/LES/RCS integration with WMS/ERP/MES plus traffic control, multi-robot collaboration, operations and simulation.

The rebuilt repository remains materially behind these commercial baselines. The immediate safety gap is no longer just entering recovery safely: compensation must not be declared complete from a command acknowledgement without post-action physical reconciliation.

Priority after this PR is verified:

1. PostgreSQL-backed authority CAS and crash/restart integration test; device-specific compensation safe-zone/interlock and AUTO handback;
2. adapter SDK plus Modbus TCP, OPC UA and VDA5050 boundaries;
3. WMS/WES idempotent ingress and durable event delivery;
4. alarm acknowledgement/recovery, operator modes and SSE/WebSocket;
5. RBAC/audit/config versioning;
6. SCADA/material tracking and deterministic twin/FAT/SAT;
7. edge snapshot, observability, HA deployment and backup/restore.

## Acceptance

Commercial delivery status: **NOT READY**.

Only merged code with successful CI is accepted software evidence. Physical-site SAT remains required for hardware-dependent safety claims. Unknown MOVE results must never be blindly replayed, rerouted around, or used to release occupied physical resources without trusted position reconciliation.

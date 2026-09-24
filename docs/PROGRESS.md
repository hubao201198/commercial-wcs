# Commercial WCS Progress

Updated: 2026-09-24

## Repository recovery status

The GitHub repository was recreated on 2026-09-24. The current `main` is a minimal AI Warehouse OS rebuild and does **not** contain the previously reported commercial-WCS implementation history. Previous claims such as PostgreSQL edge persistence, Modbus/VDA5050, segmented MOVE, two-segment lookahead, rerouting, lifecycle recovery, commissioning evidence, RBAC hardening, and 9/15 acceptance status must therefore be treated as historical context only, not as code currently present in this repository.

## This PR

`feature/rebuild-control-safety-core` starts recovery from the actual current `main`:

- adds a Go `control.ReservationManager` with atomic all-or-none resource reservation;
- records wait-for dependencies and detects cycles deterministically;
- makes trusted device-position confirmation the only physical-resource release path;
- adds regression tests proving conflict atomicity, fail-closed untrusted release, deadlock detection, and wait-edge cleanup;
- restores PR CI for Go test/vet and frontend build.

## Commercial benchmark delta

Public benchmark material shows that mature systems already combine heterogeneous equipment control, automatic/manual operations, material-position visibility, fault localization/recovery, digital twin/virtual commissioning, multi-robot routing, and upstream WMS/ERP/MES integration. The rebuilt repository is currently far behind that delivery baseline.

Priority after this safety kernel is verified:

1. device/task/alarm domain model and safe task lifecycle;
2. adapter SDK plus Modbus TCP, OPC UA and VDA5050 boundaries;
3. segmented MOVE with node confirmation, lookahead reservation, trusted release and rerouting;
4. WMS/WES idempotent ingress and durable event delivery;
5. alarm acknowledgement/recovery plus operator modes and real-time SSE/WebSocket;
6. RBAC/audit/config versioning;
7. SCADA/material tracking and deterministic twin/FAT/SAT;
8. PostgreSQL/edge snapshot, observability, HA, deployment and backup/restore.

## Acceptance

Commercial delivery status: **NOT READY**.

No historical acceptance PASS is carried forward until equivalent code and deterministic tests exist in the recreated repository. Physical-site SAT remains required for hardware-dependent safety claims. Unknown MOVE results must never be blindly replayed, and occupied physical resources must not be released without trusted position confirmation.

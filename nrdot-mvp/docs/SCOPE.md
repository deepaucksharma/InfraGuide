# NRDOT+ MVP Project Scope

## Overview

The NRDOT+ MVP (Minimum Viable Product) is a proof-of-concept implementation of the NRDOT+ v9.0 platform. This MVP focuses on demonstrating the three most technically challenging runtime behaviors of the full platform:

1. Dynamic cardinality control
2. Priority queuing with spilling to disk
3. Enhanced durability and resilience (72h durability & replay)

This document outlines the scope, boundaries, and success criteria for the MVP implementation.

## Core MVP Goals

The MVP aims to demonstrate the following key capabilities:

| Category | Goal | Success Criteria |
|----------|------|------------------|
| Performance | Low CPU usage | ≤2% CPU utilization on m6i.large (or equivalent) |
| Memory | Efficient memory usage | ≤150 MiB RSS (including 64 MiB ballast) |
| Cardinality | Dynamic control of high-cardinality metrics | Hash table size ≤65,536 unique key-sets with drop/aggregate capability |
| Priority | Priority-based processing | APQ with WRR scheduling (5:3:1 ratio) |
| Resilience | DLQ survival | Container crash & filesystem fsync with SHA-256 verification |
| Replay | Rate-limited replay | 4 MiB/s, interleaved 1:1 with live traffic |
| Monitoring | Self-observability | Prometheus metrics for APQ, DLQ, GC, drops |

## Functional Requirements

| ID | Requirement | Target |
|----|-------------|--------|
| FR-1 | Accept OTLP/HTTP metrics, logs, traces on `:4318` | gRPC off |
| FR-2 | Cardinality limiter must clamp hash-table ≤ 65,536 unique key-sets, **drop/aggregate** above threshold | ≤1 ms 95ᵖ latency |
| FR-3 | APQ must provide **critical : high : normal** WRR 5 : 3 : 1 selection | O(1) enqueue / dequeue |
| FR-4 | Spill to DLQ when queue ≥ 95 % or circuit-breaker open | No data loss |
| FR-5 | DLQ must survive container crash & filesystem fsync | SHA-256 verify |
| FR-6 | Replay rate-limited to 4 MiB/s and interleaved 1:1 with live traffic | configurable |
| FR-7 | Collector RSS ≤ 150 MiB (including 64 MiB ballast) & CPU ≤ 2 % on m6i.large with 10 k spans/s | measured |
| FR-8 | Expose self-metrics for APQ, DLQ, GC, drops | Prometheus |

## Non-Functional Targets

| Category | Target |
|----------|--------|
| **Latency** | P99 pipeline ≤ 50 ms |
| **Durability** | 24 h at 100 k items/h → 15 GiB DLQ |
| **Portability** | Docker, docker-compose, kind (K8s v1.29) |
| **Build repeatability** | `make build` from clean clone < 90 s |

## Scope Exclusions

The following items are explicitly **out of scope** for the MVP implementation:

- ETW (Event Tracing for Windows) integration
- eBPF real capture mechanisms
- Profiling capabilities
- WASM processor implementation
- Kubernetes Operator development
- Multi-node federation
- Security/compliance features

## Components

The MVP includes the following major components:

1. **OpenTelemetry Collector** (single-node, single-process)
2. **Three Experimental Plugins**:
   - CardinalityLimiter processor (entropy-based)
   - Adaptive Priority Queue (APQ) wrapper
   - Enhanced DLQ (file-storage with SHA-256)
3. **Testing Utilities**:
   - Mock upstream service
   - Workload generator
   - Outage simulation
4. **Monitoring Stack**:
   - Prometheus for metrics collection
   - Grafana for visualization

## Success Metrics

The MVP will be considered successful if:

1. All functional requirements (FR-1 through FR-8) are met
2. Non-functional targets are achieved
3. The system demonstrates resilience during simulated outages
4. Developer experience is smooth (clone → build → run in < 10 minutes)
5. All components are properly documented

## Deliverables

The following deliverables are expected from the MVP implementation:

1. Source code repository with all components
2. Docker images for easy deployment
3. Documentation covering architecture, usage, and testing
4. Demo scripts for showcasing the key capabilities
5. Performance benchmark results
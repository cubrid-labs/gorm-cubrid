# Performance Guide

This document describes performance behavior for `gorm-cubrid` and practical tuning guidance.

## Overview

`gorm-cubrid` is a GORM dialect that runs on top of the `cubrid-go` driver.

```mermaid
flowchart LR
    App[Go Application] --> GORM[GORM ORM]
    GORM --> Dialect[gorm-cubrid Dialector]
    Dialect --> Driver[cubrid-go]
    Driver --> CAS[CAS Binary Protocol]
    CAS --> Server[(CUBRID Server)]
```

```mermaid
flowchart TD
    ModelOp[GORM model operation] --> Reflect[Reflection + schema mapping]
    Reflect --> Build[SQL build]
    Build --> Exec[database/sql execute]
    Exec --> Scan[Row scan + struct mapping]
    Scan --> Hook[Optional hooks/callbacks]
```

## Benchmark Results

Source: [cubrid-benchmark](https://github.com/cubrid-labs/cubrid-benchmark)

Environment: Intel Core i5-9400F @ 2.90GHz, 6 cores, Linux x86_64, Docker containers.

Driver baseline workload: Go `cubrid-go` vs `go-sql-driver/mysql`, 1000 rows x 5 rounds.

Observed outcome: near parity (approximately 1:1 ratio).

Note: GORM adds ORM-layer overhead, mainly reflection and model mapping cost.

## Performance Characteristics

- Baseline driver parity is strong; ORM overhead determines most extra latency in app paths.
- Reflection-based model metadata and change tracking add CPU work for each operation.
- Callback/hook chains can add measurable per-row overhead in high-throughput flows.
- Bulk operations dramatically reduce ORM overhead per row.

## Optimization Tips

- Prefer batch operations (`CreateInBatches`, bulk updates) for write-heavy workloads.
- Use selective fields (`Select`, `Omit`) to reduce model mapping work.
- Minimize hooks on hot paths or move logic to set-based SQL when possible.
- Tune `database/sql` pool settings under GORM's underlying `sql.DB`.

```mermaid
flowchart TD
    Start[Optimize GORM path] --> Batch{Bulk write workload?}
    Batch -->|Yes| CreateBatch[Use CreateInBatches / bulk patterns]
    Batch -->|No| Fields{Too many mapped fields?}
    Fields -->|Yes| Narrow[Use Select/Omit]
    Fields -->|No| Hooks{Heavy callbacks enabled?}
    Hooks -->|Yes| ReduceHooks[Trim hook/callback chain]
    Hooks -->|No| PoolTune[Tune sql.DB pool]
```

## Running Benchmarks

1. Clone `https://github.com/cubrid-labs/cubrid-benchmark`.
2. Launch benchmark Docker databases (CUBRID and MySQL).
3. Run Go driver baseline benchmarks for parity reference.
4. Run equivalent GORM workloads using matching row counts and rounds.
5. Compare baseline vs ORM-layer timings to identify reflection/mapping overhead.

See benchmark repo docs for exact scripts and command options.

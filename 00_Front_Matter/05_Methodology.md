# Methodology Charter

This report applies a rigorous methodology to ensure accurate, reproducible, and actionable insights. Our approach prioritizes technical depth over breadth, with a zero-gap treatment of core functionality.

## Data Sources

All findings are based on the following primary data sources:

1. **New Relic Database (NRDB) Benchmarks**:
   - Synthetic workload tests at 1GB, 10GB, 100GB, and 1TB scale
   - Performance metrics across all query patterns
   - Cardinality explosion scenarios with controlled variable isolation

2. **Kubernetes Test Harness**:
   - Standard 3-node (4vCPU, 16GB) GKE and EKS clusters
   - Workload simulator generating configurable telemetry signals
   - Fault-injection framework for resilience testing

3. **Customer Deployment Analysis**:
   - Anonymized telemetry from 50+ enterprise deployments
   - Spanning 5M+ containers and 100K+ hosts
   - Representing diverse industry verticals (finance, retail, technology, healthcare)

4. **Vendor Documentation**:
   - Official New Relic, Datadog, and OpenTelemetry documentation
   - Internal architecture documentation (where available)
   - Open-source codebases and specifications

## Testing Methodology

### Performance Testing

All performance measurements follow these protocols:

1. **Warm-up period**: Minimum 10-minute warm-up to ensure steady-state
2. **Measurement window**: 1-hour steady-state collection
3. **Statistical significance**: Minimum 30 samples per data point
4. **Variance control**: Tests discarded if coefficient of variation > 15%
5. **Environment isolation**: Dedicated infrastructure with verified baseline performance

### Load Generation

1. **k6** for HTTP-based load testing of APIs and endpoints
2. **Terraform + Ansible** for infrastructure provisioning
3. **Custom telemetry simulators** for generating metrics, logs, and traces at configurable cardinality

## Scope Limitations

This report explicitly does not cover:

1. **Edge IoT observability** (addressed separately in Ch 61 IoT Gateways)
2. **Mobile application instrumentation**
3. **Synthetic monitoring frameworks**
4. **Business intelligence integration**
5. **Machine learning performance analysis**

## Reproducibility

All configurations, harnesses, and test scripts referenced in the report are available in the accompanying GitHub repository:

`https://github.com/newrelic/observability-landscape-benchmarks`

Readers are encouraged to reproduce findings in their own environments and contribute to the ongoing research.
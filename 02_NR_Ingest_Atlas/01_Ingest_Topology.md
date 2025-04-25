# New Relic Ingest Topology Overview

## Executive Summary

New Relic's ingest topology represents one of the most advanced and scalable observability data pipelines in the industry. This chapter provides a comprehensive analysis of the architecture, components, data flows, and integration points across the entire observability lifecycle. By understanding this topology in depth, architects can make optimal decisions for instrumenting Kubernetes environments and effectively managing telemetry at scale.

## Nine-Plane Reference Architecture

New Relic's ingest architecture is structured as a nine-plane model, with each plane handling a specific function in the observability pipeline. This architecture enables both flexibility in deployment patterns and consistent handling of telemetry data across diverse environments.

```mermaid
flowchart TD
    subgraph "Customer Environment"
        A[1. Client Instrumentation] --> B[2. Collection Layer]
        B --> C[3. Local Processing]
    end
    
    subgraph "Transit Layer"
        C --> D[4. Secure Transport]
        D --> E[5. Gateway Services]
    end
    
    subgraph "New Relic Platform"
        E --> F[6. Ingest Services]
        F --> G[7. Normalization & Enrichment]
        G --> H[8. Streaming Analytics]
        H --> I[9. NRDB Storage]
    end
    
    style A fill:#f5f5f5,stroke:#333,stroke-width:2px
    style B fill:#f5f5f5,stroke:#333,stroke-width:2px
    style C fill:#f5f5f5,stroke:#333,stroke-width:2px
    style D fill:#e6f3e6,stroke:#333,stroke-width:2px
    style E fill:#e6f3e6,stroke:#333,stroke-width:2px
    style F fill:#e6e6ff,stroke:#333,stroke-width:2px
    style G fill:#e6e6ff,stroke:#333,stroke-width:2px
    style H fill:#e6e6ff,stroke:#333,stroke-width:2px
    style I fill:#e6e6ff,stroke:#333,stroke-width:2px
```

## Detailed Plane Analysis

### 1. Client Instrumentation Plane

The instrumentation plane represents the initial point of telemetry collection, capturing data directly from applications, services, and infrastructure components.

```mermaid
flowchart LR
    subgraph "Client Instrumentation Plane"
        direction TB
        A1[APM Agents] --> A[Instrumentation Router]
        A2[Infrastructure Agents] --> A
        A3[Browser/Mobile] --> A
        A4[Kubernetes Integration] --> A
        A5[OpenTelemetry SDKs] --> A
        A6[Custom Integrations] --> A
    end
    
    subgraph "Instrumentation Methods"
        B1[Auto-Instrumentation]
        B2[Manual Instrumentation]
        B3[Agent Configuration]
        B4[Infrastructure Discovery]
        B5[In-Process Monitoring]
        B6[Out-of-Process Monitoring]
    end
    
    A --> B1
    A --> B2
    A --> B3
    A --> B4
    A --> B5
    A --> B6
    
    style A fill:#f5f5f5,stroke:#333,stroke-width:2px
    style A1 fill:#e6e6ff,stroke:#333
    style A2 fill:#e6e6ff,stroke:#333
    style A3 fill:#e6e6ff,stroke:#333
    style A4 fill:#e6e6ff,stroke:#333
    style A5 fill:#e6e6ff,stroke:#333
    style A6 fill:#e6e6ff,stroke:#333
    style B1 fill:#f9f9f9,stroke:#333
    style B2 fill:#f9f9f9,stroke:#333
    style B3 fill:#f9f9f9,stroke:#333
    style B4 fill:#f9f9f9,stroke:#333
    style B5 fill:#f9f9f9,stroke:#333
    style B6 fill:#f9f9f9,stroke:#333
```

#### Instrumentation Options by Signal Type

| Signal Type | Instrumentation Options | Data Format | K8s-Specific Considerations | 
|-------------|-------------------------|-------------|----------------------------|
| **Metrics** | Infrastructure Agent<br>Prometheus Integration<br>OpenTelemetry SDK<br>Kubernetes Integration | Dimensional<br>OTLP<br>Prometheus Exposition | High-cardinality pod/container labels<br>Resource metrics vs custom metrics<br>Scrape vs push models |
| **Logs** | Infrastructure Agent<br>Fluent Bit/Fluentd<br>OpenTelemetry Collector<br>Kubernetes Event API | JSON<br>Plain text<br>Structured logging | Container stdout/stderr<br>DaemonSet vs Sidecar patterns<br>Log volume management |
| **Traces** | APM Agents<br>OpenTelemetry SDK<br>OpenTracing (legacy)<br>Manual API Calls | W3C TraceContext<br>OTLP<br>Zipkin<br>Jaeger | Service mesh integration<br>Cross-namespace tracing<br>Sampling strategies |
| **Events** | Infrastructure Agent<br>Direct API<br>OpenTelemetry SDK<br>Kubernetes Events | JSON<br>OTLP | Control plane events<br>Deployment events<br>Custom K8s lifecycle events |
| **Synthetic** | Containerized monitors<br>In-cluster checks<br>External checks | API results<br>Check outputs | Intra-cluster connectivity<br>Ingress/service checks<br>Cross-cluster validation |

### 2. Collection Layer

The collection layer aggregates and prepares telemetry before transmission, often running within the Kubernetes cluster itself.

```mermaid
flowchart TD
    A1[Applications] --> B1[APM Agents]
    A2[Services] --> B2[OpenTelemetry SDK]
    A3[Infrastructure] --> B3[Infrastructure Agent]
    A4[Kubernetes] --> B4[Kubernetes Integration]
    
    B1 --> C[Collection Layer]
    B2 --> C
    B3 --> C
    B4 --> C
    
    C --> D1[Batching]
    C --> D2[Filtering]
    C --> D3[Transformation]
    C --> D4[Sampling]
    C --> D5[Buffering]
    
    style A1 fill:#f5f5f5,stroke:#333
    style A2 fill:#f5f5f5,stroke:#333
    style A3 fill:#f5f5f5,stroke:#333
    style A4 fill:#f5f5f5,stroke:#333
    style B1 fill:#e6e6ff,stroke:#333
    style B2 fill:#e6e6ff,stroke:#333
    style B3 fill:#e6e6ff,stroke:#333
    style B4 fill:#e6e6ff,stroke:#333
    style C fill:#f9f9f9,stroke:#333,stroke-width:2px
    style D1 fill:#e6f3e6,stroke:#333
    style D2 fill:#e6f3e6,stroke:#333
    style D3 fill:#e6f3e6,stroke:#333
    style D4 fill:#e6f3e6,stroke:#333
    style D5 fill:#e6f3e6,stroke:#333
```

#### Kubernetes Collection Patterns

| Pattern | Implementation | Pros | Cons | Best For |
|---------|----------------|------|------|----------|
| **DaemonSet** | Agent on every node | Complete node visibility<br>Low network overhead | Resource impact on nodes<br>Requires node privileges | Infrastructure metrics<br>Node-level logs |
| **Sidecar** | Per-pod agent container | Isolation of concerns<br>Namespace-level permissions | Resource overhead<br>Deployment complexity | Application logs<br>Service-specific metrics |
| **Cluster Agent** | Centralized collection | Efficient resource usage<br>Simplified maintenance | Potential bottleneck<br>Less isolation | Cluster-level metrics<br>Control plane monitoring |
| **Out-of-Cluster** | External collector | Zero in-cluster overhead<br>Independent failure domain | Network dependency<br>Limited access to internal data | Synthetic checks<br>External availability monitoring |
| **Operator-Managed** | Custom resource definitions | Kubernetes-native management<br>GitOps compatibility | Additional CRDs required<br>Operator overhead | Production environments<br>Multi-cluster deployments |

### 3. Local Processing Plane

Before transmitting data to New Relic, local processing optimizes telemetry for efficient transport and reduces unnecessary data volume.

```mermaid
flowchart LR
    A[Raw Telemetry] --> B{Local Processing}
    
    B --> C1[Aggregation]
    C1 --> D1[Statistical summaries]
    C1 --> D2[Pre-computed rollups] 
    
    B --> C2[Filtering]
    C2 --> D3[Label filtering]
    C2 --> D4[Value thresholds]
    C2 --> D5[PII removal]
    
    B --> C3[Transformation]
    C3 --> D6[Format conversion]
    C3 --> D7[Enrichment]
    C3 --> D8[Standardization]
    
    B --> C4[Sampling]
    C4 --> D9[Tail-based sampling]
    C4 --> D10[Head-based sampling]
    C4 --> D11[Adaptive rate]
    
    style A fill:#f5f5f5,stroke:#333
    style B fill:#f9f9f9,stroke:#333,stroke-width:2px
    style C1 fill:#e6e6ff,stroke:#333
    style C2 fill:#e6e6ff,stroke:#333
    style C3 fill:#e6e6ff,stroke:#333
    style C4 fill:#e6e6ff,stroke:#333
    style D1 fill:#e6f3e6,stroke:#333
    style D2 fill:#e6f3e6,stroke:#333
    style D3 fill:#e6f3e6,stroke:#333
    style D4 fill:#e6f3e6,stroke:#333
    style D5 fill:#e6f3e6,stroke:#333
    style D6 fill:#e6f3e6,stroke:#333
    style D7 fill:#e6f3e6,stroke:#333
    style D8 fill:#e6f3e6,stroke:#333
    style D9 fill:#e6f3e6,stroke:#333
    style D10 fill:#e6f3e6,stroke:#333
    style D11 fill:#e6f3e6,stroke:#333
```

#### Local Processing Optimization Strategies

| Strategy | Technique | Typical Reduction | Use Cases | K8s Implementation |
|----------|-----------|-------------------|-----------|-------------------|
| **Metric Aggregation** | Pre-compute statistics | 90-99% | High-frequency metrics<br>System-level metrics | ConfigMap-based aggregation rules<br>OTel Collector processors |
| **Dimensional Filtering** | Drop unnecessary labels | 30-70% | High-cardinality metrics<br>Auto-generated labels | Label allow/deny lists<br>Relabeling configurations |
| **Log Filtering** | Pattern-based exclusion | 40-80% | Container logs<br>System logs | Fluent Bit filters<br>Vector transforms |
| **Log Parsing** | Extract structured data | 10-30% | Unstructured logs<br>Multi-line logs | Parser configurations<br>Regex extraction rules |
| **Trace Sampling** | Head/tail-based decisions | 90-99% | High-volume services<br>Background processes | Sampling processors<br>Service-level configuration |
| **Semantic Conventions** | Standardize naming | N/A (quality) | Cross-team observability<br>Service correlation | OTel semantic conventions<br>Custom resource attributes |

### 4. Secure Transport Plane

The transport layer ensures secure, reliable transmission of telemetry data to New Relic's ingest endpoints.

```mermaid
flowchart TD
    A[Local Processing] --> B[Secure Transport]
    
    B --> C1[Encryption]
    C1 --> D1[TLS 1.3+]
    C1 --> D2[Certificate Validation]
    
    B --> C2[Compression]
    C2 --> D3[gzip]
    C2 --> D4[deflate]
    
    B --> C3[Reliability]
    C3 --> D5[Retry Logic]
    C3 --> D6[Circuit Breaking]
    C3 --> D7[Load Shedding]
    
    B --> C4[Efficiency]
    C4 --> D8[Batch Sizing]
    C4 --> D9[Connection Pooling]
    C4 --> D10[Persistent Connections]
    
    style A fill:#f5f5f5,stroke:#333
    style B fill:#f9f9f9,stroke:#333,stroke-width:2px
    style C1 fill:#e6e6ff,stroke:#333
    style C2 fill:#e6e6ff,stroke:#333
    style C3 fill:#e6e6ff,stroke:#333
    style C4 fill:#e6e6ff,stroke:#333
    style D1 fill:#e6f3e6,stroke:#333
    style D2 fill:#e6f3e6,stroke:#333
    style D3 fill:#e6f3e6,stroke:#333
    style D4 fill:#e6f3e6,stroke:#333
    style D5 fill:#e6f3e6,stroke:#333
    style D6 fill:#e6f3e6,stroke:#333
    style D7 fill:#e6f3e6,stroke:#333
    style D8 fill:#e6f3e6,stroke:#333
    style D9 fill:#e6f3e6,stroke:#333
    style D10 fill:#e6f3e6,stroke:#333
```

#### Transport Optimization Matrix

| Transport Parameter | Optimal Value | Impact on Performance | Tradeoffs | K8s Consideration |
|--------------------|---------------|----------------------|-----------|-------------------|
| **Batch Size** | 1-5 MB | 5-10× efficiency gain | Memory usage<br>Data freshness | Container memory limits<br>Resource requests |
| **Compression Level** | 6-7 (gzip) | 5-10× bandwidth reduction | CPU usage<br>Latency | CPU limits<br>QoS class |
| **Connection Pooling** | 5-10 connections | Reduced connection overhead | Resource consumption | Network policies<br>Egress traffic |
| **Retry Strategy** | Exp. backoff + jitter | Resilience during instability | Potential data delay | Pod disruption budgets<br>Graceful termination |
| **Send Frequency** | 5-15 seconds | Balance of freshness vs efficiency | Burst potential<br>Battery/resource usage | Liveness/readiness probes<br>Resource limits |
| **Payload Protocol** | Protobuf/OTLP | 30-50% more efficient than JSON | Tooling compatibility<br>Debugging difficulty | Protocol support in collectors |

### 5. Gateway Services Plane

Gateway services represent the first contact point within New Relic's infrastructure, handling authentication, validation, and initial routing.

```mermaid
flowchart LR
    A[Secure Transport] --> B[Gateway Services]
    
    B --> C1[Authentication]
    C1 --> D1[License Key Validation]
    C1 --> D2[API Key Authentication]
    
    B --> C2[Validation]
    C2 --> D3[Schema Validation]
    C2 --> D4[Size Limits]
    C2 --> D5[Rate Limiting]
    
    B --> C3[Routing]
    C3 --> D6[Regional Routing]
    C3 --> D7[Service Discovery]
    C3 --> D8[Load Balancing]
    
    B --> C4[Protection]
    C4 --> D9[DDoS Mitigation]
    C4 --> D10[Traffic Shaping]
    
    style A fill:#f5f5f5,stroke:#333
    style B fill:#f9f9f9,stroke:#333,stroke-width:2px
    style C1 fill:#e6e6ff,stroke:#333
    style C2 fill:#e6e6ff,stroke:#333
    style C3 fill:#e6e6ff,stroke:#333
    style C4 fill:#e6e6ff,stroke:#333
    style D1 fill:#e6f3e6,stroke:#333
    style D2 fill:#e6f3e6,stroke:#333
    style D3 fill:#e6f3e6,stroke:#333
    style D4 fill:#e6f3e6,stroke:#333
    style D5 fill:#e6f3e6,stroke:#333
    style D6 fill:#e6f3e6,stroke:#333
    style D7 fill:#e6f3e6,stroke:#333
    style D8 fill:#e6f3e6,stroke:#333
    style D9 fill:#e6f3e6,stroke:#333
    style D10 fill:#e6f3e6,stroke:#333
```

#### Gateway Regional Distribution

New Relic maintains global gateway infrastructure to minimize latency and provide regional compliance options:

| Region | Location | Endpoints | Compliance Certifications | Best For |
|--------|----------|-----------|---------------------------|----------|
| **US East** | Virginia | metrics-api.newrelic.com<br>log-api.newrelic.com<br>trace-api.newrelic.com | SOC 2<br>ISO 27001<br>GDPR | North American workloads |
| **US West** | Oregon | collector.newrelic.com<br>insights-collector.newrelic.com | SOC 2<br>ISO 27001<br>FedRAMP | US West Coast<br>Gov Cloud |
| **EU** | Frankfurt | eu01-metrics-api.newrelic.com<br>eu01-log-api.newrelic.com | SOC 2<br>ISO 27001<br>GDPR<br>Schrems II | European workloads<br>GDPR requirements |
| **Asia Pacific** | Singapore | ap01-metrics-api.newrelic.com<br>ap01-log-api.newrelic.com | SOC 2<br>ISO 27001 | APAC workloads |
| **Dedicated** | Customer-specific | custom endpoints | Custom certifications | Air-gapped environments<br>Special compliance needs |

### 6. Ingest Services Plane

Specialized services process different types of telemetry, with dedicated pipelines optimized for each data type.

```mermaid
flowchart TD
    A[Gateway Services] --> B[Ingest Services]
    
    B --> C1[Metric Pipeline]
    C1 --> D1[Dimensional Metrics]
    C1 --> D2[Time-Series Metrics]
    
    B --> C2[Event Pipeline]
    C2 --> D3[Structured Events]
    C2 --> D4[Custom Events]
    C2 --> D5[APM Events]
    
    B --> C3[Log Pipeline]
    C3 --> D6[Structured Logs]
    C3 --> D7[Plain Text Logs]
    C3 --> D8[System Logs]
    
    B --> C4[Trace Pipeline]
    C4 --> D9[Distributed Traces]
    C4 --> D10[Span Events]
    C4 --> D11[Trace Context]
    
    B --> C5[Entity Pipeline]
    C5 --> D12[Entity Synthesis]
    C5 --> D13[Relationship Mapping]
    
    style A fill:#f5f5f5,stroke:#333
    style B fill:#f9f9f9,stroke:#333,stroke-width:2px
    style C1 fill:#e6e6ff,stroke:#333
    style C2 fill:#e6e6ff,stroke:#333
    style C3 fill:#e6e6ff,stroke:#333
    style C4 fill:#e6e6ff,stroke:#333
    style C5 fill:#e6e6ff,stroke:#333
    style D1 fill:#e6f3e6,stroke:#333
    style D2 fill:#e6f3e6,stroke:#333
    style D3 fill:#e6f3e6,stroke:#333
    style D4 fill:#e6f3e6,stroke:#333
    style D5 fill:#e6f3e6,stroke:#333
    style D6 fill:#e6f3e6,stroke:#333
    style D7 fill:#e6f3e6,stroke:#333
    style D8 fill:#e6f3e6,stroke:#333
    style D9 fill:#e6f3e6,stroke:#333
    style D10 fill:#e6f3e6,stroke:#333
    style D11 fill:#e6f3e6,stroke:#333
    style D12 fill:#e6f3e6,stroke:#333
    style D13 fill:#e6f3e6,stroke:#333
```

#### Signal-Specific Ingest Endpoints

| Signal Type | Endpoint | Protocol | Format | K8s Relevance |
|-------------|----------|----------|--------|---------------|
| **Metrics** | metrics-api.newrelic.com | HTTP/S | JSON, Protobuf | Container metrics<br>Node metrics<br>Service metrics |
| **Events** | insights-collector.newrelic.com | HTTP/S | JSON | Deployment events<br>Scaling events<br>Pod lifecycle |
| **Logs** | log-api.newrelic.com | HTTP/S | JSON, Text | Container logs<br>Application logs<br>Control plane logs |
| **Traces** | trace-api.newrelic.com | HTTP/S | JSON, Protobuf | Service-to-service communication<br>Microservice interactions |
| **Infrastructure** | infra-api.newrelic.com | HTTP/S | JSON | Host metrics<br>Container lifecycle<br>Kubernetes events |
| **OpenTelemetry** | otlp.nr-data.net | gRPC, HTTP/S | Protobuf, JSON | Full K8s telemetry<br>Cross-cutting concerns |

### 7. Normalization & Enrichment Plane

Raw telemetry is enhanced with contextual information, improving query capabilities and analysis.

```mermaid
flowchart TD
    A[Ingest Services] --> B[Normalization & Enrichment]
    
    B --> C1[Schema Normalization]
    C1 --> D1[Field Standardization]
    C1 --> D2[Type Conversion]
    
    B --> C2[Entity Decoration]
    C2 --> D3[Entity Identification]
    C2 --> D4[Entity Linking]
    
    B --> C3[Metadata Enrichment]
    C3 --> D5[K8s Metadata]
    C3 --> D6[Cloud Provider Data]
    C3 --> D7[Geographic Data]
    
    B --> C4[Relationship Mapping]
    C4 --> D8[Dependency Discovery]
    C4 --> D9[Service Maps]
    
    style A fill:#f5f5f5,stroke:#333
    style B fill:#f9f9f9,stroke:#333,stroke-width:2px
    style C1 fill:#e6e6ff,stroke:#333
    style C2 fill:#e6e6ff,stroke:#333
    style C3 fill:#e6e6ff,stroke:#333
    style C4 fill:#e6e6ff,stroke:#333
    style D1 fill:#e6f3e6,stroke:#333
    style D2 fill:#e6f3e6,stroke:#333
    style D3 fill:#e6f3e6,stroke:#333
    style D4 fill:#e6f3e6,stroke:#333
    style D5 fill:#e6f3e6,stroke:#333
    style D6 fill:#e6f3e6,stroke:#333
    style D7 fill:#e6f3e6,stroke:#333
    style D8 fill:#e6f3e6,stroke:#333
    style D9 fill:#e6f3e6,stroke:#333
```

#### Kubernetes Metadata Enrichment

New Relic automatically enhances telemetry with Kubernetes context:

| K8s Dimension | Added Automatically | Source | Benefits | Query Example |
|---------------|---------------------|--------|----------|---------------|
| **Cluster Name** | Yes | Kube API/Config | Cross-cluster analysis | `FROM K8sContainerSample WHERE clusterName = 'prod-east'` |
| **Namespace** | Yes | Kube API | Multi-tenant isolation | `FROM Metric WHERE namespaceName = 'staging'` |
| **Pod Name** | Yes | Kube API | Pod-level correlation | `FROM Log WHERE podName LIKE 'web-frontend-%'` |
| **Container Name** | Yes | Kube API | Container-specific analysis | `FROM ProcessSample WHERE containerName = 'api-server'` |
| **Node Name** | Yes | Kube API | Node-based correlation | `FROM K8sNodeSample WHERE nodeName = 'worker-12'` |
| **Deployment** | Yes | Kube API | Deployment-level aggregation | `FROM Metric WHERE deploymentName = 'payment-service'` |
| **Service** | Yes | Kube API | Service-level metrics | `FROM K8sContainerSample WHERE serviceName = 'checkout'` |
| **Labels** | Configurable | Kube API | Custom organization | `FROM Metric WHERE labels.app = 'inventory'` |
| **Annotations** | Configurable | Kube API | Custom metadata | `FROM Metric WHERE annotations.version = 'v2.3.4'` |

### 8. Streaming Analytics Plane

Real-time processing occurs before final storage, enabling alerting, anomaly detection, and derived metrics.

```mermaid
flowchart TD
    A[Normalization & Enrichment] --> B[Streaming Analytics]
    
    B --> C1[Real-time Alerting]
    C1 --> D1[Condition Evaluation]
    C1 --> D2[Anomaly Detection]
    
    B --> C2[Stream Processing]
    C2 --> D3[Windowed Aggregation]
    C2 --> D4[Pattern Detection]
    
    B --> C3[Derived Signals]
    C3 --> D5[Calculated Metrics]
    C3 --> D6[Golden Signals Derivation]
    
    B --> C4[SLO Tracking]
    C4 --> D7[Error Budget Calculation]
    C4 --> D8[Burndown Analysis]
    
    style A fill:#f5f5f5,stroke:#333
    style B fill:#f9f9f9,stroke:#333,stroke-width:2px
    style C1 fill:#e6e6ff,stroke:#333
    style C2 fill:#e6e6ff,stroke:#333
    style C3 fill:#e6e6ff,stroke:#333
    style C4 fill:#e6e6ff,stroke:#333
    style D1 fill:#e6f3e6,stroke:#333
    style D2 fill:#e6f3e6,stroke:#333
    style D3 fill:#e6f3e6,stroke:#333
    style D4 fill:#e6f3e6,stroke:#333
    style D5 fill:#e6f3e6,stroke:#333
    style D6 fill:#e6f3e6,stroke:#333
    style D7 fill:#e6f3e6,stroke:#333
    style D8 fill:#e6f3e6,stroke:#333
```

#### Streaming Analytics Use Cases for Kubernetes

| Analytics Type | Implementation | K8s Use Case | Benefits |
|----------------|----------------|-------------|----------|
| **Container Health** | CPU/Memory outlier detection | Identify problematic containers | Proactive resource management |
| **Pod Lifecycle** | Restart pattern detection | Spot crash loops and instability | Faster debugging of deployment issues |
| **Service Latency** | Percentile tracking + baselines | Track service degradation | Early warning of performance issues |
| **Resource Contention** | Cross-metric correlation | Identify noisy neighbors | Better workload placement |
| **Scale Event Analysis** | Event sequence detection | Validate autoscaling effectiveness | Optimization of HPA configurations |
| **Control Plane Health** | API server latency monitoring | Ensure cluster responsiveness | Prevent cluster-wide issues |
| **Deployment Success** | Rolling update tracking | Verify deployment health | Automatic rollback triggers |

### 9. NRDB Storage Plane

The final destination for all telemetry data, optimized for analytical queries and long-term storage.

```mermaid
flowchart TD
    A[Streaming Analytics] --> B[NRDB Storage]
    
    B --> C1[Ingest Pipeline]
    C1 --> D1[Schema Management]
    C1 --> D2[Write Optimization]
    
    B --> C2[Storage Management]
    C2 --> D3[Compression]
    C2 --> D4[Partitioning]
    C2 --> D5[Retention Policies]
    
    B --> C3[Query Engine]
    C3 --> D6[NRQL Processing]
    C3 --> D7[Query Optimization]
    
    B --> C4[Data Lifecycle]
    C4 --> D8[Aggregation Rollups]
    C4 --> D9[Data Expiration]
    
    style A fill:#f5f5f5,stroke:#333
    style B fill:#f9f9f9,stroke:#333,stroke-width:2px
    style C1 fill:#e6e6ff,stroke:#333
    style C2 fill:#e6e6ff,stroke:#333
    style C3 fill:#e6e6ff,stroke:#333
    style C4 fill:#e6e6ff,stroke:#333
    style D1 fill:#e6f3e6,stroke:#333
    style D2 fill:#e6f3e6,stroke:#333
    style D3 fill:#e6f3e6,stroke:#333
    style D4 fill:#e6f3e6,stroke:#333
    style D5 fill:#e6f3e6,stroke:#333
    style D6 fill:#e6f3e6,stroke:#333
    style D7 fill:#e6f3e6,stroke:#333
    style D8 fill:#e6f3e6,stroke:#333
    style D9 fill:#e6f3e6,stroke:#333
```

#### NRDB Event Types for Kubernetes

| Event Type | Content | Retention | K8s Use Cases |
|------------|---------|-----------|---------------|
| **K8sContainerSample** | Container metrics | 13 months | Container resource utilization<br>Application performance<br>Resource planning |
| **K8sPodSample** | Pod-level metrics | 13 months | Pod health<br>Scheduling effectiveness<br>Workload analysis |
| **K8sNodeSample** | Node metrics | 13 months | Cluster capacity<br>Node performance<br>Hardware issues |
| **K8sClusterSample** | Cluster metrics | 13 months | Control plane health<br>API server performance<br>Overall cluster health |
| **K8sEvent** | Kubernetes events | 7 days | Deployment events<br>Pod scheduling<br>System warnings |
| **SystemSample** | Host-level metrics | 13 months | Node-level performance<br>OS metrics<br>Hardware utilization |
| **Log** | Container/system logs | 30 days (configurable) | Application logs<br>System logs<br>Control plane logs |
| **Span** | Distributed traces | 8 days | Service interactions<br>Request flows<br>Performance bottlenecks |
| **ProcessSample** | Process metrics | 8 days | Detailed process monitoring<br>Container internals<br>Resource utilization |

## Kubernetes-Specific Integration Patterns

New Relic's ingest topology includes specialized patterns for Kubernetes environments:

```mermaid
flowchart TD
    subgraph "Kubernetes Cluster"
        A1[Infrastructure Agent<br>DaemonSet] --> B1[Host + Container<br>Metrics]
        A2[Kubernetes Integration<br>Deployment] --> B2[Cluster Metadata<br>K8s API Data]
        A3[Prometheus Integration<br>Deployment] --> B3[Scrape Endpoint<br>Metrics]
        A4[Kube State Metrics<br>Deployment] --> B4[K8s Resource<br>State Metrics]
        A5[OpenTelemetry<br>Collector] --> B5[OTLP Data]
        A6[Log Forwarders<br>DaemonSet] --> B6[Container Logs<br>System Logs]
    end
    
    B1 --> C[New Relic<br>Ingest]
    B2 --> C
    B3 --> C
    B4 --> C
    B5 --> C
    B6 --> C
    
    style A1 fill:#f5f5f5,stroke:#333
    style A2 fill:#f5f5f5,stroke:#333
    style A3 fill:#f5f5f5,stroke:#333
    style A4 fill:#f5f5f5,stroke:#333
    style A5 fill:#f5f5f5,stroke:#333
    style A6 fill:#f5f5f5,stroke:#333
    style B1 fill:#e6e6ff,stroke:#333
    style B2 fill:#e6e6ff,stroke:#333
    style B3 fill:#e6e6ff,stroke:#333
    style B4 fill:#e6e6ff,stroke:#333
    style B5 fill:#e6e6ff,stroke:#333
    style B6 fill:#e6e6ff,stroke:#333
    style C fill:#f9f9f9,stroke:#333,stroke-width:2px
```

### Kubernetes Integration Components

| Component | Deployment Method | Data Collection | Required Permissions | Resource Impact |
|-----------|-------------------|----------------|---------------------|----------------|
| **Infrastructure Agent** | DaemonSet | Host metrics<br>Container metrics<br>Kubernetes events | privileged<br>hostNetwork | 100-200MB RAM<br>5-15% CPU per node |
| **Kubernetes Integration** | Config in Infra Agent | Kubernetes API data<br>Cluster metrics<br>Resource metadata | cluster-admin<br>or custom RBAC | Minimal (uses existing agent) |
| **Prometheus Integration** | Config in Infra Agent | Scrape Prometheus endpoints<br>Convert to NR format | Basic pod permissions | Depends on scrape targets<br>Usually 50-100MB RAM |
| **OpenTelemetry Collector** | Deployment or DaemonSet | OTLP data<br>Multiple data types<br>Custom configuration | Depends on collectors<br>Usually namespace-scoped | 100-300MB RAM<br>5-20% CPU per instance |
| **Kube State Metrics** | Deployment | K8s object states<br>Resource counts<br>State transitions | read-only to cluster API | 50-100MB RAM<br>Minimal CPU |
| **Log Forwarder** | DaemonSet | Container logs<br>Node logs<br>Application logs | Access to log paths<br>Usually privileged | 50-150MB RAM<br>5-10% CPU per node |

## Performance Characteristics

The ingest topology is designed for high throughput, low latency, and exceptional reliability:

| Metric | Capability | K8s Cluster Support | Scaling Factors |
|--------|------------|---------------------|----------------|
| **Ingest Rate** | >25M data points/second/account | 1000+ node clusters | Container count<br>Metric cardinality<br>Collection frequency |
| **End-to-End Latency** | <10 seconds (p95)<br><30 seconds (p99) | Real-time monitoring | Network latency<br>Batch size<br>Processing complexity |
| **Query Performance** | Sub-second for common queries<br>1-5s for complex queries | Interactive troubleshooting | Query complexity<br>Time range<br>Cardinality of results |
| **Reliability** | 99.99% uptime commitment | Production-grade SLA | Region redundancy<br>Client-side buffering |
| **Global Distribution** | 15+ regions worldwide | Data sovereignty compliance | Regional deployment<br>Latency requirements |
| **Data Compression** | 10-20× reduction from raw data | Cost-efficient monitoring | Data types<br>Signal repetitiveness |

## Multi-Signal Correlation

One of the key advantages of New Relic's unified ingest topology is the ability to correlate across different signal types:

```mermaid
flowchart TD
    A[Kubernetes Pod/Container] --> B1[Metrics]
    A --> B2[Logs]
    A --> B3[Traces]
    A --> B4[Events]
    
    B1 --> C[New Relic Ingest Topology]
    B2 --> C
    B3 --> C
    B4 --> C
    
    C --> D[Entity Correlation]
    
    D --> E1[Performance Correlation]
    D --> E2[Root Cause Analysis]
    D --> E3[Incident Investigation]
    D --> E4[Capacity Planning]
    
    style A fill:#f5f5f5,stroke:#333
    style B1 fill:#e6e6ff,stroke:#333
    style B2 fill:#e6e6ff,stroke:#333
    style B3 fill:#e6e6ff,stroke:#333
    style B4 fill:#e6e6ff,stroke:#333
    style C fill:#f9f9f9,stroke:#333,stroke-width:2px
    style D fill:#e6f3e6,stroke:#333
    style E1 fill:#f5f5f5,stroke:#333
    style E2 fill:#f5f5f5,stroke:#333
    style E3 fill:#f5f5f5,stroke:#333
    style E4 fill:#f5f5f5,stroke:#333
```

### Cross-Signal Correlation Examples

| Correlation Type | NRQL Example | K8s Use Case | Business Value |
|------------------|--------------|--------------|---------------|
| **Metric-to-Log** | `FROM Metric, Log SELECT Metric.value, Log.message WHERE Metric.podId = Log.podId AND Metric.value > threshold` | Identify log entries when pod CPU spikes | Faster debugging of resource issues |
| **Trace-to-Metric** | `FROM Span, K8sContainerSample SELECT Span.duration, K8sContainerSample.cpuUsedCores WHERE Span.containerId = K8sContainerSample.containerId` | Correlate service latency with container performance | Identify resource-constrained services |
| **Event-to-Log** | `FROM K8sEvent, Log SELECT * WHERE K8sEvent.involvedObjectName = Log.podName AND K8sEvent.reason = 'Failed'` | Connect pod failures with log errors | Complete picture of failure scenarios |
| **Metric-to-Event** | `FROM Metric, Deployment SELECT Metric.value WHERE Metric.deploymentName = Deployment.entityName TIMESERIES` | View metrics during deployment events | Validate deployment impact |
| **Trace-to-Log** | `FROM Span, Log SELECT Span.duration, Log.message WHERE Span.traceId = Log.traceId AND Span.duration > 1` | Find error logs for slow traces | End-to-end transaction visibility |

## Implementation Decision Framework

When designing a New Relic implementation for Kubernetes, several factors should guide your ingest topology decisions:

```mermaid
flowchart TD
    A[Kubernetes Monitoring<br>Requirements] --> B{Cluster Size?}
    
    B -->|Small<br><20 nodes| C1[Simple Topology]
    B -->|Medium<br>20-100 nodes| C2[Standard Topology]
    B -->|Large<br>>100 nodes| C3[Enterprise Topology]
    
    C1 --> D1[Infrastructure Agent<br>+ K8s Integration]
    C2 --> D2[Infra Agent + K8s Integration<br>+ Prometheus + Logs]
    C3 --> D3[Infra Agent + OTel Collector<br>+ Advanced Pipelines]
    
    D1 --> E1[Single Collection Path]
    D2 --> E2[Signal-Specific Collection]
    D3 --> E3[Distributed Collection<br>with Local Processing]
    
    style A fill:#f5f5f5,stroke:#333,stroke-width:2px
    style B fill:#f9f9f9,stroke:#333,stroke-width:2px
    style C1 fill:#e6e6ff,stroke:#333
    style C2 fill:#e6e6ff,stroke:#333
    style C3 fill:#e6e6ff,stroke:#333
    style D1 fill:#e6f3e6,stroke:#333
    style D2 fill:#e6f3e6,stroke:#333
    style D3 fill:#e6f3e6,stroke:#333
    style E1 fill:#f5f5f5,stroke:#333
    style E2 fill:#f5f5f5,stroke:#333
    style E3 fill:#f5f5f5,stroke:#333
```

### Implementation Decision Matrix

Use this matrix to guide your implementation decisions:

| Factor | Simple Topology | Standard Topology | Enterprise Topology |
|--------|----------------|-------------------|---------------------|
| **Cluster Size** | <20 nodes | 20-100 nodes | >100 nodes |
| **Collection Approach** | Infrastructure Agent + integrations | Agent + specialized collectors | Distributed collection with local processing |
| **Deployment Method** | Manual/basic Helm | Helm with custom values | Operator with CRDs |
| **Data Volume** | <1M events/minute | 1-10M events/minute | >10M events/minute |
| **Cardinality Strategy** | Default settings | Selective filtering | Advanced dimensional management |
| **Local Processing** | Minimal | Basic aggregation and filtering | Advanced pipeline processing |
| **Resource Requirements** | Low | Medium | High but optimized |
| **Management Complexity** | Minimal | Moderate | Complex but automated |
| **Ideal Use Cases** | Development<br>Small production<br>Quick setup | Standard production<br>Multi-app clusters | Large enterprise<br>Multi-cluster<br>Regulated environments |

## Best Practices for Kubernetes Implementations

### Topology Optimization

1. **Right-size agent deployments**
   - Use resource limits and requests appropriate for cluster size
   - Consider node resource utilization in DaemonSet deployment

2. **Optimize collection frequency**
   - Standard metrics: 15-30 seconds
   - Critical metrics: 5-10 seconds
   - Long-term trends: 60 seconds

3. **Implement hierarchical monitoring**
   - Cluster-level golden signals
   - Namespace-level health metrics
   - Pod/container detailed telemetry

### Data Management

1. **Control metric cardinality**
   - Filter high-cardinality labels before transmission
   - Group metrics by relevant dimensions (namespace, deployment)
   - Sample high-volume, low-value signals

2. **Implement intelligent log management**
   - Use log pattern recognition to reduce volume
   - Sample debug/verbose logs in production
   - Retain error logs longer than informational logs

3. **Apply appropriate sampling strategies**
   - Infrastructure: Minimal sampling, focus on aggregation
   - Traces: Adaptive sampling based on service importance
   - Logs: Pattern-based and level-based sampling

## Conclusion

New Relic's ingest topology provides a comprehensive framework for monitoring Kubernetes environments at any scale. By understanding each of the nine planes and their interactions, you can design an observability implementation that balances completeness, performance, and cost-effectiveness.

The modular nature of the architecture allows for flexible adoption, from simple single-cluster monitoring to complex multi-cluster enterprise deployments with specialized requirements. The unified data model enables powerful cross-signal correlations that provide deeper insights than siloed monitoring approaches.

When implementing New Relic for Kubernetes, consider your specific requirements around scale, data volume, and analysis needs to select the appropriate components and configuration. The decision frameworks and best practices in this chapter will guide you toward an optimal implementation for your specific needs.
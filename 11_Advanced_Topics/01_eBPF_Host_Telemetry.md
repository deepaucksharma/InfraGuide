# eBPF & Host Telemetry

## Introduction

Extended Berkeley Packet Filter (eBPF) represents a revolutionary approach to kernel observability, enabling unprecedented visibility into system behavior with minimal overhead. This chapter explores how eBPF integrates with New Relic and other observability tools to provide deep insights into host performance, security, and application behavior.

## eBPF Fundamentals

eBPF allows programs to run safely within the Linux kernel, capturing detailed telemetry without modifying kernel code or loading custom modules. This powerful capability enables:

1. **Fine-grained performance analysis**: CPU, memory, I/O, and network behavior at process and system levels
2. **Security observability**: System call auditing, network activity monitoring, and anomaly detection
3. **Application tracing**: Request flows, latency breakdown, and dependency tracking without instrumentation
4. **Network observability**: Packet analysis, connection tracking, and protocol-specific insights

### eBPF Architecture

<!-- DG-58A: eBPF Probe Path Diagram -->

```mermaid
graph TD
    subgraph "User Space"
        A[eBPF Program Source]
        H[New Relic Infrastructure Agent]
        I[OTel Collector]
        J[Custom Applications]
    end
    
    subgraph "eBPF Runtime"
        B[eBPF Compiler]
        C[eBPF Verifier]
        D[eBPF VM]
    end
    
    subgraph "Attachment Points"
        E[Kprobes/Kretprobes]
        F[Tracepoints]
        G[XDP/TC]
        K[Uprobes/Uretprobes]
        L[Raw Tracepoints]
        M[Perf Events]
    end
    
    subgraph "Kernel Space"
        N[System Calls]
        O[Network Stack]
        P[File System]
        Q[Process Scheduler]
        R[Memory Management]
    end
    
    A -->|Compile| B
    B -->|Verify| C
    C -->|Load| D
    
    D -->|Attach| E
    D -->|Attach| F
    D -->|Attach| G
    D -->|Attach| K
    D -->|Attach| L
    D -->|Attach| M
    
    E --> N
    E --> O
    E --> P
    E --> Q
    E --> R
    
    F --> N
    F --> O
    F --> P
    F --> Q
    F --> R
    
    G --> O
    
    K --> J
    
    D -->|Maps/Ring Buffers| H
    D -->|Maps/Ring Buffers| I
    D -->|Maps/Ring Buffers| J
    
    style A fill:#bbf,stroke:#66f
    style H fill:#fbb,stroke:#f66
    style I fill:#fbb,stroke:#f66
    style J fill:#fbb,stroke:#f66
    
    style B fill:#fdb,stroke:#fa6
    style C fill:#fdb,stroke:#fa6
    style D fill:#fdb,stroke:#fa6
    
    style E fill:#bfb,stroke:#6f6
    style F fill:#bfb,stroke:#6f6
    style G fill:#bfb,stroke:#6f6
    style K fill:#bfb,stroke:#6f6
    style L fill:#bfb,stroke:#6f6
    style M fill:#bfb,stroke:#6f6
    
    style N fill:#ddf,stroke:#99f
    style O fill:#ddf,stroke:#99f
    style P fill:#ddf,stroke:#99f
    style Q fill:#ddf,stroke:#99f
    style R fill:#ddf,stroke:#99f
```

### Key eBPF Integration Points

| Attachment Point | Description | Observability Use Cases |
|------------------|-------------|-------------------------|
| **Kprobes/Kretprobes** | Dynamic instrumentation of kernel functions | • Syscall performance<br>• Disk I/O tracking<br>• Memory allocations |
| **Tracepoints** | Static instrumentation points in kernel | • Scheduler events<br>• Network events<br>• File system operations |
| **XDP/TC** | Network packet processing hooks | • DDoS protection<br>• Network traffic analysis<br>• Protocol-specific metrics |
| **Uprobes/Uretprobes** | User-space function instrumentation | • Application function profiling<br>• Library call analysis<br>• Custom application tracing |
| **Perf Events** | Hardware and software performance counters | • CPU performance monitoring<br>• Cache utilization<br>• Memory bus activity |
| **LSM (Linux Security Modules)** | Security enforcement hooks | • Security policy auditing<br>• Privilege escalation detection<br>• Container escape monitoring |

## New Relic eBPF Integration

New Relic's integration with eBPF technology provides enhanced visibility across three main areas:

### 1. Infrastructure Monitoring

The New Relic Infrastructure agent leverages eBPF to collect detailed system telemetry:

| Capability | Metrics Collected | Visualization |
|------------|-------------------|---------------|
| **Process Drill-Down** | • CPU usage by thread<br>• Memory allocation patterns<br>• File descriptor usage<br>• Syscall frequency | Process activity heatmap |
| **I/O Analysis** | • Per-process I/O operations<br>• Disk latency distribution<br>• File system cache effectiveness<br>• Block device saturation | I/O operation flame graphs |
| **Network Flow Visibility** | • Connection establishment rate<br>• Connection duration<br>• Throughput by process<br>• Protocol error rates | Network topology map |
| **Resource Contention** | • Lock contention<br>• CPU run queue latency<br>• Memory pressure indicators<br>• I/O wait analysis | Resource contention heat map |

### 2. Application Performance Enhancement

eBPF extends APM capabilities without requiring additional instrumentation:

<!-- DG-58B: eBPF APM Enhancement Flow -->

```mermaid
graph LR
    subgraph "Traditional APM"
        A1[Agent Instrumentation]
        A2[Code-level Visibility]
        A3[Transaction Tracing]
    end
    
    subgraph "eBPF Enhancement"
        B1[Syscall Tracing]
        B2[Network Traffic Analysis]
        B3[Runtime Behavior]
        B4[OS-Application Interaction]
    end
    
    subgraph "Enhanced Insights"
        C1[Complete Latency Attribution]
        C2[Hidden Dependency Discovery]
        C3[Resource Efficiency Analysis]
        C4[Security Context]
    end
    
    A1 --> C1
    A2 --> C1
    A3 --> C1
    
    B1 --> C1
    B1 --> C3
    
    B2 --> C1
    B2 --> C2
    
    B3 --> C3
    B4 --> C2
    B4 --> C4
    
    style A1 fill:#bbf,stroke:#66f
    style A2 fill:#bbf,stroke:#66f
    style A3 fill:#bbf,stroke:#66f
    
    style B1 fill:#fdb,stroke:#fa6
    style B2 fill:#fdb,stroke:#fa6
    style B3 fill:#fdb,stroke:#fa6
    style B4 fill:#fdb,stroke:#fa6
    
    style C1 fill:#bfb,stroke:#6f6
    style C2 fill:#bfb,stroke:#6f6
    style C3 fill:#bfb,stroke:#6f6
    style C4 fill:#bfb,stroke:#6f6
```

### 3. Security Observability

eBPF provides critical security telemetry for threat detection and analysis:

| Security Dimension | eBPF Data Points | Alert Indicators |
|--------------------|------------------|------------------|
| **Process Execution** | • Process creation events<br>• Binary execution<br>• Command line arguments<br>• Parent-child relationships | • Unusual binary execution paths<br>• Known malicious patterns<br>• Privilege escalation sequences |
| **File System Activity** | • File access patterns<br>• Permission changes<br>• Sensitive file operations<br>• File integrity indicators | • Access to sensitive configurations<br>• Unexpected permission changes<br>• Binary/configuration modifications |
| **Network Behavior** | • Connection establishment<br>• Data transfer volumes<br>• DNS queries<br>• Protocol anomalies | • Connections to suspicious IPs<br>• Unusual data transfer patterns<br>• Anomalous protocol behavior |
| **Container Boundaries** | • Namespace transitions<br>• Capability usage<br>• Mount operations<br>• Resource access patterns | • Container escape attempts<br>• Unexpected privileged operations<br>• Suspicious mount activity |

## Kubernetes Observability with eBPF

When deployed in Kubernetes environments, eBPF provides unique visibility into:

### Cluster-Level Performance

```mermaid
graph TD
    subgraph "Control Plane"
        A[API Server]
        B[etcd]
        C[Scheduler]
        D[Controller Manager]
    end
    
    subgraph "Node Components"
        E[kubelet]
        F[kube-proxy]
        G[Container Runtime]
        H[Node Resources]
    end
    
    subgraph "Workloads"
        I[Pod Resources]
        J[Inter-Pod Communication]
        K[Pod-Service Communication]
        L[Ingress/Egress Traffic]
    end
    
    subgraph "eBPF Telemetry"
        M[Control Plane Metrics]
        N[Node Performance]
        O[Network Flow Maps]
        P[Workload Interactions]
    end
    
    A --> M
    B --> M
    C --> M
    D --> M
    
    E --> N
    F --> N
    G --> N
    H --> N
    
    I --> P
    J --> O
    K --> O
    L --> O
    
    style M fill:#fbb,stroke:#f66
    style N fill:#fbb,stroke:#f66
    style O fill:#fbb,stroke:#f66
    style P fill:#fbb,stroke:#f66
    
    style A fill:#bbf,stroke:#66f
    style B fill:#bbf,stroke:#66f
    style C fill:#bbf,stroke:#66f
    style D fill:#bbf,stroke:#66f
    
    style E fill:#bfb,stroke:#6f6
    style F fill:#bfb,stroke:#6f6
    style G fill:#bfb,stroke:#6f6
    style H fill:#bfb,stroke:#6f6
    
    style I fill:#fdb,stroke:#fa6
    style J fill:#fdb,stroke:#fa6
    style K fill:#fdb,stroke:#fa6
    style L fill:#fdb,stroke:#fa6
```

### Pod-Level Network Visibility

eBPF provides detailed network flow mapping without service mesh overhead:

| Metric Category | Traditional Visibility | eBPF-Enhanced Visibility |
|-----------------|------------------------|--------------------------|
| **Connection Establishment** | Endpoints only | • Full TCP handshake timing<br>• Connection setup latency<br>• Retransmit patterns<br>• Connection tracking table visibility |
| **Traffic Analysis** | Volume metrics only | • Protocol breakdown<br>• Packet size distribution<br>• Header analysis<br>• Throughput vs. goodput |
| **Service Communication** | Black-box latency | • DNS resolution time<br>• TLS handshake duration<br>• HTTP header processing time<br>• Response generation time |
| **Error Detection** | Status codes only | • TCP retransmissions<br>• Silent packet drops<br>• Protocol errors<br>• Timeout root causes |

## Host Observability Beyond Metrics

### System Call Tracing and Analysis

System call tracing provides visibility into application-kernel interactions:

<!-- DG-58C: System Call Flow Map -->

```mermaid
graph LR
    subgraph "Application Space"
        A1[Web Server]
        A2[Database Client]
        A3[Cache Client]
    end
    
    subgraph "System Call Interface"
        B1[read/write]
        B2[socket/connect]
        B3[epoll/select]
        B4[futex]
    end
    
    subgraph "Kernel Subsystems"
        C1[File System]
        C2[Network Stack]
        C3[Process Management]
        C4[Memory Management]
    end
    
    A1 --> B1
    A1 --> B2
    A1 --> B3
    
    A2 --> B1
    A2 --> B2
    
    A3 --> B1
    A3 --> B4
    
    B1 --> C1
    B2 --> C2
    B3 --> C2
    B3 --> C3
    B4 --> C3
    B4 --> C4
    
    style A1 fill:#bbf,stroke:#66f
    style A2 fill:#bbf,stroke:#66f
    style A3 fill:#bbf,stroke:#66f
    
    style B1 fill:#fdb,stroke:#fa6
    style B2 fill:#fdb,stroke:#fa6
    style B3 fill:#fdb,stroke:#fa6
    style B4 fill:#fdb,stroke:#fa6
    
    style C1 fill:#bfb,stroke:#6f6
    style C2 fill:#bfb,stroke:#6f6
    style C3 fill:#bfb,stroke:#6f6
    style C4 fill:#bfb,stroke:#6f6
    
    classDef highest stroke-width:4px
    class B1,B2 highest
```

### Block I/O and File System Analysis

Detailed I/O analysis helps identify performance bottlenecks:

| I/O Dimension | Key Metrics | Visualization Technique |
|---------------|-------------|-------------------------|
| **Latency Distribution** | • Block I/O latency percentiles<br>• Queue time vs. device time<br>• Request size impact | Heatmap by operation size and type |
| **I/O Stack Breakdown** | • VFS layer time<br>• Block layer time<br>• Device driver time<br>• Hardware time | Stacked bar charts by layer |
| **I/O Patterns** | • Sequential vs. random access<br>• Read/write ratio<br>• Block size distribution<br>• IO depth | Access pattern visualization |
| **Cache Effectiveness** | • Page cache hit ratio<br>• Dirty page writeback rate<br>• Cache eviction pressure<br>• Read-ahead effectiveness | Time-series correlation with application latency |

### Memory Subsystem Visibility

eBPF provides deeper visibility into memory behavior:

```mermaid
graph TD
    subgraph "Application Memory Events"
        A1[Memory Allocation]
        A2[Memory Access]
        A3[Memory Deallocation]
    end
    
    subgraph "Kernel Memory Management"
        B1[Page Allocation]
        B2[Page Fault Handling]
        B3[Memory Reclamation]
        B4[Huge Page Management]
    end
    
    subgraph "Hardware Interaction"
        C1[TLB Operations]
        C2[Cache Line Activity]
        C3[NUMA Access Patterns]
        C4[Memory Bus Saturation]
    end
    
    A1 --> B1
    A2 --> B2
    A3 --> B3
    
    B1 --> C2
    B1 --> C3
    B2 --> C1
    B2 --> C2
    B3 --> C3
    B4 --> C1
    
    style A1 fill:#bbf,stroke:#66f
    style A2 fill:#bbf,stroke:#66f
    style A3 fill:#bbf,stroke:#66f
    
    style B1 fill:#fdb,stroke:#fa6
    style B2 fill:#fdb,stroke:#fa6
    style B3 fill:#fdb,stroke:#fa6
    style B4 fill:#fdb,stroke:#fa6
    
    style C1 fill:#bfb,stroke:#6f6
    style C2 fill:#bfb,stroke:#6f6
    style C3 fill:#bfb,stroke:#6f6
    style C4 fill:#bfb,stroke:#6f6
```

## Implementing eBPF-Enhanced Observability

### New Relic Implementation Options

| Implementation Approach | Description | Best For |
|-------------------------|-------------|----------|
| **Enhanced Infrastructure Agent** | Built-in eBPF capabilities in the standard New Relic Infrastructure agent | • General-purpose monitoring<br>• Broad system visibility<br>• Low operational complexity |
| **Custom eBPF Programs with NR Ingest** | Specialized eBPF programs sending telemetry to New Relic ingest endpoints | • Targeted deep analysis<br>• Custom security monitoring<br>• Specialized application visibility |
| **OTel Collector with eBPF Receiver** | OpenTelemetry collector with eBPF receiver sending to New Relic | • Standardized instrumentation<br>• Multi-destination telemetry<br>• Integration with existing OTel pipeline |

### eBPF Program Deployment Matrix

| Deployment Mechanism | Pros | Cons | Best For |
|----------------------|------|------|----------|
| **BCC (BPF Compiler Collection)** | • Full programming flexibility<br>• Python interface<br>• Access to all eBPF features | • Requires kernel headers<br>• Development complexity<br>• Higher resource usage | Custom observability solutions |
| **bpftrace** | • Simple one-liners<br>• Quick debugging<br>• DTrace-like syntax | • Limited programmatic capability<br>• Less performant for production use | Ad-hoc investigation |
| **libbpf + CO-RE (Compile Once – Run Everywhere)** | • Kernel version independence<br>• High performance<br>• Production ready | • C programming required<br>• Steeper learning curve | Production deployments |
| **Cilium/Hubble** | • Kubernetes-native<br>• Network policy enforcement<br>• Pre-built observability | • Primarily network-focused<br>• Requires specific CNI | Kubernetes environments |
| **New Relic eBPF integration** | • Zero configuration<br>• Managed lifecycle<br>• Automatic correlation | • Less customizable<br>• Limited to supported features | Enterprise monitoring |

### Performance and Overhead Considerations

eBPF programs introduce minimal but measurable overhead:

| Resource | Typical Overhead | Optimization Technique |
|----------|------------------|------------------------|
| **CPU** | 0.5-3% | • Limit probe frequency<br>• Use sampling where appropriate<br>• Optimize map access patterns |
| **Memory** | 50-200MB | • Control map sizes<br>• Limit event buffering<br>• Manage perf buffer sizes |
| **I/O** | Negligible | • Batch event processing<br>• Throttle event emission<br>• Filter events at source |
| **Network** | 1-5% of monitored traffic | • Control export frequency<br>• Apply early filtering<br>• Compress exported data |

## New Relic Visualization and Analysis

### Interactive System Topology

New Relic's visualization capabilities leverage eBPF data to create interactive system maps:

```mermaid
graph TD
    subgraph "Host: prod-app-01"
        A1[nginx]
        A2[app-server]
        A3[redis]
        A4[postgres]
        
        A1 -->|HTTP/100rps| A2
        A2 -->|GET/25rps| A3
        A2 -->|SELECT/35rps| A4
    end
    
    subgraph "Host: prod-app-02"
        B1[nginx]
        B2[app-server]
        B3[redis]
        B4[postgres]
        
        B1 -->|HTTP/120rps| B2
        B2 -->|GET/30rps| B3
        B2 -->|SELECT/42rps| B4
    end
    
    subgraph "Host: prod-db-01"
        C1[postgres-primary]
        
        A4 -->|Replication| C1
        B4 -->|Replication| C1
    end
    
    subgraph "External Services"
        D1[payment-api]
        D2[auth-service]
        
        A2 -->|API/15rps| D1
        A2 -->|AUTH/40rps| D2
        B2 -->|API/18rps| D1
        B2 -->|AUTH/45rps| D2
    end
    
    style A1 fill:#bbf,stroke:#66f
    style A2 fill:#bbf,stroke:#66f
    style A3 fill:#bbf,stroke:#66f
    style A4 fill:#bbf,stroke:#66f
    
    style B1 fill:#fdb,stroke:#fa6
    style B2 fill:#fdb,stroke:#fa6
    style B3 fill:#fdb,stroke:#fa6
    style B4 fill:#fdb,stroke:#fa6
    
    style C1 fill:#bfb,stroke:#6f6
    
    style D1 fill:#fbb,stroke:#f66
    style D2 fill:#fbb,stroke:#f66
```

### Advanced Analysis Techniques

| Analysis Type | Description | Visualization | NRQL Example |
|---------------|-------------|---------------|--------------|
| **Syscall Heatmaps** | Visualize system call patterns across processes | Time-based heatmap | `SELECT syscall, count(*) FROM SyscallSample FACET process, syscall TIMESERIES` |
| **Network Flow Analysis** | Map network connections and identify bottlenecks | Network topology with edge weights | `SELECT sum(bytes_sent) FROM NetworkSample FACET source_process, destination_process, destination_port` |
| **Latency Breakdown** | Detailed analysis of where time is spent | Stacked area chart | `SELECT average(latency_ns) FROM SyscallSample FACET syscall_group WHERE process = 'nginx' TIMESERIES` |
| **Resource Contention** | Identify processes competing for resources | Contention matrix | `SELECT count(*) FROM ContentionSample FACET waiting_process, holding_process WHERE resource_type = 'lock'` |

## Integration with Other Observability Signals

### Correlation with APM

eBPF data provides context to application performance:

<!-- DG-58D: eBPF-APM Correlation -->

```mermaid
graph TD
    subgraph "Application Performance (APM)"
        A1[Transaction Traces]
        A2[Span Events]
        A3[Error Events]
    end
    
    subgraph "eBPF Telemetry"
        B1[System Calls]
        B2[Network Activity]
        B3[Resource Usage]
        B4[I/O Operations]
    end
    
    subgraph "Correlation Points"
        C1[time]
        C2[process.pid]
        C3[hostname]
        C4[container.id]
    end
    
    A1 --> C1
    A1 --> C2
    A1 --> C3
    A1 --> C4
    
    B1 --> C1
    B1 --> C2
    B1 --> C3
    B1 --> C4
    
    B2 --> C1
    B2 --> C2
    B2 --> C3
    B2 --> C4
    
    style A1 fill:#bbf,stroke:#66f
    style A2 fill:#bbf,stroke:#66f
    style A3 fill:#bbf,stroke:#66f
    
    style B1 fill:#fdb,stroke:#fa6
    style B2 fill:#fdb,stroke:#fa6
    style B3 fill:#fdb,stroke:#fa6
    style B4 fill:#fdb,stroke:#fa6
    
    style C1 fill:#bfb,stroke:#6f6
    style C2 fill:#bfb,stroke:#6f6
    style C3 fill:#bfb,stroke:#6f6
    style C4 fill:#bfb,stroke:#6f6
```

### Cross-Signal Analysis

| Scenario | Signals to Correlate | Insights Gained |
|----------|----------------------|-----------------|
| **Slow Database Queries** | • APM database spans<br>• eBPF syscall latency<br>• eBPF disk I/O<br>• Infrastructure metrics | • Is slowness in application code, query execution, or I/O?<br>• Are there filesystem cache misses?<br>• Is there disk contention from other processes? |
| **Network Latency Issues** | • APM external service calls<br>• eBPF network flow data<br>• eBPF TCP state metrics<br>• Infrastructure network metrics | • Is latency in connection establishment or data transfer?<br>• Are there retransmits or packet drops?<br>• Is DNS resolution causing delays? |
| **Memory Pressure** | • APM memory metrics<br>• eBPF memory allocation events<br>• eBPF page fault tracking<br>• Infrastructure memory metrics | • Is the application allocating excessively?<br>• Is the kernel reclaiming memory aggressively?<br>• Are there specific allocation patterns causing fragmentation? |

## Building Custom eBPF Observability

### Case Study: Custom Latency Attribution

A custom eBPF program can provide detailed breakdowns of where time is spent in the system:

| Component | Measured Dimension | Value to Observability |
|-----------|-------------------|------------------------|
| **Application** | • Function execution time<br>• Lock contention<br>• Memory allocation patterns | Identify code-level bottlenecks |
| **Syscall Interface** | • System call latency<br>• Parameter patterns<br>• Error rates | Understand application-kernel boundary |
| **File System** | • VFS operations<br>• File access patterns<br>• Cache effectiveness | Optimize data access patterns |
| **Network Stack** | • Protocol processing time<br>• Buffer utilization<br>• Connection state transitions | Tune network parameters |
| **Block I/O** | • Queue time<br>• Device service time<br>• Request merging effectiveness | Optimize storage configuration |
| **Scheduler** | • Run queue latency<br>• Context switch overhead<br>• CPU affinity effects | Improve CPU utilization |

### Security Monitoring with eBPF

eBPF enables enhanced security observability:

| Security Dimension | eBPF Capability | Security Insight |
|--------------------|----------------|------------------|
| **Process Execution** | Track exec() syscalls with full command line | Detect unexpected process execution |
| **File Access** | Monitor open(), read(), write() with path information | Identify access to sensitive files |
| **Network Activity** | Track connect(), accept(), send(), and recv() calls | Detect unauthorized connections |
| **User Behavior** | Monitor setuid(), setgid(), and capability changes | Identify privilege escalation |
| **Container Boundaries** | Track namespace operations and privilege changes | Detect container escape attempts |

## Future Directions

eBPF technology continues to evolve rapidly:

| Emerging Capability | Description | Potential Impact |
|--------------------|-------------|------------------|
| **BTF (BPF Type Format)** | Kernel type information embedded in kernel image | Eliminates need for kernel headers, simplifying deployment |
| **BPF LSM (Linux Security Modules)** | Security policy enforcement with eBPF | More flexible and dynamic security monitoring |
| **BPF Iterators** | Efficient iteration over kernel objects | Lower-overhead enumeration of system state |
| **kprobe Multi** | Attach to multiple kprobe points with single program | More efficient system-wide tracing |
| **Sleepable BPF Programs** | Allow eBPF programs to sleep | More complex programs with I/O operations |

## Conclusion

eBPF technology transforms Linux observability by providing unprecedented visibility with minimal overhead. When integrated with New Relic:

1. **Deep System Insights**: Visibility into kernel, system calls, and hardware interactions
2. **Enhanced Correlation**: Connect application behavior to underlying system activity
3. **Reduced Blind Spots**: Observe previously hidden interactions between components
4. **Minimal Overhead**: Gain these insights with negligible performance impact

Organizations implementing eBPF-enhanced observability can identify subtle performance issues, detect security anomalies, and understand complex system behaviors that would otherwise remain hidden.

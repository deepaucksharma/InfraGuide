# Compliance Framework for Observability

## Introduction

Organizations operating under regulatory frameworks must ensure that their observability practices align with compliance requirements. This chapter outlines a comprehensive framework for implementing compliant observability with New Relic, addressing data governance, access controls, and audit capabilities.

## Regulatory Landscape for Observability Data

Observability data often contains information subject to various regulations:

1. **Personal Data**: Customer identifiers, IP addresses, user behaviors
2. **Operational Data**: System configurations, security states, access patterns
3. **Financial Data**: Transaction metrics, payment processing latencies
4. **Health Information**: Service performance affecting patient care systems

### Key Regulations Affecting Observability

```mermaid
mindmap
  root((Observability Compliance))
    Data Privacy
      GDPR
        ::icon(fa fa-globe)
        Right to be Forgotten
        Data Minimization
        Cross-border Transfer
      CCPA/CPRA
        ::icon(fa fa-building)
        Consumer Rights
        Data Inventory
      LGPD
        ::icon(fa fa-flag)
    Financial
      PCI-DSS
        ::icon(fa fa-credit-card)
        Cardholder Data
        Network Monitoring
      SOX
        ::icon(fa fa-balance-scale)
        Audit Trails
      GLBA
        ::icon(fa fa-bank)
    Healthcare
      HIPAA
        ::icon(fa fa-heartbeat)
        PHI Protection
        Audit Controls
      HITECH
        ::icon(fa fa-hospital)
    Industry-specific
      FedRAMP
        ::icon(fa fa-government)
      ISO 27001
        ::icon(fa fa-shield)
      SOC 2
        ::icon(fa fa-lock)
```

## Compliance Matrix for Observability Systems

<!-- TB-68A: Compliance Matrix -->

The following matrix maps regulatory requirements to specific observability implementation controls:

| Compliance Requirement | Applicable Regulations | New Relic Implementation | Verification Method |
|------------------------|------------------------|--------------------------|-------------------|
| **Data Classification** | GDPR, CCPA, HIPAA, PCI-DSS | • Attribute tagging<br>• Data dictionary<br>• PII detection | • Data inventory audit<br>• Classification report |
| **Data Residency** | GDPR, LGPD, FedRAMP | • Region selection<br>• Data center selection<br>• Edge collection | • Infrastructure verification<br>• Network flow analysis |
| **Retention Controls** | GDPR, SOX, HIPAA | • Retention policies<br>• Data lifecycle management<br>• Data purging | • Retention policy audit<br>• Data age verification |
| **Access Controls** | PCI-DSS, SOC 2, HIPAA | • RBAC implementation<br>• SSO integration<br>• User access reviews | • Access matrix review<br>• Permission testing<br>• User entitlement reports |
| **Audit Trails** | SOX, HIPAA, PCI-DSS | • Platform audit logs<br>• Query audit trails<br>• Configuration change tracking | • Audit log completeness test<br>• Chain of custody verification |
| **Encryption** | HIPAA, PCI-DSS, GDPR | • TLS/SSL for transmission<br>• Data encryption at rest<br>• Key management | • Encryption verification<br>• Certificate validation |
| **Anonymization** | GDPR, CCPA, HIPAA | • PII obfuscation<br>• Data masking<br>• Aggregation techniques | • Data sampling review<br>• Reconstruction testing |
| **Breach Notification** | GDPR, HIPAA, CCPA | • Anomaly detection<br>• Access alerting<br>• Unusual query monitoring | • Alert verification testing<br>• Response time measurement |
| **Data Subject Requests** | GDPR, CCPA, LGPD | • Data search capabilities<br>• Deletion workflows<br>• Export processes | • DSR response testing<br>• Process time measurement |
| **Vendor Management** | GDPR, SOC 2, HIPAA | • Vendor assessment<br>• NR compliance documentation<br>• Sub-processor tracking | • Vendor audit review<br>• Documentation completeness |

## Data Governance Architecture

Implementing proper data governance requires a structured approach across the telemetry pipeline:

```mermaid
flowchart TD
    subgraph "Governance Framework"
        direction TB
        A[Data Classification] --> B[Collection Controls]
        B --> C[Processing Controls]
        C --> D[Storage Controls]
        D --> E[Access Controls]
        E --> F[Retention Controls]
        F --> G[Deletion Controls]
    end
    
    subgraph "Implementation Layer"
        direction TB
        A1[Classification Metadata] --> B1[Collector Filters]
        B1 --> C1[Processors/Transformers]
        C1 --> D1[NRDB Policies]
        D1 --> E1[RBAC/User Management]
        E1 --> F1[Data Lifecycle Policies]
        F1 --> G1[Data Purge Automation]
    end
    
    A --> A1
    B --> B1
    C --> C1
    D --> D1
    E --> E1
    F --> F1
    G --> G1
    
    style A fill:#f9f,stroke:#333,stroke-width:1px
    style B fill:#bbf,stroke:#333,stroke-width:1px
    style C fill:#bbf,stroke:#333,stroke-width:1px
    style D fill:#bbf,stroke:#333,stroke-width:1px
    style E fill:#bbf,stroke:#333,stroke-width:1px
    style F fill:#bbf,stroke:#333,stroke-width:1px
    style G fill:#f9f,stroke:#333,stroke-width:1px
    
    style A1 fill:#f9f,stroke:#333,stroke-width:1px,stroke-dasharray: 5 5
    style B1 fill:#bbf,stroke:#333,stroke-width:1px,stroke-dasharray: 5 5
    style C1 fill:#bbf,stroke:#333,stroke-width:1px,stroke-dasharray: 5 5
    style D1 fill:#bbf,stroke:#333,stroke-width:1px,stroke-dasharray: 5 5
    style E1 fill:#bbf,stroke:#333,stroke-width:1px,stroke-dasharray: 5 5
    style F1 fill:#bbf,stroke:#333,stroke-width:1px,stroke-dasharray: 5 5
    style G1 fill:#f9f,stroke:#333,stroke-width:1px,stroke-dasharray: 5 5
```

## Data Classification Implementation

### Classification Taxonomy

Establish a consistent classification schema for all telemetry data:

| Classification Level | Description | Examples | Handling Requirements |
|----------------------|-------------|----------|----------------------|
| **Public** | Non-sensitive operational data | • Service uptime<br>• Public endpoint response times<br>• Open-source component versions | • Standard retention<br>• Broad access permitted<br>• No special handling |
| **Internal** | Business operational data | • Internal API performance<br>• Non-production environments<br>• Resource utilization | • Standard encryption<br>• Employee-only access<br>• Aggregation for external sharing |
| **Confidential** | Business sensitive data | • Customer metrics (anonymized)<br>• Business process performance<br>• Production configuration | • Enhanced access controls<br>• Limited data sharing<br>• Approval for access changes |
| **Restricted** | Regulated or sensitive data | • Transaction values<br>• User behavior patterns<br>• Health system performance | • Strict access controls<br>• Enhanced encryption<br>• Audit logging<br>• Limited retention |
| **Highly Restricted** | Personal or regulated data | • PII indicators<br>• IP addresses<br>• Session identifiers<br>• Payment processing data | • Maximum protection<br>• Masking/anonymization<br>• Minimal retention<br>• Full audit trail |

## PII Management

### PII Detection and Handling Matrix

| PII Type | Detection Method | Handling Technique | Implementation Approach |
|----------|------------------|-------------------|------------------------|
| Email Addresses | Pattern matching | Tokenization | • Replace with consistent token<br>• Preserve domain for business analysis |
| IP Addresses | Field identification | Partial masking | • Mask last octet<br>• Geographic aggregation |
| Financial Data | Semantic analysis | Complete removal | • Drop high-risk fields<br>• Replace with coarse categories |
| Session IDs | Field identification | One-way hashing | • Hash with consistent salt<br>• Preserve correlation capability |
| User IDs | Dictionary lookups | Pseudonymization | • Replace with consistent pseudonym<br>• Maintain separation from identity store |
| Geolocation | Pattern recognition | Generalization | • Reduce precision to city/region level<br>• Convert to categorical data |
| Timestamps | Field identification | Time-window bucketing | • Round to hour/day<br>• Convert to relative time periods |

### PII Handling Decision Tree

```mermaid
graph TD
    A[Identify Data Element] --> B{Contains PII?}
    B -->|Yes| C{Regulatory Scope?}
    B -->|No| D[Standard Processing]
    
    C -->|GDPR| E[Apply GDPR Controls]
    C -->|PCI-DSS| F[Apply PCI Controls]
    C -->|HIPAA| G[Apply HIPAA Controls]
    C -->|Multiple| H[Apply Most Restrictive]
    
    E --> I{Business Need?}
    F --> I
    G --> I
    H --> I
    
    I -->|Required Raw| J[Apply Strong Controls]
    I -->|Analytics Only| K[Anonymize]
    I -->|Correlation Only| L[Pseudonymize]
    I -->|Not Required| M[Remove Completely]
    
    J --> N[Document Justification]
    K --> O[Implement Technical Controls]
    L --> O
    M --> O
    
    N --> P[Implement Access Controls]
    P --> Q[Set Retention Policy]
    O --> Q
    
    style B fill:#f99,stroke:#f66,stroke-width:2px
    style C fill:#99f,stroke:#66f,stroke-width:2px
    style I fill:#9f9,stroke:#6f6,stroke-width:2px
```

## Access Control Framework

### Role-Based Access Control Matrix

| Role | Data Classification Access | Functional Capabilities | Administrative Rights |
|------|----------------------------|-------------------------|----------------------|
| **Executive** | • Public: Full<br>• Internal: Full<br>• Confidential: Aggregated<br>• Restricted: Dashboards only<br>• Highly Restricted: None | • View dashboards<br>• Run saved queries<br>• Access reports | • None |
| **Platform Admin** | • Public: Full<br>• Internal: Full<br>• Confidential: Full<br>• Restricted: Limited<br>• Highly Restricted: None | • All query capabilities<br>• Configure data sources<br>• Manage users and roles<br>• Define data classification | • User management<br>• System configuration<br>• Integration management |
| **DevOps Engineer** | • Public: Full<br>• Internal: Full<br>• Confidential: Limited<br>• Restricted: None<br>• Highly Restricted: None | • Query telemetry data<br>• Configure alerts<br>• Manage dashboards<br>• Deploy instrumentation | • Dashboard creation<br>• Alert configuration |
| **Security Analyst** | • Public: Full<br>• Internal: Full<br>• Confidential: Full<br>• Restricted: Partial<br>• Highly Restricted: Masked only | • Query security telemetry<br>• Configure security alerts<br>• Review audit logs<br>• Investigate incidents | • Security policy configuration<br>• Audit log access |
| **Compliance Officer** | • Public: Metadata only<br>• Internal: Metadata only<br>• Confidential: Metadata only<br>• Restricted: Metadata only<br>• Highly Restricted: Metadata only | • View data inventory<br>• Access compliance reports<br>• Review data classifications<br>• Monitor retention policies | • Compliance policy configuration<br>• Retention policy management |
| **Developer** | • Public: Full<br>• Internal: Full<br>• Confidential: Limited to own apps<br>• Restricted: None<br>• Highly Restricted: None | • Query app telemetry<br>• Create dashboards<br>• Configure basic alerts<br>• Implement instrumentation | • None |

### Workload Identity Model

```mermaid
graph TB
    subgraph "Application Identity"
        A1[Kubernetes Service Account]
        A2[Cloud IAM Role]
        A3[X.509 Certificate]
    end
    
    subgraph "Authentication"
        B1[OIDC Provider]
        B2[JWT Validation]
        B3[Certificate Authority]
    end
    
    subgraph "Authorization"
        C1[New Relic API Keys]
        C2[Data Access Policies]
        C3[Functional Permissions]
    end
    
    subgraph "Telemetry Access"
        D1[Metrics Pipeline]
        D2[Logs Pipeline] 
        D3[Traces Pipeline]
        D4[Events Pipeline]
    end
    
    A1 --> B1
    A2 --> B1
    A3 --> B3
    
    B1 --> C1
    B2 --> C2
    B3 --> C1
    
    C1 --> D1
    C1 --> D2
    C1 --> D3
    C1 --> D4
    
    C2 --> D1
    C2 --> D2
    C2 --> D3
    C2 --> D4
    
    C3 -.-> C1
    C3 -.-> C2
    
    style A1 fill:#bbf,stroke:#66f
    style A2 fill:#bbf,stroke:#66f
    style A3 fill:#bbf,stroke:#66f
    
    style B1 fill:#fbb,stroke:#f66
    style B2 fill:#fbb,stroke:#f66
    style B3 fill:#fbb,stroke:#f66
    
    style C1 fill:#bfb,stroke:#6f6
    style C2 fill:#bfb,stroke:#6f6
    style C3 fill:#bfb,stroke:#6f6
    
    style D1 fill:#fdb,stroke:#fa6
    style D2 fill:#fdb,stroke:#fa6
    style D3 fill:#fdb,stroke:#fa6
    style D4 fill:#fdb,stroke:#fa6
```

## Data Residency Implementation

### Regional Data Flow Architecture

```mermaid
graph TB
    subgraph "EU Region"
        EU_App[Application]
        EU_OTel[OTel Collector]
        EU_Edge[Edge Proxy]
        
        EU_App -->|Local Telemetry| EU_OTel
        EU_OTel -->|Filtered Data| EU_Edge
    end
    
    subgraph "US Region"
        US_App[Application]
        US_OTel[OTel Collector]
        US_NR[New Relic US]
        
        US_App -->|Local Telemetry| US_OTel
        US_OTel -->|Full Data| US_NR
    end
    
    subgraph "APAC Region"
        APAC_App[Application]
        APAC_OTel[OTel Collector]
        APAC_Edge[Edge Proxy]
        
        APAC_App -->|Local Telemetry| APAC_OTel
        APAC_OTel -->|Filtered Data| APAC_Edge
    end
    
    EU_Edge -->|Compliant Data| EU_NR[New Relic EU]
    APAC_Edge -->|Compliant Data| APAC_NR[New Relic APAC]
    
    style EU_App fill:#bbf,stroke:#66f
    style EU_OTel fill:#fdb,stroke:#fa6
    style EU_Edge fill:#fbb,stroke:#f66
    style EU_NR fill:#bfb,stroke:#6f6
    
    style US_App fill:#bbf,stroke:#66f
    style US_OTel fill:#fdb,stroke:#fa6
    style US_NR fill:#bfb,stroke:#6f6
    
    style APAC_App fill:#bbf,stroke:#66f
    style APAC_OTel fill:#fdb,stroke:#fa6
    style APAC_Edge fill:#fbb,stroke:#f66
    style APAC_NR fill:#bfb,stroke:#6f6
```

### Data Residency Decision Matrix

| Data Type | EU Requirements | US Requirements | APAC Requirements | Implementation Approach |
|-----------|----------------|-----------------|-------------------|------------------------|
| **Metrics** | • Store in EU<br>• No PII in dimensions | • No special requirements | • Varies by country<br>• Some require local storage | • Regional collectors<br>• Filter high-risk dimensions<br>• Deploy region-specific instances |
| **Logs** | • Store in EU<br>• Mask PII<br>• Data subject controls | • Sector-specific (HIPAA, etc.)<br>• Varies by state | • Strict localization in some countries<br>• Content filtering | • Regional log storage<br>• Edge processing for masking<br>• Field-level filtering |
| **Traces** | • Store in EU<br>• Strip PII from spans<br>• Limited retention | • No special requirements<br>• Sector-specific controls | • Local processing required<br>• Export controls in some regions | • Distributed tracing with regional boundaries<br>• Trace truncation at borders<br>• PII scrubbing before storage |
| **User Sessions** | • Explicit consent required<br>• Right to access/delete<br>• Minimize collection | • Opt-out mechanism<br>• Privacy policy disclosure | • Consent requirements<br>• Government access considerations | • Consent management integration<br>• Session data minimization<br>• Geographical routing logic |

## Audit and Evidence Collection

### Audit Capabilities Matrix

| Auditable Activity | Control Objective | New Relic Capability | Evidence Artifacts |
|--------------------|-------------------|----------------------|-------------------|
| **Data Access** | Prevent unauthorized access | • API key audit logs<br>• User access logs<br>• Query logs | • Access log reports<br>• User entitlement reviews<br>• Alert notifications |
| **Data Modification** | Maintain data integrity | • Configuration change tracking<br>• Dashboard version history<br>• NRQL mutation logs | • Change audit reports<br>• Configuration snapshots<br>• Rollback history |
| **Configuration Changes** | Control system behavior | • Alert configuration history<br>• Integration change logs<br>• User permission changes | • Change request records<br>• Before/after comparisons<br>• Approval workflows |
| **Retention Compliance** | Meet regulatory requirements | • Retention policy logs<br>• Data purge confirmations<br>• Data age metrics | • Retention compliance reports<br>• Purge execution logs<br>• Data inventory aging |
| **Security Events** | Detect potential breaches | • Authentication failures<br>• Unusual query patterns<br>• Administrative actions | • Security incident tickets<br>• Trend analysis reports<br>• Response time measurements |

### Evidence Collection Architecture

```mermaid
graph LR
    A[Telemetry Sources] --> B{Evidence Collection Points}
    B --> C[Platform Activity]
    B --> D[Data Access]
    B --> E[Configuration Changes]
    B --> F[User Actions]
    
    C --> G[Evidence Repository]
    D --> G
    E --> G
    F --> G
    
    G --> H{Evidence Utilization}
    H --> I[Compliance Reporting]
    H --> J[Audit Support]
    H --> K[Incident Investigation]
    H --> L[Continuous Monitoring]
    
    style A fill:#bbf,stroke:#66f
    style B fill:#fdb,stroke:#fa6
    style G fill:#bfb,stroke:#6f6
    style H fill:#fbb,stroke:#f66
```

## Compliance Implementation Checklist

<!-- RB-68A: Compliance Implementation Checklist -->

| Category | Implementation Task | Validation Method | Owner | Priority |
|----------|---------------------|-------------------|-------|----------|
| **Planning** | Complete data inventory and classification | Classification validation workshop | Data Governance Team | High |
| **Planning** | Document regulatory requirements by data type | Regulatory mapping review | Compliance Officer | High |
| **Planning** | Define compliance architecture and controls | Architecture review board | Security Architect | High |
| **Technical** | Configure data collection filters for PII | Sample data review | DevOps Team | Critical |
| **Technical** | Implement data residency controls | Network flow analysis | Cloud Team | Critical |
| **Technical** | Configure retention policies | Policy verification testing | Platform Team | High |
| **Technical** | Set up role-based access controls | Permission matrix validation | Security Team | Critical |
| **Technical** | Enable comprehensive audit logging | Log completeness testing | Platform Team | High |
| **Technical** | Implement data subject request workflows | Process testing | Data Governance Team | Medium |
| **Technical** | Configure data purge automation | Purge verification test | Platform Team | Medium |
| **Process** | Document compliance procedures | Procedure review | Compliance Officer | Medium |
| **Process** | Train teams on compliance requirements | Knowledge assessment | Training Team | Medium |
| **Process** | Establish regular compliance reviews | Audit calendar creation | Compliance Officer | Medium |
| **Process** | Create incident response procedure | Tabletop exercise | Security Team | High |
| **Validation** | Perform compliance readiness assessment | Gap analysis | External Auditor | High |
| **Validation** | Conduct penetration testing | Vulnerability assessment | Security Team | High |
| **Validation** | Complete data protection impact assessment | DPIA review | Privacy Officer | Medium |
| **Validation** | Run end-to-end compliance scenarios | Scenario validation | Compliance Officer | Medium |

## Regional Compliance Requirements

### GDPR-Specific Implementation

| GDPR Requirement | New Relic Implementation | Validation Method |
|------------------|--------------------------|-------------------|
| **Lawful Basis for Processing** | • Documentation of legitimate interest<br>• Technical controls mapping | • Legal review<br>• Control testing |
| **Data Minimization** | • Attribute filtering<br>• Sampling configurations<br>• PII scrubbing | • Data inventory review<br>• Sample analysis |
| **Storage Limitation** | • Retention policies<br>• Automated purging<br>• Data lifecycle management | • Retention testing<br>• Age verification |
| **Security of Processing** | • Encryption configurations<br>• Access controls<br>• Security monitoring | • Security assessment<br>• Control validation |
| **Right to Erasure** | • NRQL deletion capabilities<br>• Data subject workflow<br>• Identifier mapping | • DSR simulation<br>• Process validation |
| **Cross-Border Transfers** | • EU data center selection<br>• Standard contractual clauses<br>• Transfer impact assessment | • Data flow analysis<br>• Legal documentation review |

### PCI-DSS Specific Implementation

| PCI-DSS Requirement | New Relic Implementation | Validation Method |
|----------------------|--------------------------|-------------------|
| **Protect Cardholder Data** | • Field masking<br>• Data classification<br>• Restricted access | • Data sampling review<br>• Control testing |
| **Maintain Vulnerability Program** | • Security monitoring<br>• Integration with vulnerability management | • Monitoring validation<br>• Alert testing |
| **Strong Access Controls** | • Fine-grained RBAC<br>• MFA integration<br>• Least privilege design | • Access review<br>• Permission testing |
| **Network Monitoring** | • Network telemetry collection<br>• Baselines and anomaly detection | • Alert validation<br>• Coverage assessment |
| **Regular Testing** | • Telemetry verification<br>• Control validation procedures | • Test execution<br>• Results documentation |
| **Information Security Policy** | • Policy integration<br>• Observability governance | • Policy review<br>• Alignment validation |

## Compliance Monitoring Dashboard

The following key metrics should be included in a compliance monitoring dashboard:

| Metric Category | Key Metrics | Visualization Type | Alert Threshold |
|-----------------|------------|-------------------|-----------------|
| **Data Protection** | • PII detection count<br>• Masked field volume<br>• Filtering effectiveness | Line chart + Heatmap | • >0 unmasked PII<br>• >5% filtering bypass |
| **Access Control** | • Authentication failures<br>• Privileged access events<br>• Permission changes | Bar chart + Timeline | • >3 failures per user<br>• Unusual time patterns |
| **Data Lifecycle** | • Retention policy compliance<br>• Data age distribution<br>• Purge completion status | Gauge + Distribution | • >0% retention violations<br>• Failed purge jobs |
| **Audit Coverage** | • Logged events volume<br>• Audit log completeness<br>• Coverage by system | Completeness heatmap | • <95% coverage<br>• Log interruptions |
| **Regulatory Events** | • Data subject requests<br>• Compliance incidents<br>• Audit findings | Timeline + Counter | • >24h DSR response<br>• Open findings >30 days |

## Conclusion

Implementing a comprehensive compliance framework for observability requires balancing regulatory requirements with operational needs. By integrating compliance controls into the telemetry pipeline and establishing proper governance, organizations can maintain full observability while adhering to their regulatory obligations.

Key takeaways:

1. **Design for Compliance**: Embed compliance requirements into the observability architecture from the beginning
2. **Data Classification**: Implement consistent classification to drive appropriate controls
3. **Layered Controls**: Apply technical, administrative, and physical controls appropriate to data sensitivity
4. **Continuous Validation**: Regularly test and verify the effectiveness of compliance measures
5. **Documentation**: Maintain clear documentation of compliance controls for audit readiness

By following this framework, organizations can implement New Relic and other observability tools in ways that satisfy even the most stringent regulatory requirements.

# SLOs & Error Budgets

## Introduction

Service Level Objectives (SLOs) and Error Budgets provide a systematic approach to balancing reliability and innovation velocity. This chapter explores how to implement effective SLOs using New Relic's observability platform, with a focus on practical implementation patterns, measurement techniques, and organizational alignment.

## SLO Fundamentals

### The SLO Pyramid

```mermaid
graph TD
    A[Service Level Agreements\nSLAs] --> B[Service Level Objectives\nSLOs]
    B --> C[Service Level Indicators\nSLIs]
    C --> D[Telemetry\nMetrics, Events, Logs, Traces]
    
    style A fill:#f9a,stroke:#f66
    style B fill:#adf,stroke:#66c
    style C fill:#ad9,stroke:#6c6
    style D fill:#ddd,stroke:#999
```

### Key Definitions

| Term | Definition | Example |
|------|------------|---------|
| **SLA** | Contract defining consequences for service quality, usually with financial penalties | "99.9% monthly uptime or 10% refund" |
| **SLO** | Internal reliability target, stricter than SLAs | "99.95% monthly availability" |
| **SLI** | Measurement of service quality aspects | "% of HTTP requests with < 300ms latency" |
| **Error Budget** | Allowed amount of unreliability within SLO | "0.05% downtime = ~22 minutes per month" |

### SLO Types and Examples

| SLO Type | Description | Example SLIs | Target Range |
|----------|-------------|--------------|--------------|
| **Availability** | Service responds when needed | • Success rate (non-5xx responses)<br>• Health check pass rate | 99.0% - 99.99% |
| **Latency** | Service responds quickly enough | • Requests completed within threshold<br>• p95 response time | 95% - 99.9% |
| **Throughput** | Service handles required load | • Requests/second capability<br>• Successful transactions/minute | Varies by service |
| **Correctness** | Service responds with right data | • Data validation success rate<br>• Business logic error rate | 99.9% - 100% |
| **Freshness** | Data is sufficiently up-to-date | • Data age within threshold<br>• Update frequency met | Varies by use case |
| **Coverage** | Service handles expected scope | • % of required functionality available<br>• Geographic availability | 95% - 100% |

## Implementing SLOs in New Relic

### SLI Selection Framework

Choosing the right SLIs is critical for effective SLOs:

```mermaid
flowchart TD
    A[Identify User Journeys] --> B{Critical?}
    B -->|Yes| C[Map to Technical Services]
    B -->|No| D[Deprioritize]
    C --> E[Determine Key Metrics]
    E --> F{Direct User Impact?}
    F -->|Yes| G[Primary SLI Candidate]
    F -->|No| H[Secondary/Supporting SLI]
    G --> I[Validate Measurement Feasibility]
    H --> I
    I --> J{Can Measure?}
    J -->|Yes| K[Define SLO]
    J -->|No| L[Add Instrumentation]
    L --> K
    
    style A fill:#bbf,stroke:#66f
    style E fill:#bbf,stroke:#66f
    style G fill:#bfb,stroke:#6f6
    style K fill:#bfb,stroke:#6f6
```

### SLO Implementation Methods in New Relic

| Method | Description | Best For | Limitations |
|--------|-------------|----------|------------|
| **NRQL-based SLOs** | Use NRQL queries to define SLIs and calculate SLO attainment | • Custom business logic<br>• Complex conditions<br>• Composite metrics | • Requires careful query optimization<br>• Manual error budget calculation |
| **New Relic SLO Entity** | Native SLO creation using built-in UI and entity | • Standardized SLOs<br>• Error budget visualization<br>• Integrated alerts | • Less flexibility for complex metrics<br>• Limited to standard patterns |
| **APM Service Levels** | Service-specific SLOs directly in APM | • Application-focused SLOs<br>• Developer accessibility<br>• Quick implementation | • Limited to APM-instrumented services<br>• Standardized SLIs only |
| **Synthetic Monitoring SLOs** | SLOs based on synthetic monitor results | • End-user perspective<br>• Geographic variation<br>• Public-facing services | • Simulation, not real user traffic<br>• Limited transaction coverage |

### SLO Data Flow Architecture

<!-- DG-59A: SLO Data Flow Architecture -->

```mermaid
graph TD
    subgraph "Data Sources"
        A1[APM Transaction Data]
        A2[Browser Monitoring]
        A3[Synthetic Checks]
        A4[Custom Events]
        A5[Infrastructure Metrics]
    end
    
    subgraph "SLI Computation"
        B1[NRQL Aggregations]
        B2[SLI Service]
        B3[Ratio Calculations]
    end
    
    subgraph "SLO Management"
        C1[SLO Definitions]
        C2[Target Settings]
        C3[Time Window Configuration]
    end
    
    subgraph "Outputs"
        D1[SLO Dashboards]
        D2[Error Budget Alerts]
        D3[Burndown Visualizations]
        D4[Compliance Reports]
    end
    
    A1 --> B1
    A2 --> B1
    A3 --> B1
    A4 --> B1
    A5 --> B1
    
    B1 --> B3
    B2 --> B3
    
    B3 --> C1
    C1 --> C2
    C2 --> C3
    
    C3 --> D1
    C3 --> D2
    C3 --> D3
    C3 --> D4
    
    style A1 fill:#bbf,stroke:#66f
    style A2 fill:#bbf,stroke:#66f
    style A3 fill:#bbf,stroke:#66f
    style A4 fill:#bbf,stroke:#66f
    style A5 fill:#bbf,stroke:#66f
    
    style B1 fill:#fdb,stroke:#fa6
    style B2 fill:#fdb,stroke:#fa6
    style B3 fill:#fdb,stroke:#fa6
    
    style C1 fill:#bfb,stroke:#6f6
    style C2 fill:#bfb,stroke:#6f6
    style C3 fill:#bfb,stroke:#6f6
    
    style D1 fill:#fbb,stroke:#f66
    style D2 fill:#fbb,stroke:#f66
    style D3 fill:#fbb,stroke:#f66
    style D4 fill:#fbb,stroke:#f66
```

## Example SLO Implementations

### Common SLO Patterns

| Service Type | Recommended SLO | SLI Implementation | Target |
|--------------|-----------------|-------------------|--------|
| **API Service** | Availability | `percentage(count(*), WHERE statusCode < 500) FROM Transaction WHERE appName = 'API-Service'` | 99.9% |
| **API Service** | Latency | `percentage(count(*), WHERE duration < 0.5) FROM Transaction WHERE appName = 'API-Service'` | 95.0% |
| **Web Application** | Page Load Time | `percentage(count(*), WHERE duration < 3) FROM PageView WHERE appName = 'WebApp'` | 90.0% |
| **Database** | Query Performance | `percentage(count(*), WHERE duration < 0.1) FROM Transaction WHERE appName = 'Database' AND transactionType = 'Web'` | 99.0% |
| **Background Job** | Completion Rate | `percentage(count(*), WHERE statusCode = 'SUCCESS') FROM Transaction WHERE transactionType = 'Background'` | 99.5% |
| **Streaming System** | Freshness | `percentage(count(*), WHERE (now() - timestamp) < 60) FROM KafkaLagSample` | 99.0% |

### SLO Time Windows

| Window Type | Use Case | Pros | Cons |
|-------------|----------|------|------|
| **Calendar** | Matching billing cycles | • Aligns with SLAs<br>• Clear reporting boundaries | • Abrupt budget resets<br>• Encourages end-of-period risk |
| **Rolling** | Continuous improvement | • Smooth transitions<br>• Consistent incentives | • Complex calculations<br>• Harder to report |
| **Multi-window** | Balanced approach | • Short and long-term visibility<br>• Early warning capability | • Implementation complexity<br>• Multiple targets to track |

### Error Budget Visualization

<!-- DG-59B: Error Budget Burndown Chart -->

```mermaid
xychart-beta
    title "Error Budget Burndown - API Service (30 day window)"
    x-axis [Day 1, Day 5, Day 10, Day 15, Day 20, Day 25, Day 30]
    y-axis "Remaining Budget (%)" 100 --> 0
    bar [100, 92, 81, 72, 65, 48, 35]
    line [100, 90, 80, 70, 60, 50, 40]
    
```

*Legend: Blue bars represent remaining error budget. Red line represents ideal linear burn rate.*

## Error Budget Policies

### Error Budget Policy Framework

<!-- DG-59C: Error Budget Policy Decision Tree -->

```mermaid
graph TD
    A[Monitor Error Budget] --> B{Budget Status?}
    
    B -->|Healthy > 75%| C[Normal Operations]
    B -->|Warning 25-75%| D[Caution Mode]
    B -->|Critical < 25%| E[Conservation Mode]
    B -->|Depleted 0%| F[Freeze Mode]
    
    C --> G[Full Feature Velocity]
    
    D --> H[Prioritize Risk Reduction]
    D --> I[Review Deployment Frequency]
    
    E --> J[Reduce Change Rate]
    E --> K[Require Extra Review]
    E --> L[Focus on Reliability Features]
    
    F --> M[Stop Non-Critical Changes]
    F --> N[Incident Response Mode]
    F --> O[All-Hands Reliability Focus]
    
    style A fill:#bbf,stroke:#66f,stroke-width:2px
    style B fill:#bbf,stroke:#66f,stroke-width:2px
    
    style C fill:#bfb,stroke:#6f6
    style D fill:#fdb,stroke:#fa6
    style E fill:#fbb,stroke:#f66
    style F fill:#f88,stroke:#f33,stroke-width:2px
    
    classDef action fill:#eee,stroke:#999
    class G,H,I,J,K,L,M,N,O action
```

### Error Budget Policy Components

| Component | Description | Example |
|-----------|-------------|---------|
| **Budget Calculation** | How to measure consumed budget | `(1 - SLO attainment) / (1 - SLO target)` |
| **Measurement Window** | Time period for budget calculation | Rolling 30-day window |
| **Alerting Thresholds** | When to notify about budget status | • Warning: 50% consumed<br>• Critical: 75% consumed<br>• Depleted: 100% consumed |
| **Response Actions** | Required actions based on budget | • <25% used: Normal operations<br>• 25-75% used: Increased scrutiny<br>• >75% used: Only reliability improvements |
| **Escalation Path** | Who is notified at each threshold | • Warning: Team Lead<br>• Critical: Engineering Manager<br>• Depleted: CTO/VP Engineering |
| **Exemption Process** | How to handle exceptions | Documented approval process with justification |
| **Replenishment** | How budget is reset/restored | Automatic reset on calendar month boundary |

### Sample Error Budget Response Matrix

| Budget Status | Engineering Focus | Release Process | On-Call Response | Meeting Cadence |
|---------------|-------------------|----------------|------------------|-----------------|
| **Healthy (>75%)** | • Feature development<br>• Planned reliability work | • Normal CI/CD<br>• Self-service deployments | • Normal rotation<br>• Standard escalation | • Regular sprint planning<br>• Normal SLO review |
| **Warning (25-75%)** | • Critical features<br>• Increased reliability focus | • Increased testing<br>• Additional review gates | • Senior engineer shadow<br>• Lower escalation threshold | • Weekly budget review<br>• Risk analysis for features |
| **Critical (<25%)** | • Only high-value features<br>• Major reliability improvements | • Deployment windows<br>• Manual approval required | • Additional on-call staff<br>• Proactive monitoring | • Daily status checks<br>• Executive updates |
| **Depleted (0%)** | • Only reliability fixes<br>• Incident remediation | • Emergency changes only<br>• Executive sign-off | • All-hands support<br>• War room activated | • Daily incident response<br>• Post-mortem planning |

## Advanced SLO Patterns

### User Journey SLOs

Complex user journeys can be modeled as composite SLOs:

```mermaid
graph LR
    A[Browse Products] -->|Add to Cart| B[Shopping Cart]
    B -->|Checkout| C[Payment Processing]
    C -->|Complete| D[Order Confirmation]
    
    subgraph "User Journey: Purchase Flow"
        A
        B
        C
        D
    end
    
    style A fill:#bbf,stroke:#66f
    style B fill:#bbf,stroke:#66f
    style C fill:#bbf,stroke:#66f
    style D fill:#bbf,stroke:#66f
```

| Journey Step | SLO Type | Target | Weight in Composite |
|--------------|----------|--------|---------------------|
| Browse Products | Page Load Time < 2s | 95% | 10% |
| Browse Products | Search Results < 1s | 95% | 15% |
| Shopping Cart | Cart Update < 500ms | 99% | 20% |
| Payment Processing | Processing Time < 3s | 99.5% | 30% |
| Order Confirmation | E2E Success Rate | 99.9% | 25% |

### Tiered SLOs

Different user segments can have distinct SLO targets:

| User Tier | Latency SLO | Availability SLO | Justification |
|-----------|------------|------------------|---------------|
| **Premium** | 95% < 200ms | 99.99% | • Revenue impact<br>• Contractual requirements<br>• Strategic relationships |
| **Standard** | 90% < 500ms | 99.9% | • Majority of user base<br>• Reasonable expectations<br>• Cost-effective delivery |
| **Free Tier** | 85% < 1000ms | 99.5% | • Limited business impact<br>• Acceptable degradation<br>• Cost optimization |

### Maturity-Based SLO Implementation

| Maturity Stage | SLO Approach | Error Budget Usage | Organizational Integration |
|----------------|--------------|-------------------|----------------------------|
| **Initial** | • Basic uptime monitoring<br>• Simple availability SLOs | • Manual tracking<br>• Post-incident analysis | • Individual champions<br>• Limited visibility |
| **Defined** | • Standard latency and availability SLOs<br>• Consistent implementation | • Regular reporting<br>• Informal policies | • Team-level adoption<br>• Engineering awareness |
| **Managed** | • Custom SLIs for business metrics<br>• User journey mapping | • Automated tracking<br>• Formal error budget policies | • Engineering-wide practice<br>• Management buy-in |
| **Optimized** | • Business-aligned objectives<br>• Adaptive targets<br>• Predictive modeling | • Dynamic allocation<br>• Automated enforcement<br>• ML-driven forecasting | • Executive visibility<br>• Cross-functional alignment<br>• Cultural cornerstone |

## SLO Analytics and Optimization

### Comparative SLO Analysis

| Service | SLO Target | Actual Performance | Error Budget Used | Time to Exhaustion |
|---------|------------|-------------------|-------------------|-------------------|
| **API Gateway** | 99.9% | 99.97% | 30% | 21 days |
| **User Service** | 99.5% | 99.82% | 36% | 19 days |
| **Payment Service** | 99.95% | 99.92% | 60% | 12 days |
| **Product Catalog** | 99.8% | 99.91% | 45% | 16 days |
| **Recommendation Engine** | 99.0% | 99.7% | 30% | 21 days |

### SLO Impact Analysis

<!-- DG-59D: SLO Impact Analysis -->

```mermaid
graph LR
    subgraph "Incident Categories"
        A1[Planned Maintenance]
        A2[Infrastructure Failures]
        A3[Application Bugs]
        A4[Dependency Failures]
        A5[Configuration Changes]
    end
    
    subgraph "Error Budget Impact"
        B1[15% of Budget]
        B2[30% of Budget]
        B3[25% of Budget]
        B4[20% of Budget]
        B5[10% of Budget]
    end
    
    A1 --> B1
    A2 --> B2
    A3 --> B3
    A4 --> B4
    A5 --> B5
    
    style A1 fill:#bbf,stroke:#66f
    style A2 fill:#fbb,stroke:#f66
    style A3 fill:#fdb,stroke:#fa6
    style A4 fill:#bfb,stroke:#6f6
    style A5 fill:#fdf,stroke:#f6f
    
    style B1 fill:#ddd,stroke:#999
    style B2 fill:#ddd,stroke:#999
    style B3 fill:#ddd,stroke:#999
    style B4 fill:#ddd,stroke:#999
    style B5 fill:#ddd,stroke:#999
```

### SLO Optimization Techniques

| Technique | Description | Implementation | Benefits |
|-----------|-------------|----------------|----------|
| **Target Adjustment** | Tune SLO targets based on user impact data | • Analyze user behavior at different performance levels<br>• Correlate with business metrics | • More realistic targets<br>• Better alignment with user experience |
| **Seasonal Variation** | Adjust targets for known traffic patterns | • Define calendar-aware SLOs<br>• Set different targets by day/week/season | • More accurate budget consumption<br>• Better capacity planning |
| **Progressive SLOs** | Steadily increase targets as service matures | • Start with conservative targets<br>• Increase gradually with improvement | • Realistic initial goals<br>• Continuous improvement path |
| **Business-Weighted SLOs** | Weight SLO components by business impact | • Assign value to transactions<br>• Weight SLO calculations accordingly | • Focus on highest-impact reliability<br>• Better business alignment |

## Integrating SLOs with DevOps Practices

### CI/CD Integration

```mermaid
graph TD
    A[Code Commit] --> B[Automated Tests]
    B --> C[Build Artifact]
    C --> D{SLO Budget Check}
    
    D -->|Healthy Budget| E[Automated Deployment]
    D -->|Limited Budget| F[Manual Approval]
    D -->|Depleted Budget| G[Block Deployment]
    
    E --> H[Canary Deployment]
    F --> H
    
    H --> I[SLI Monitoring]
    I --> J{Canary Healthy?}
    
    J -->|Yes| K[Full Deployment]
    J -->|No| L[Automatic Rollback]
    
    style D fill:#bbf,stroke:#66f,stroke-width:2px
    style I fill:#bbf,stroke:#66f,stroke-width:2px
    style J fill:#bbf,stroke:#66f,stroke-width:2px
    
    style G fill:#fbb,stroke:#f66
    style L fill:#fbb,stroke:#f66
```

### DevOps Integration Matrix

| DevOps Phase | SLO Integration | Error Budget Application | New Relic Integration |
|--------------|----------------|--------------------------|----------------------|
| **Planning** | • SLO-aligned feature priorities<br>• Reliability work allocation | • Budget status influences work mix<br>• Risk assessment for features | • SLO dashboards in planning<br>• Historical trends analysis |
| **Development** | • SLO testing in development<br>• Pre-commit SLI validation | • Feature complexity limited by budget<br>• Technical debt prioritization | • Local testing with SLI validation<br>• Development environment monitoring |
| **Integration** | • SLO regression testing<br>• Performance impact analysis | • Reject changes that threaten SLOs<br>• Additional testing when budget low | • CI pipeline SLO validation<br>• Automated test reporting |
| **Deployment** | • Deployment velocity tied to budget<br>• Canary analysis with SLIs | • Progressive deployment gates<br>• Automatic rollback triggers | • Deployment markers<br>• Change tracking correlation |
| **Operations** | • SLO-based alerting<br>• Incident priority from SLO impact | • Incident response prioritization<br>• Problem management focus | • Incident correlation<br>• SLO impact visualization |
| **Feedback** | • SLO-based retrospectives<br>• Continuous target refinement | • Budget consumption analysis<br>• Reliability investment planning | • Long-term trend analysis<br>• Business impact correlation |

## Advanced Alerting with SLOs

### Multi-Signal Alerting Matrix

| Alert Type | Triggering Condition | Response Action | Target Audience |
|------------|---------------------|-----------------|-----------------|
| **Burn Rate Alert** | Budget consumption rate exceeds sustainable pace | • Investigate recent changes<br>• Prepare mitigation options | SRE Team |
| **Step Function Alert** | Sudden significant drop in SLI | • Immediate investigation<br>• Potential rollback | On-Call Engineer |
| **Forecast Alert** | Projected to exhaust budget before window end | • Review upcoming changes<br>• Increase testing requirements | Engineering Manager |
| **Recovery Time Alert** | Time to restore SLO compliance exceeds threshold | • Escalate incident response<br>• Activate additional resources | Incident Commander |
| **Comparative Alert** | Significant deviation from historical patterns | • Analyze pattern changes<br>• Look for environmental factors | Performance Engineer |

### Error Budget Forecast Visualization

<!-- DG-59E: Error Budget Forecast -->

```mermaid
xychart-beta
    title "Error Budget Forecast - Payment API"
    x-axis [Week 1, Week 2, Week 3, Week 4]
    y-axis "Error Budget %" 100 --> 0
    line [75, 54, 32, 10]
    line [75, 50, 25, 0]
    line [75, 65, 55, 45]
```

*Legend: Blue line represents actual consumption. Red line represents predicted consumption at current burn rate. Green line represents target consumption.*

## Operational Excellence with SLOs

### SLO Review Process

| Review Type | Frequency | Participants | Key Questions | Outputs |
|-------------|-----------|--------------|--------------|---------|
| **Tactical Review** | Weekly | SRE Team | • Current budget status?<br>• Recent incidents?<br>• Upcoming risks? | • Alert adjustments<br>• Short-term actions |
| **Engineering Review** | Monthly | SRE + Dev Teams | • Systemic reliability issues?<br>• Technical debt impact?<br>• Feature vs. reliability balance? | • Reliability initiatives<br>• Development guidelines |
| **Strategic Review** | Quarterly | Leadership + SRE | • Business impact of reliability?<br>• Resource allocation?<br>• Target adjustments needed? | • SLO target revisions<br>• Investment decisions |
| **Customer Impact Review** | Quarterly | Product + SRE + Support | • Customer complaints vs. SLOs?<br>• User satisfaction correlation?<br>• Revenue/retention impact? | • Product roadmap input<br>• SLI refinements |

### SLO Governance Model

```mermaid
graph TD
    subgraph "Executive Level"
        A1[SLO Strategy Council]
        A2[Resource Allocation]
        A3[Quarterly SLO Review]
    end
    
    subgraph "Management Level"
        B1[SLO Working Group]
        B2[Error Budget Governance]
        B3[Cross-team Coordination]
    end
    
    subgraph "Team Level"
        C1[Service Owners]
        C2[SRE Embedded Partners]
        C3[Daily SLO Monitoring]
    end
    
    A1 --> B1
    A2 --> B2
    A3 --> B3
    
    B1 --> C1
    B2 --> C2
    B3 --> C3
    
    C1 -.-> B1
    C2 -.-> B2
    C3 -.-> B3
    
    style A1 fill:#bbf,stroke:#66f
    style A2 fill:#bbf,stroke:#66f
    style A3 fill:#bbf,stroke:#66f
    
    style B1 fill:#fdb,stroke:#fa6
    style B2 fill:#fdb,stroke:#fa6
    style B3 fill:#fdb,stroke:#fa6
    
    style C1 fill:#bfb,stroke:#6f6
    style C2 fill:#bfb,stroke:#6f6
    style C3 fill:#bfb,stroke:#6f6
```

## Case Studies

### E-Commerce Platform SLO Implementation

| Challenge | Solution | Results |
|-----------|----------|---------|
| **Balancing feature velocity with reliability** | • Implemented tiered SLOs by service criticality<br>• Created error budget policies linked to deployment gates | • 42% reduction in customer-impacting incidents<br>• Maintained feature development velocity |
| **Black Friday preparedness** | • Seasonal SLOs with adjusted targets<br>• Error budget spending plans for peak periods | • 99.98% availability during peak sales period<br>• No major incidents during holiday season |
| **Microservice dependencies** | • Service-level SLOs with upstream/downstream awareness<br>• Dependency-weighted error budgets | • Better cross-team alignment<br>• Improved incident response coordination |

### Financial Services SLO Implementation

| Challenge | Solution | Results |
|-----------|----------|---------|
| **Regulatory compliance requirements** | • SLAs translated to stricter internal SLOs<br>• Compliance-focused SLI selection | • Met 100% of regulatory reporting requirements<br>• Simplified audit preparation |
| **Different customer tiers** | • Customer-segment specific SLOs<br>• Prioritized error budget spending | • Improved premium customer experience<br>• Optimized resource allocation |
| **Trading hour criticality** | • Time-of-day adaptive SLOs<br>• Market-hours weighted alerting | • 99.995% availability during trading hours<br>• Maintenance work safely scheduled |

## Conclusion

SLOs and error budgets provide a structured framework for balancing reliability and innovation. When implemented effectively with New Relic, they enable:

1. **Data-Driven Reliability**: Replace subjective reliability discussions with objective measurements
2. **Balanced Innovation**: Create a clear framework for managing the pace of change
3. **Business Alignment**: Connect technical metrics to user experience and business outcomes
4. **Cultural Transformation**: Build a shared language for engineering and business stakeholders

Organizations that adopt SLOs typically see measurable improvements in both system reliability and development velocity by focusing reliability investments where they matter most.

The next chapter explores trace correlation and exemplars, which provide deeper insights into the performance characteristics captured by SLOs.

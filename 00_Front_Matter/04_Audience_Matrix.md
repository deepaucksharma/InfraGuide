# Audience Matrix

This table maps practitioner roles to recommended sections of the report, helping readers navigate directly to content most relevant to their responsibilities.

| Role/Persona | Primary Sections | Secondary Sections | Key Artifacts |
|--------------|------------------|-------------------|--------------|
| **SRE / Platform Engineer** | Ch 1: Telemetry Theory<br>Ch 2: K8s Signal Taxonomy<br>Ch 4: Anti-Patterns<br>Ch 13: OTel in NR | Ch 24: Advanced Pipelines<br>Ch 29: Scalability & Tuning<br>Ch 41: Loki Integration | DG-2B: Control-plane flow<br>CF-10B: Flex YAML<br>CF-24A: WASM filter<br>LB-13A: OTTL filter lab |
| **DevOps Engineer** | Ch 9: Ingest Topology<br>Ch 10: Infra & Flex Agents<br>Ch 31-35: Hybrid Architecture | Ch 17: Agent Internals<br>Ch 38: Query Cookbook<br>Ch 58: eBPF & Host Telemetry | RB-10A: Agent troubleshooting<br>CF-32A: Dual-install Helm<br>CF-45A: DogStatsD config |
| **FinOps / Cost Optimization** | Ch 5: Data-Gravity Economics<br>Ch 49-52: Unified Cost & FinOps<br>Ch 54: Benchmark Tables | Ch 21: NRQL Cost<br>Ch 26: Low-Data Mode<br>Ch 34: Cost/Performance Matrix | EQ-0A: TCO model<br>TB-49B: Unit-cost derivation<br>CF-52A: Show-back dashboard |
| **Architect / Technical Lead** | Ch 31: Hybrid Decision Tree<br>Ch 53-57: Case Studies & ADRs<br>Ch 68: Compliance Matrix | Ch 19: NRDB Internals<br>Ch 44-47: Datadog Comparison<br>Ch 60: Trace Correlation | DG-31A: Decision DAG<br>DG-9A: Nine-plane diagram<br>TB-34A: Cost/performance matrix |
| **Database / Performance Eng.** | Ch 19: NRDB Column Store<br>Ch 21: NRQL Optimizer<br>Ch 38: Query Cookbook | Ch 17: Agent Runtime<br>Ch 39: Prometheus Internals<br>Ch 54: Benchmark Tables | DG-19B: Storage flow<br>TB-19C: Compression ratios<br>CF-21A: Query optimization |
| **Security / Compliance** | Ch 68: Compliance Frameworks<br>Ch 69: Security Configurations<br>Appendix D: Compliance Matrix | Ch 5: Data-Gravity<br>Ch 13: OTel Security<br>Ch 63: GitOps Patterns | RB-68A: Compliance checklist<br>TB-68A: Compliance matrix<br>CF-63D: Security configs |
| **Executive / Decision Maker** | Executive Abstract<br>Ch 34: Cost/Performance Matrix<br>Ch 55: Case Studies | Ch 5: Data-Gravity Economics<br>Ch 31: Decision Framework<br>Ch A: Glossary | DG-0A: Heat map<br>EQ-0A: TCO model<br>TB-54A: Performance summary |

## How to Use This Report

1. **For immediate tactical needs**: Locate your role in the matrix above and focus on the Primary Sections and Key Artifacts.

2. **For strategic planning**: Begin with the Executive Abstract, then explore Case Studies (Ch 53-57) and the Decision Framework (Ch 31).

3. **For implementation guidance**: The Lab-boxes (LB) and Run-books (RB) throughout provide hands-on, executable steps with validation criteria.

4. **For comparative analysis**: Use the Benchmark Tables (Ch 54) and Cost/Performance Matrix (Ch 34) to evaluate different approaches.

5. **For deep technical understanding**: Follow the chapter sequence within each part for a progressive build-up of concepts.

> **Note**: If your role spans multiple areas, consider the intersections of recommended sections. For example, a Platform Engineer with cost optimization responsibilities should prioritize Ch 4 (Anti-Patterns), Ch 26 (Low-Data Mode), and Ch 29 (Scalability).
---
description: Prevent unreviewed modifications to shared/production systems
version: "1.0.0"
scope: global
---

## Shared/Production Systems — Mandatory Confirmation

**NEVER modify live shared systems without explicit user approval**, even in auto mode. This includes:

- **Datasource configs** (Grafana, Prometheus, Loki, etc.) — changing auth, URLs, or credentials
- **Cloud provider settings** (IAM, secrets, service accounts, network rules)
- **CI/CD pipelines** (GitHub Actions, workflows, deploy configs)
- **Database operations** (migrations, schema changes, connection settings)
- **DNS / Load balancer / Ingress** configurations
- **Third-party integrations** (Slack webhooks, API keys, OAuth configs)
- **Monitoring & alerting** (contact points, notification channels, PagerDuty routes)

### What to do instead

1. **Stop** before making the change
2. **Explain** what you want to test, what could break, and the blast radius
3. **Propose alternatives** (e.g. query the API read-only, create a test copy, check config without modifying)
4. **Wait for explicit go-ahead** with full understanding of the risks

### Key principle

Read-only investigation is always safe. **Writes to shared systems are never safe to assume.**

- Querying a health endpoint = investigating (safe)
- Changing an auth method on a live datasource = modifying (requires confirmation)
- Reading terraform state = investigating (safe)
- Running terraform apply = modifying (requires confirmation)

If in doubt, explain the risk and ask:
> "I'd like to modify X, which could break Y for Z users. Here's what I recommend instead: ..."

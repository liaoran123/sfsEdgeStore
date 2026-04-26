# sfsEdgeStore Documentation

Lightweight Industrial IoT Edge Data Storage Adapter for EdgeX Foundry.

## Data Sovereignty & Compliance

**sfsEdgeStore is designed for data sovereignty.** All data is stored locally on the edge device - no cloud dependency, no data transfer to third parties.

- **GDPR Compliant**: Data never leaves your premises. No cross-border data transfer.
- **Local-First Architecture**: Complete offline operation with optional, controlled sync.
- **Encryption at Rest**: AES-256 database encryption ensures data is unreadable without your keys.
- **Zero Vendor Lock-in**: Pure Go binary, embedded database, no external services required.

This makes sfsEdgeStore ideal for industries with strict data residency requirements: manufacturing, energy, healthcare, and critical infrastructure.

## Table of Contents

### Getting Started
- [Quick Start](./quick-start.md) - Get up and running in 5 minutes
- [Installation](./installation.md) - Installation and deployment guide
- [First Deployment](./first-deployment.md) - Deploy in production

### Reference
- [API Reference](./api-reference.md) - REST API documentation
- [Configuration](./configuration.md) - All configuration options
- [CLI Commands](./cli-commands.md) - Command-line interface

### Architecture & Design
- [Architecture](./architecture.md) - System architecture overview
- [Data Flow](./data-flow.md) - How data moves through the system
- [Database Design](./database-design.md) - sfsDb integration details

### Operations
- [Monitoring & Alerting](./monitoring.md) - Monitor and configure alerts
- [Backup & Restore](./backup-restore.md) - Data backup procedures
- [Data Retention](./data-retention.md) - Configure data retention policies
- [Troubleshooting](./troubleshooting.md) - Common issues and solutions

### Security
- [Security](./security.md) - Authentication, TLS, and RBAC

### Development
- [Contributing](../CONTRIBUTING.md) - How to contribute
- [Code of Conduct](../CODE_OF_CONDUCT.md) - Community guidelines

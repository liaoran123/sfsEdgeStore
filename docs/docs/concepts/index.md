---
sidebar_position: 1
---

# Concepts

Explanations of sfsEdgeStore's architecture and design principles.

## Available Concepts

- [Edge Computing](edge-computing.md) - What is edge computing and why it matters
- [EdgeX Integration](edgex-integration.md) - How sfsEdgeStore integrates with EdgeX Foundry
- [Database Design](database-design.md) - sfsDb architecture and LevelDB internals
- [Security Model](security-model.md) - Authentication, TLS, and data protection
- [Licensing](licensing.md) - License model and edition differences

## Why Edge Storage?

Traditional IoT architectures send all data to the cloud for processing. Edge storage keeps data local, reducing:

- **Network bandwidth** - No need to stream all raw data
- **Latency** - Local processing is instant
- **Cloud costs** - Less data = lower storage costs
- **Downtime risk** - Works even without internet

## sfsEdgeStore Design Philosophy

1. **Minimal Resource Usage** - Designed for devices with as little as 128MB RAM
2. **No External Dependencies** - Single binary with embedded database
3. **EdgeX Native** - Full compatibility with EdgeX Foundry data format
4. **Production Ready** - Automatic retry, queue buffering, health monitoring

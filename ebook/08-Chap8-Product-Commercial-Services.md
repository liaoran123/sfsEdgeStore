# Chapter 8: Product Introduction and Commercial Services

## 8.1 sfsEdgeStore Product Lineup

After reading this book, you have several options for using sfsEdgeStore in your projects. This chapter explains the different product versions and commercial services available.

### Product Comparison

| Feature | Community (Free) | Professional | Enterprise |
|---------|-----------------|--------------|-------------|
| **Price** | Free | $29/year | $99/year |
| **License** | AGPL-3.0 | Commercial | Commercial |
| **Support** | GitHub Issues | Email (48h) | Priority (4h) |
| **Updates** | Community | Latest | Latest + Backports |
| **MQTT Client** | ✅ | ✅ | ✅ |
| **Local Storage** | ✅ | ✅ | ✅ |
| **HTTP API** | ✅ | ✅ | ✅ |
| **Data Queue** | ✅ | ✅ | ✅ |
| **Authentication** | Basic | Advanced | Advanced + SSO |
| **Encryption** | ❌ | ✅ AES-256 | ✅ AES-256 + Key Rotation |
| **Cloud Sync** | ❌ | ✅ Basic | ✅ Advanced |
| **Plugins** | ❌ | ✅ 3 included | ✅ Unlimited |
| **SLA** | None | 99.5% | 99.9% |

### Community Version (Free)

The community version is fully functional and suitable for:

- **Learning**: Study the source code, understand edge storage
- **Prototyping**: Build PoCs and demos
- **Small projects**: Non-critical personal or hobby projects
- **Evaluation**: Test before purchasing commercial license

**Getting Started with Community Version:**

```bash
# Clone from GitHub
git clone https://github.com/liaoran123/sfsEdgeStore.git
cd sfsEdgeStore

# Build
go build -o sfsedgestore main.go

# Run
./sfsedgestore
```

**License:** AGPL-3.0 (requires you to open-source derivative works)

### Professional Version ($29/year)

Ideal for professional developers and small teams:

**Includes:**
- Commercial license (no AGPL obligations)
- Email support with 48-hour response
- Access to professional plugins:
  - Cloud Sync Plugin
  - Advanced Alert Plugin
  - Data Analytics Plugin
- Monthly updates and security patches

**Purchase:** [Gumroad Store](https://gumroad.com/sfsedgestore)

### Enterprise Version ($99/year)

For organizations requiring enterprise-grade support:

**Includes:**
- Everything in Professional
- Priority support with 4-hour response
- Unlimited plugin access
- Custom plugin development consultation
- 99.9% uptime SLA
- Dedicated Slack channel
- Quarterly security audits

**Contact:** [sales@sfsedgestore.com](mailto:sales@sfsedgestore.com)

## 8.2 Licensing Models

### Understanding Open Source vs Commercial

**Open Source (AGPL-3.0):**
- Free to use and modify
- Must open-source derivative works
- No warranty or support
- Suitable for learning and non-commercial use

**Commercial License:**
- No obligation to open-source
- Includes support and updates
- Legal protection for your business
- Professional peace of mind

### When You Need a Commercial License

| Scenario | License Required |
|----------|-----------------|
| Personal learning | AGPL-3.0 (Community) |
| Open source project | AGPL-3.0 (Community) |
| Internal business tools | Commercial |
| Embedded in commercial product | Commercial |
| SaaS offering using sfsEdgeStore | Commercial |
| Resale with your product | Commercial |

### How Licensing Works

**1. Purchase:**
```bash
# Visit Gumroad store
# Select your plan
# Complete purchase (PayPal, Credit Card)
# Receive license key immediately
```

**2. Activation:**
```bash
# Configure license
./sfsedgestore --license-key "your-license-key"

# Or in config.json
{
  "LicenseKey": "your-license-key"
}
```

**3. Verification:**
```bash
# Check license status
curl http://localhost:8081/api/license/status

# Response:
{
  "valid": true,
  "plan": "professional",
  "expires": "2027-01-01",
  "features": ["encryption", "cloud-sync", "plugins"]
}
```

## 8.3 Technical Support Options

### Community Support (Free)

For community version users:

- **GitHub Issues**: [Report bugs](https://github.com/liaoran123/sfsEdgeStore/issues)
- **Documentation**: This book and [docs folder](https://github.com/liaoran123/sfsEdgeStore/tree/main/docs)
- **Response Time**: Variable (best effort)

### Professional Support

For Professional license holders:

**What's Included:**
- Email support: [support@sfsedgestore.com](mailto:support@sfsedgestore.com)
- Response within 48 hours
- Bug fixes and patches
- Configuration assistance
- Performance tuning guidance

**How to Get Help:**

```bash
# 1. Check documentation (Chapter 5)
# 2. Search existing GitHub issues
# 3. Collect diagnostics:
mkdir -p diagnostics
cp /etc/sfsedgestore/config.json diagnostics/
journalctl -u sfsedgestore > diagnostics/logs.txt
ps aux > diagnostics/process.txt
free -h > diagnostics/memory.txt
tar -czf diagnostics.tar.gz diagnostics/

# 4. Email with diagnostic bundle
# Subject: [Professional] Brief problem description
# Attach: diagnostics.tar.gz
```

### Enterprise Support

For Enterprise license holders:

**What's Included:**
- Priority email/slack support
- 4-hour response SLA
- Dedicated technical account manager
- Quarterly architecture reviews
- Emergency escalation path
- On-site support (optional, additional cost)

**Getting Started:**

```bash
# 1. Receive welcome email with Slack invite
# 2. Schedule kickoff call (30 min)
# 3. Share your infrastructure details
# 4. Establish communication channels
# 5. Set up monitoring integration
```

## 8.4 Custom Development Services

Need something specific? We offer custom development:

### Plugin Development

**Examples:**
- Custom data transformation plugins
- Integration with proprietary systems
- Specialized alert handlers
- Custom authentication providers

**Process:**
1. Requirements gathering (free consultation)
2. Proposal and estimate
3. Development and testing
4. Delivery and documentation

**Pricing:** Starting at $2,000 (varies by complexity)

### System Integration

**Examples:**
- EdgeX Foundry custom configurations
- Enterprise system integrations (SAP, Oracle, etc.)
- Cloud platform connections (AWS IoT, Azure IoT Hub, GCP)
- Legacy system modernization

**Pricing:** Starting at $5,000 (varies by scope)

### Training and Workshops

**Options:**
- 1-day hands-on workshop: $1,500
- 3-day deep dive: $4,000
- Remote training (up to 10 participants): $2,000

**Topics:**
- sfsEdgeStore architecture
- Performance optimization
- Production deployment
- Troubleshooting techniques

### Architecture Consulting

**For complex deployments:**

- System design review
- Performance assessment
- Scalability planning
- Security audit

**Pricing:** $200/hour (minimum 4 hours)

## 8.5 Consulting Services

Beyond technical support, we offer strategic consulting:

### Edge Computing Strategy

**What's Included:**
- Assessment of current infrastructure
- Edge computing opportunity identification
- Architecture recommendations
- Migration planning
- ROI analysis

**Deliverables:**
- Technical report (20-30 pages)
- Architecture diagrams
- Implementation roadmap
- Risk assessment

**Investment:** Starting at $3,000

### Data Architecture Review

**For organizations processing large volumes of IoT data:**

- Data flow analysis
- Storage optimization
- Query performance review
- Retention policy design

**Deliverables:**
- Current state assessment
- Bottleneck identification
- Optimization recommendations
- Implementation plan

**Investment:** Starting at $2,500

## 8.6 Success Stories

### Case 1: Smart Factory (50 Edge Gateways)

**Challenge:**
- 50 Raspberry Pi-based gateways collecting sensor data
- Frequent memory overflow issues with default storage
- Data loss during network interruptions
- 4+ hours weekly maintenance time

**Solution:**
- Deployed sfsEdgeStore Professional on all gateways
- Implemented data queue for fault tolerance
- Configured 30-day retention

**Results:**
| Metric | Before | After |
|--------|--------|-------|
| Memory Usage | 150 MB | 25 MB |
| Data Loss | 2.3% | 0% |
| Maintenance | 4 hrs/week | 30 min/week |
| Query Response | 500ms | 10ms |

### Case 2: Building Management System

**Challenge:**
- 200 sensors across 5 buildings
- Need real-time HVAC monitoring
- Historical data for 2 years
- Limited IT resources

**Solution:**
- sfsEdgeStore Enterprise on edge devices
- Cloud sync for central monitoring
- Custom alert integration with building management

**Results:**
- 20% energy savings from real-time optimization
- 99.9% data capture rate
- 50% reduction in HVAC-related complaints

### Case 3: Agricultural IoT Platform

**Challenge:**
- 1000 sensors across farms
- Harsh environments with intermittent connectivity
- Need to store 5 years of data
- Limited budget

**Solution:**
- Community version with custom data compression
- Solar-powered edge gateways
- Satellite connectivity fallback

**Results:**
- 99.5% uptime despite harsh conditions
- 60% storage cost reduction vs cloud-only
- Data available even during connectivity outages

## 8.7 Getting Started

### For Technical Users

1. **Start with Community** - Download, build, and test
2. **Read This Book** - Chapters 1-5 cover production deployment
3. **Join Community** - GitHub discussions, issues
4. **Evaluate** - Run in your environment for 30 days

### For Decision Makers

1. **Contact Sales** - Schedule a demo call
2. **Proof of Concept** - We'll help you evaluate
3. **Pilot Program** - Deploy to a subset of devices
4. **Full Rollout** - Scale with our support

### Contact Information

**Sales:**
- [sales@sfsedgestore.com](mailto:sales@sfsedgestore.com)
- [Gumroad Store](https://gumroad.com/sfsedgestore)

**Support:**
- [support@sfsedgestore.com](mailto:support@sfsedgestore.com)
- [GitHub Issues](https://github.com/liaoran123/sfsEdgeStore/issues)

**Community:**
- [GitHub Discussions](https://github.com/liaoran123/sfsEdgeStore/discussions)
- [Reddit](https://reddit.com/r/sfsedgestore)

## 8.8 Thank You

Thank you for reading this book. We hope it has provided you with the knowledge to successfully deploy sfsEdgeStore in your edge computing projects.

**Remember:**

> "The best edge storage solution is the one that works reliably in your specific environment. sfsEdgeStore is designed to be that solution, but your understanding of the technology is what makes the difference."

### Next Steps

**For Community Users:**
- Star the GitHub repo
- Share your feedback
- Contribute improvements

**For Potential Customers:**
- Start a free evaluation
- Schedule a technical demo
- Request a custom quote

**For Everyone:**
- Deploy in production
- Share your success story
- Join our growing community

---

**Good luck with your edge computing journey! 🚀**

---

➡️ Next: [Appendix A: Complete API Reference](../ebook/Appendix-A-API-Reference.md)

➡️ Back to [Table of Contents](../ebook/00-Preface-and-Table-of-Contents.md)
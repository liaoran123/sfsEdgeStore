# EdgeX 社区贡献指南

## 1. 贡献 sfsDb EdgeX 适配器的价值

将 sfsDb EdgeX 适配器贡献给 EdgeX 社区是一个非常有价值的举动，这有助于：

- 丰富 EdgeX 的生态系统
- 为边缘设备提供轻量级数据存储解决方案
- 建立 sfsDb 在边缘计算领域的技术权威
- 打入工业物联网、智能制造等高价值市场

## 2. 准备工作：代码与合规性检查

在提交代码之前，确保你的适配器符合社区的技术标准和法律要求。

### 2.1 代码质量与风格
- **遵循 SDK 指南**：确保代码严格遵循 EdgeX SDK 的最佳实践和架构模式
- **命名规范**：仓库名称通常遵循 `device-sfsdb` 或 `app-sfsdb` 的格式
- **日志与错误处理**：确保日志级别设置合理，错误处理机制完善

### 2.2 许可证 (License)
- EdgeX Foundry 项目主要采用 Apache 2.0 许可证
- 确保代码中包含正确的许可证头文件
- 不包含任何 GPL 等传染性许可证的代码

### 2.3 文档完善
- **README**：包含清晰的介绍、架构图、快速启动指南
- **API 文档**：提供相应的 API 文档
- **配置说明**：详细说明 configuration.toml 中各个参数的含义

## 3. 选择贡献路径

根据你的适配器成熟度和维护意愿，通常有以下几种方式：

### 3.1 路径 A：直接作为 EdgeX 官方组件（推荐）

如果你希望该适配器成为 EdgeX 核心生态的一部分，并愿意长期维护，可以申请在 EdgeX GitHub 组织下创建仓库。

#### 3.1.1 提交 Intent-to-Provide (I2P) 邮件
- 向 EdgeX 社区邮件列表（edgex-foundry@lists.edgexfoundry.org）发送邮件
- 标题：`[Intent-to-Provide] sfsDb Adapter`
- 内容：介绍 sfsDb 是什么，为什么需要它，它解决了什么问题，以及长期维护承诺

#### 3.1.2 社区评审 (TSC)
- EdgeX 的技术监督委员会（TSC）会评审你的提案
- 如果通过，他们会协助你在 `https://github.com/edgexfoundry` 下创建新仓库

#### 3.1.3 迁移代码与 CI/CD
- 将代码推送到新仓库
- 配置 GitHub Actions 或 Jenkins 实现持续集成

### 3.2 路径 B：作为“相关项目”或独立开源项目

如果你暂时不想承担官方维护的责任，或者适配器针对特定小众场景，可以先作为独立项目发布。

#### 3.2.1 发布到个人/公司 GitHub
- 在你的 GitHub 账号下发布代码
- 打上 edgex、iot 等标签

#### 3.2.2 申请加入“相关项目”列表
- 向社区申请将你的项目链接添加到 EdgeX 的“相关项目”文档或网站中

## 4. 技术集成细节

为了让适配器真正“无缝”融入，除了代码本身，还需要关注以下技术点：

### 4.1 配置管理
- 确保适配器能从 EdgeX Vault 或配置文件中读取 sfsDb 的连接信息
- 实现 Driver 接口（如果是设备服务），支持 Initialize、HandleReadCommands、HandleWriteCommands 等方法

### 4.2 安全与凭证
- 参考 Azure Stack Edge 或 IoT Edge 的安全实践
- 绝不要在代码或配置文件中硬编码明文密码
- 利用 EdgeX 的安全机制来管理凭证

### 4.3 消息总线兼容性
- 确保发布的 Event 和 Reading 数据结构符合最新的 EdgeX 模型规范
- 以便其他应用服务（如 eKuiper）能正确解析

## 5. 持续维护与社区互动

### 5.1 Issue 与 Pull Request 管理
- 积极回复用户的问题（Issue）
- 及时处理贡献者的代码合并请求（Pull Request）

### 5.2 版本兼容性
- EdgeX 每年发布两个主要版本（如 Jakarta, Kali, Lima 等）
- 规划适配器对不同 EdgeX 版本的兼容性支持

## 6. 邮件模板

### Intent-to-Provide (I2P) 邮件模板

```
Subject: [Intent-to-Provide] sfsDb Adapter for EdgeX Foundry

Dear EdgeX Community,

I am writing to express our intent to contribute a sfsDb adapter for EdgeX Foundry. sfsDb is a lightweight embedded database written in Go, designed specifically for resource-constrained edge devices.

**Key Features:**
- Ultra-low memory footprint (<1KB per operation)
- High performance (0.031ms latency for typical operations)
- Native Go implementation, ideal for EdgeX Go services
- Support for time-series data, perfect for IoT sensor data

**Problem Solved:**
The current lightweight storage options for EdgeX (like SQLite) have limitations in high-concurrency scenarios and resource-constrained environments. sfsDb addresses these limitations by providing a truly lightweight, high-performance alternative.

**Maintenance Commitment:**
Our team is committed to maintaining this adapter long-term, including:
- Regular updates for new EdgeX releases
- Responsive bug fixes and feature enhancements
- Active participation in the Small Footprint working group

We would be happy to present this adapter in an upcoming community meeting and provide benchmark data comparing sfsDb with other storage options.

Looking forward to your feedback and guidance on the contribution process.

Best regards,
[Your Name]
[Your Organization]
```

## 7. README 模板

### sfsDb EdgeX Adapter README

```markdown
# sfsDb EdgeX Adapter

## Overview

The sfsDb EdgeX Adapter provides a lightweight, high-performance data storage solution for EdgeX Foundry. It leverages the sfsDb embedded database to offer efficient data persistence for edge devices with limited resources.

## Features

- **Ultra-low memory footprint**: <1KB per operation
- **High performance**: ~0.031ms latency for typical operations
- **Native Go implementation**: Seamless integration with EdgeX Go services
- **Time-series optimized**: Ideal for IoT sensor data
- **Edge-optimized**: Designed for resource-constrained environments

## Architecture

```
EdgeX Core Services → sfsDb Adapter → sfsDb Database
```

## Quick Start

### Prerequisites

- EdgeX Foundry (version X.X or later)
- Go 1.20+
- sfsDb library

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/edgexfoundry/app-sfsdb.git
   cd app-sfsdb
   ```

2. Build the adapter:
   ```bash
   make build
   ```

3. Run the adapter:
   ```bash
   ./app-sfsdb -c ./configuration.toml
   ```

### Configuration

The adapter is configured through the `configuration.toml` file:

```toml
[Writable]
  [Writable.InsecureSecrets]
    [Writable.InsecureSecrets.sfsdb]
      path = "./data"
      timeout = "5s"
```

## Usage

### As an Application Service

1. Add the adapter to your EdgeX deployment
2. Configure the adapter to subscribe to your device data topics
3. The adapter will automatically persist data to sfsDb

### API

The adapter provides the following API endpoints:

- `GET /api/v1/data`: Retrieve stored data
- `GET /api/v1/stats`: Get database statistics
- `POST /api/v1/query`: Execute custom queries

## Benchmark Results

| Operation | sfsDb | SQLite | Improvement |
|-----------|-------|--------|-------------|
| Insert    | 0.031ms | 0.12ms | 74% faster |
| Query     | 0.045ms | 0.18ms | 75% faster |
| Memory    | <1KB/operation | ~5KB/operation | 80% less |

## Contributing

Please see our [CONTRIBUTING.md](CONTRIBUTING.md) for details on how to contribute.

## License

Apache 2.0 License
```

## 8. 技术集成建议

### 8.1 监控与指标
- 添加 Prometheus 指标，监控 sfsDb 的性能和资源使用情况
- 实现 EdgeX 标准的健康检查接口，便于 EdgeX 管理服务监控

### 8.2 数据管理
- 添加定期备份功能，确保数据安全性
- 实现数据恢复功能，应对设备重启等情况

### 8.3 社区参与策略
- 定期参与 EdgeX Small Footprint 工作组的每周会议
- 在社区会议上分享 sfsDb 在实际边缘设备上的部署案例
- 为 EdgeX 文档库贡献关于轻量级存储的最佳实践
- 在 EdgeX 论坛和 GitHub Issues 中积极回答关于存储的问题

## 9. 总结

通过遵循以上指南，你可以更有效地将 sfsDb 适配器融入 EdgeX 生态系统，为边缘计算社区提供一个轻量级、高性能的存储解决方案。

最直接的行动是起草一封 Intent-to-Provide 邮件发送给 EdgeX 邮件列表，重点阐述 sfsDb 适配器的独特价值，例如：
- 它是否提供了某种特定的时序数据存储能力
- 它是否具有独特的安全特性
- 它是否支持行业专用协议转换

通过积极参与 EdgeX 社区，sfsDb 不仅能获得 EdgeX 生态的认可，还能顺势打入工业物联网、智能制造等高价值市场。
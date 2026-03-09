# 性能测试指南

本文档说明如何运行 sfsEdgeStore 的性能基准测试，以及如何复现 README.md 中展示的性能数据。

---

## 运行基准测试

### 运行所有基准测试

```bash
go test -bench=. -benchmem ./time/
```

### 运行特定模块的基准测试

```bash
# 时间模块基准测试
go test -bench=. -benchmem ./time/

# 数据库模块基准测试
go test -bench=. -benchmem ./database/

# 队列模块基准测试
go test -bench=. -benchmem ./queue/
```

### 运行并发基准测试

```bash
go test -bench=BenchmarkConcurrent -benchmem ./time/
```

---

## 基准测试结果说明

### 测试输出格式

```
goos: windows
goarch: amd64
pkg: github.com/liaoran123/sfsEdgeStore/time
cpu: Intel(R) Core(TM) i7-XXXX
BenchmarkTimeToUnixTimestamp-8        1000000000   0.2845 ns/op   0 B/op   0 allocs/op
BenchmarkUnixTimestampToTime-8        1000000000   0.2813 ns/op   0 B/op   0 allocs/op
```

### 结果解读

- `BenchmarkXxx-8`：测试名称和 CPU 核心数
- `1000000000`：运行次数
- `0.2845 ns/op`：每次操作耗时（纳秒）
- `0 B/op`：每次操作分配的内存（字节）
- `0 allocs/op`：每次操作的内存分配次数

---

## 整体性能测试

### 内存占用测试

测试环境：
- 操作系统：Windows 10/11 或 Linux
- Go 版本：1.25+

测试步骤：
```bash
# 1. 编译项目
go build -o sfsedgestore

# 2. 运行程序
./sfsedgestore

# 3. 在另一个终端监控内存占用
# Windows: 任务管理器
# Linux: top / htop / ps
```

### 启动时间测试

```bash
# 使用 time 命令测试启动时间（Linux/Mac）
time ./sfsedgestore -test.run=^$

# Windows PowerShell
Measure-Command { .\sfsedgestore.exe -test.run=^$ }
```

### CPU 使用率测试

```bash
# 运行程序并监控 CPU 使用率
# Linux: top -p <pid>
# Windows: 任务管理器
```

### 数据库大小测试

1. 插入 18,681 条测试数据
2. 查看 `data/` 目录大小

---

## 测试环境参考

README.md 中的性能数据是在以下环境中测试的：

- **操作系统**：Windows 10/11
- **CPU**：Intel Core i7 或同等性能
- **内存**：16GB+
- **Go 版本**：1.25+
- **测试时间**：2026-03-08

---

## 贡献测试数据

如果你在不同的硬件环境下运行了测试，欢迎提交 PR 分享你的测试结果！

---

**最后更新**: 2026-03-09

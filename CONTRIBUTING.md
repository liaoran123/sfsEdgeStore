# 贡献指南

感谢您对 sfsEdgeStore 项目的关注！我们欢迎任何形式的贡献。

## 行为准则

请确保您的行为符合我们的 [贡献者公约](./CODE_OF_CONDUCT.md)。

## 开始之前

### 查找现有 Issue

在创建新 Issue 之前，请先搜索 [现有 Issues](https://github.com/your-username/sfsEdgeStore/issues)，看看是否已经有人报告了类似的问题或提出了类似的功能请求。

## 贡献方式

### 1. 报告 Bug

如果您发现了 Bug，请使用 [Bug 报告模板](.github/ISSUE_TEMPLATE/bug_report.md) 创建一个新的 Issue。

请提供尽可能详细的信息，包括：
- 复现步骤
- 预期行为
- 实际行为
- 环境信息（操作系统、Go 版本等）
- 日志输出

### 2. 提出功能请求

如果您有新功能的想法，请使用 [功能请求模板](.github/ISSUE_TEMPLATE/feature_request.md) 创建一个新的 Issue。

请描述：
- 您想要的功能
- 这个功能解决了什么问题
- 建议的实现方式
- 使用场景

### 3. 提问和寻求帮助

如果您有使用问题或需要帮助，请使用 [问题咨询模板](.github/ISSUE_TEMPLATE/question.md) 创建一个新的 Issue。

💡 **提示**：如果您需要专业的技术支持，可以考虑购买我们的[商业服务](./docs/pricing/SERVICES.md)！

### 4. 提交代码

我们欢迎代码贡献！请按照以下步骤操作：

#### 步骤 1：Fork 仓库

1. 在 GitHub 上 Fork 本仓库
2. Clone 您的 Fork 到本地：
   ```bash
   git clone https://github.com/your-username/sfsEdgeStore.git
   cd sfsEdgeStore
   ```

#### 步骤 2：创建分支

为您的更改创建一个新分支：
```bash
git checkout -b feature/your-feature-name
# 或
git checkout -b fix/your-bug-fix
```

#### 步骤 3：进行更改

- 遵循现有的代码风格和约定
- 使用 `go fmt` 格式化代码
- 为新功能添加测试
- 确保所有现有测试通过
- 根据需要更新文档

#### 步骤 4：提交更改

使用清晰简洁的提交信息：
```bash
git add .
git commit -m "Add feature: 描述您的更改"
```

#### 步骤 5：推送更改

```bash
git push origin feature/your-feature-name
```

#### 步骤 6：创建 Pull Request

1. 转到原始仓库
2. 点击 "Pull requests"，然后点击 "New pull request"
3. 选择您的分支并提交 PR
4. 填写 [PR 模板](.github/PULL_REQUEST_TEMPLATE.md)
5. 提供清晰的更改描述
6. 引用相关的 Issue（如果有）

## 代码风格

- 遵循 Go 的标准代码风格
- 使用 `go fmt` 格式化代码
- 保持函数小而专注
- 为复杂逻辑添加注释
- 使用描述性的变量和函数名

## 测试

### 运行测试

```bash
# 运行所有测试
go test -v ./...

# 运行带竞争检测
go test -v -race ./...

# 只运行特定包的测试
go test -v ./database/...
```

### 测试覆盖率

我们鼓励为新功能添加测试，保持高测试覆盖率。

## 文档

- 如果更改了功能，请更新 README.md
- 为代码添加注释
- 更新相关的文档文件

## 许可证

通过向本项目贡献代码，您同意您的贡献将根据 [Apache 2.0 许可证](./LICENSE) 进行许可。

## 社区

- 关注我们的 GitHub Discussions
- 参与技术讨论
- 分享您的使用案例

## 认可贡献者

我们会在项目中感谢所有贡献者！

---

再次感谢您的贡献！🎉

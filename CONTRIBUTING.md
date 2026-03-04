# Contributing to sfsDb EdgeX Adapter

We welcome contributions to the sfsDb EdgeX Adapter project! This guide will help you get started with contributing.

## Getting Started

### Prerequisites
- Go 1.25 or higher
- Git
- EdgeX Foundry development environment (optional for testing)

### Fork and Clone the Repository
1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/your-username/sfsdb-edgex-adapter.git
   cd sfsdb-edgex-adapter
   ```
3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/your-org/sfsdb-edgex-adapter.git
   ```

## Development Workflow

### Create a Branch
Create a new branch for your changes:
```bash
git checkout -b feature/your-feature-name
```

### Make Changes
- Follow the existing code style and conventions
- Add tests for any new functionality
- Ensure all existing tests pass
- Update documentation as needed

### Commit Changes
Commit your changes with a clear and concise commit message:
```bash
git commit -m "Add feature: description of your changes"
```

### Push Changes
Push your changes to your fork:
```bash
git push origin feature/your-feature-name
```

### Create a Pull Request
1. Go to the original repository on GitHub
2. Click "Pull requests" and then "New pull request"
3. Select your branch and submit the pull request
4. Provide a clear description of your changes
5. Reference any related issues

## Code Style

- Follow Go's standard code style
- Use `go fmt` to format your code
- Keep functions small and focused
- Add comments for complex logic
- Use descriptive variable and function names

## Testing

### Run Tests
```bash
go test ./...
```

### Test Coverage
Strive for high test coverage, especially for new functionality.

## Documentation

- Update README.md if you change functionality
- Add comments to your code
- Update any relevant documentation files

## Reporting Issues

If you find a bug or have a feature request:
1. Check if the issue already exists in the issue tracker
2. If not, create a new issue with:
   - A clear title
   - A detailed description
   - Steps to reproduce (for bugs)
   - Expected behavior
   - Actual behavior
   - Environment information

## Code of Conduct

Please be respectful and inclusive in all interactions. Follow the project's code of conduct.

## License

By contributing to this project, you agree that your contributions will be licensed under the Apache 2.0 License.

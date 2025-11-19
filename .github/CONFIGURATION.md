# GitHub Configuration

This directory contains GitHub-related configuration files for the DRS API repository.

## 📁 Directory Structure

```
.github/
├── workflows/                 # GitHub Actions workflows
│   ├── go-services.yml       # Go services CI
│   ├── dashboard.yml         # Dashboard CI
│   └── proto.yml             # Protocol Buffers CI
├── pull_request_template.md  # PR template
├── CI.md                     # CI/CD documentation
└── README.md                 # This file
```

## 🚀 GitHub Actions Workflows

### [go-services.yml](workflows/go-services.yml)
CI for Go services (API Gateway, Module Manager, DRS CLI)

- **Lint**: Static analysis with golangci-lint
- **Test**: Unit tests with coverage reporting
- **Build**: Build binaries for AMD64/ARM64

### [dashboard.yml](workflows/dashboard.yml)
CI for React/TypeScript Dashboard

- **Lint**: Code quality checks with ESLint
- **Type Check**: TypeScript type checking
- **Build**: Production build validation
- **Test**: Security audit

### [proto.yml](workflows/proto.yml)
CI for Protocol Buffers definitions

- **Validate**: Lint and breaking change detection with Buf
- **Generate**: Code generation validation

See [CI.md](CI.md) for more details.

## 📝 Pull Request Template

[pull_request_template.md](pull_request_template.md) defines the template used when creating pull requests.

**Includes:**
- Description of changes
- Related issues
- Type of change
- Testing details
- Screenshots
- Checklist

## 🔗 References

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [CI/CD Documentation](CI.md)
- [Repository README](../README.md)

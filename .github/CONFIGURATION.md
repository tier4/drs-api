# GitHub Configuration

This directory contains GitHub-related configuration files for the DRS API repository.

## 📁 Directory Structure

```
.github/
├── workflows/                 # GitHub Actions workflows
│   ├── go-services.yml       # Go services CI
│   ├── dashboard.yml         # Dashboard CI
│   ├── proto.yml             # Protocol Buffers CI
│   └── pr-title-check.yml    # PR title validation
├── pull_request_template.md  # PR template
├── dependabot.yml            # Dependabot configuration
├── release.yml               # Release notes configuration
├── CI.md                     # CI/CD documentation
└── CONFIGURATION.md          # This file
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

### [pr-title-check.yml](workflows/pr-title-check.yml)
PR title validation with Semantic PR format

- **Check**: Validates PR titles follow conventional commits format
- **Allowed types**: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `chore`
- **Scope**: Optional (not required)

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

## 🤖 Dependabot Configuration

[dependabot.yml](dependabot.yml) configures automated dependency updates.

**Monitored ecosystems:**
- **Go modules**: API Gateway, Module Manager, DRS CLI
- **npm**: Dashboard
- **GitHub Actions**: Workflow dependencies

**Settings:**
- Schedule: Weekly updates on Mondays
- Commit message prefix: `chore` (complies with PR title checker)
- Open PR limit: 5 per ecosystem

## 📋 Release Notes Configuration

[release.yml](release.yml) configures automatic release notes generation.

**Features:**
- Automatically categorizes PRs based on title prefix (feat, fix, docs, etc.)
- Excludes Dependabot PRs from main changelog (shown separately)
- Categories:
  - 🎉 New Features (`feat:`)
  - 🐛 Bug Fixes (`fix:`)
  - 📚 Documentation (`docs:`)
  - ⚡ Performance Improvements (`perf:`)
  - 🔧 Refactoring (`refactor:`)
  - 🧪 Tests (`test:`)
  - 🎨 Styling (`style:`)
  - 🔨 Maintenance (`chore:`)
  - 📦 Dependencies (auto-collapsed after 10 items)

**Usage:**
1. Navigate to Releases page on GitHub
2. Click "Draft a new release"
3. Choose a tag or create a new one
4. Click "Generate release notes"
5. Release notes will be automatically generated based on merged PRs

## 🔗 References

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [CI/CD Documentation](CI.md)
- [Repository README](../README.md)

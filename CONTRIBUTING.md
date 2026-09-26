# Contributing to vault-tool

## Getting Started

### Prerequisites
- Go 1.27 or later
- golangci-lint 2.14.0 or later
- Make
- Git

### Local Development

```bash
# Clone the repository
git clone https://github.com/rossigee/vault-tool.git
cd vault-tool

# Build the binary
make build

# Run tests
make test

# Run linters
make lint
```

## Development Workflow

1. **Create a branch** for your feature/fix:
   ```bash
   git checkout -b feat/your-feature-name
   ```

2. **Make changes** and test locally:
   ```bash
   make test
   make lint
   ```

3. **Commit with conventional commits**:
   ```bash
   git commit -m "feat: add new feature"
   git commit -m "fix: resolve issue with unsealing"
   git commit -m "perf: optimize key decryption"
   ```

4. **Push and create a pull request**:
   ```bash
   git push origin feat/your-feature-name
   ```

## Commit Message Conventions

Follow [Conventional Commits](https://www.conventionalcommits.org/) for commit messages:

- `feat:` New features
- `fix:` Bug fixes
- `perf:` Performance improvements
- `refactor:` Code refactoring
- `security:` Security fixes or improvements
- `build:` Build system or dependency changes
- `ci:` CI/CD workflow changes
- `docs:` Documentation updates
- `test:` Test additions or updates
- `chore:` Maintenance tasks

Example: `feat: add support for custom vault address`

**Note:** Only commits matching `^- (feat|fix|perf|refactor|security|build):` are included in auto-generated release changelogs.

## Running Tests

```bash
# Run all tests with coverage
make test

# Run specific test package
go test ./pkg/vault -v

# Run with race detector
go test ./... -race
```

## Code Quality

Before submitting a PR, ensure:

1. ✅ All tests pass: `make test`
2. ✅ Linting passes: `make lint`
3. ✅ Code is formatted: `make fmt`
4. ✅ `go vet` passes: `make vet`
5. ✅ Binary builds: `make build`

## Release Process

Releases are automated via GoReleaser:

1. **Tag the release**: `git tag v0.2.0` (must match semantic versioning)
2. **Push to trigger release**: `git push origin v0.2.0`

The GitHub Actions workflow will:
- Build binaries for Linux amd64 and arm64
- Package both .tar.gz archives and .deb packages for each architecture
- Generate a changelog from commits matching `^(feat|fix|perf|refactor|security|build):`
- Create a GitHub release with all artifacts attached
- Mark as prerelease if tag contains `-rc`, `-beta`, or `-alpha`

**Local testing:** Run `make release-snapshot` to dry-run the full release build locally without pushing to GitHub.

## Security

- Never commit secrets or credentials to the repository
- Report security vulnerabilities privately to the project maintainers
- All CI builds are scanned with Trivy for vulnerabilities

## Questions?

Open an issue on the [project repository](https://github.com/rossigee/vault-tool) or contact the maintainers.

# CLAUDE.md — Go HTTP Template

A Go net/http backend starter wired to the centralized CI/CD in
[`s3ntin3l8/.github`](https://github.com/s3ntin3l8/.github). If you are an AI agent
or developer working in a repo created from this template, read this first.

## First steps after creating a repo from this template

1. Rename the placeholders: `module` in `go.mod`, the `# Project Name` title in
   `README.md`, and the `module` path across `.go` files.
2. `make install-hooks` — installs pre-commit and pre-push hooks.
3. `make build` — verify everything compiles.
4. Decide your CI coverage floor: `ci-cd.yml` ships `coverage-fail-under: '0'`
   (a starter floor) — **ratchet it up** as you add real code.

## Commands (Makefile)

| Command | Does |
|---------|------|
| `make install-hooks` | Install pre-commit + pre-push hooks. |
| `make test` | Run Go tests with race detection and coverage. |
| `make lint` | Run pre-commit on all files (golangci-lint + go vet + detect-secrets). |
| `make fmt` | Format Go code with gofmt (and goimports if available). |
| `make vet` | Run go vet. |
| `make tidy` | Run go mod tidy. |
| `make vulncheck` | Run govulncheck for known vulnerabilities. |
| `make build` | Build all packages. |
| `make clean` | Remove build artifacts and test caches. |

## Layout

- `cmd/server/main.go` — entrypoint: flag parsing, config loading, graceful shutdown.
- `internal/config/` — YAML config loader with `${VAR}` environment expansion.
- `internal/httpapi/` — `net/http` server setup: route mux, security headers,
  panic recovery, request logging. `/health` endpoint.
- `config.example.yaml` — reference config with `${VAR}` placeholders.
- `Dockerfile` — multi-stage build → distroless non-root runtime with HEALTHCHECK.
- `.github/workflows/` — thin callers of the reusable workflows in `s3ntin3l8/.github`.
- `.editorconfig` — shared editor settings (LF, UTF-8, final newline; tabs for Go).
- `.claude/` — `settings.json` + `hooks/session-start.sh`: a SessionStart hook that
  installs Go deps and tooling (pre-commit, golangci-lint, govulncheck) so
  [Claude Code on the web](https://code.claude.com/docs/en/claude-code-on-the-web)
  sessions can build, test, and lint. Runs only in the remote env.

## CI/CD — uses centralized reusable workflows

Workflows here are **callers** of `s3ntin3l8/.github/.github/workflows/*.yml@main`:
`ci-cd.yml` (ci-go + docker-publish), `codeql.yml`, `dependency-review.yml`,
`release-please.yml`, `cleanup-ghcr.yml`.

**The #1 thing to get right:** a caller job that invokes a reusable workflow needing
write scopes **must declare a `permissions:` block** — the default `GITHUB_TOKEN`
is read-only and the run otherwise fails at startup with zero jobs. `build-docker`
needs `contents: read` + `packages: write`; `codeql` needs `security-events: write`;
`release-please` needs `contents: write` + `pull-requests: write`. See the
`s3ntin3l8/.github` README for the full table.

`ci-go` reads the Go version from `go.mod`, runs gofmt, go vet, go build, go test
-race with coverage, and govulncheck. The `pre-build-commands` input is available
for project-specific setup (most commonly stubbing `//go:embed` assets).

## Conventions

- **Go 1.26+, stdlib-first.** `net/http` router, no framework dependency.
- **Conventional Commits** — Release Please cuts versions/changelogs from them.
- **Linting enforced** by golangci-lint and go vet (config in `.pre-commit-config.yaml`);
  run `make lint` before pushing (the pre-push hook runs govulncheck).
- **Secrets:** never commit real credentials; `detect-secrets` runs in pre-commit
  and CI against `.secrets.baseline` (regenerate with
  `detect-secrets scan > .secrets.baseline` after vetting new detections).

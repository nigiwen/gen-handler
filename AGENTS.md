# Repository Guidelines

## Project Structure & Module Organization
This repository is a standalone Go CLI tool, but its primary consumer is `C:\www\golang\bsi\axis\devopsx`. `main.go` is the CLI entrypoint and routes subcommands such as `handler` and `data`. Put command orchestration in `cmd/`. Keep reusable logic in `internal/`, which is already split by responsibility: `generator/` for code generation, `parser/` and `scanner/` for source discovery, `updater/` for ProviderSet updates, `selector/` for interactive selection, `util/` for shared helpers, and `types/` for shared structs. Use `docs/` for release, publishing, and manual test notes. Keep cross-platform build scripts at the repository root (`build.sh`, `build.bat`).

## Build, Test, and Development Commands
Use `go build .` to compile the local CLI binary. Use `go run . handler -help` or `go run . data -help` to verify command wiring and flag behavior during development. Run `go test ./...` before opening a PR; add tests when changing non-trivial logic. Use `go install` to install the tool into your local `GOBIN`. Validate real generation flows from the consumer project root, typically `C:\www\golang\bsi\axis\devopsx`, for example by running the built tool there against `./internal/proto/axis/devopsx` and `./cmd/devopsx`. Release artifacts are built with `./build.sh v1.2.0` on Unix-like systems or `build.bat v1.2.0` on Windows; both write archives to `dist/`.

## Coding Style & Naming Conventions
Follow standard Go formatting with `gofmt -w .`; use `goimports` if imports need cleanup. Keep package names lowercase and descriptive. Exported identifiers use PascalCase; internal helpers use camelCase. Name files by responsibility in snake_case, for example `grpc_parser.go` or `selector.go`. Keep CLI messages concise and consistent with the existing Chinese-facing output unless a change intentionally standardizes messaging across commands.

## Testing Guidelines
Place unit tests beside the code they cover and name them `*_test.go`. Prefer table-driven tests for parser, scanner, updater, and utility logic. When generation behavior changes, combine `go test ./...` with a manual command check based on `docs/TEST.md`, using the `devopsx` project layout as the primary integration scenario. There is no formal coverage gate yet, but each behavior change should add or update meaningful test coverage.

## Commit & Pull Request Guidelines
Match the existing Conventional Commit style visible in history: `feat:`, `refactor:`, and `chore:` are the common prefixes. Keep subjects short, imperative, and scoped to one change. Pull requests should explain the affected command or package, list validation commands run, link related issues when applicable, and include sample CLI output when generator behavior, docs, or release packaging changes.

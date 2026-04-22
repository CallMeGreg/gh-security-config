# Copilot instructions for `gh-security-config`

This repository is a **precompiled Go-based GitHub CLI extension**. When contributing code,
follow the conventions below so changes stay consistent with the rest of the codebase.

## Project overview

- Binary name: `gh-security-config` (invoked as `gh security-config`).
- Module: `github.com/callmegreg/gh-security-config`.
- Built as a precompiled Go extension distributed via GitHub Releases using the
  [`cli/gh-extension-precompile`](https://github.com/cli/gh-extension-precompile) action.
- Key dependencies:
  - [`spf13/cobra`](https://github.com/spf13/cobra) — command structure.
  - [`pterm/pterm`](https://github.com/pterm/pterm) — terminal UI (spinners, tables, prompts, colors).
  - [`cli/go-gh/v2`](https://github.com/cli/go-gh) — authenticated GitHub API access and `gh` invocation.

## GitHub CLI extension best practices

- **Use `go-gh` for API calls.** Prefer `api.NewRESTClient` / `api.NewGraphQLClient` from
  `github.com/cli/go-gh/v2/pkg/api` over hand-rolled HTTP so the user's `gh auth` context,
  hosts, and headers are honored automatically. Isolate all GitHub API usage in
  `internal/api/` — commands and processors must not talk to the API directly.
- **Respect host overrides.** Accept a `--hostname` flag (see `cmd/shared_flags.go`) and
  pass it through to the API client so the extension works with GHES/GHEC data residency.
- **Never prompt in non-interactive mode.** When `--yes`/`--force` or piped input is
  detected, skip confirmations. Gate `pterm.DefaultInteractiveConfirm` behind an
  interactivity check.
- **Exit codes matter.** Return non-zero on any failure. Use cobra's `RunE` and return
  errors rather than calling `os.Exit` from deep in the call stack.
- **Do not commit the compiled binary.** `gh-security-config` is in `.gitignore`; keep it
  that way. Release artifacts are produced by CI via `release.sh` / the precompile action.
- **Keep the CLI scriptable.** Support `--csv` input/output, quiet modes, and stable
  machine-readable output where relevant. Avoid ANSI escapes when stdout is not a TTY
  (pterm handles this automatically if you use its printers rather than raw `fmt.Println`
  with colors).

## Cobra conventions

- One command per file under `cmd/` (`apply.go`, `delete.go`, `generate.go`, `modify.go`,
  `root.go`). New subcommands follow the same pattern.
- Define commands as package-level `*cobra.Command` variables and register them from
  `init()` with `rootCmd.AddCommand(...)`.
- Use `RunE func(cmd *cobra.Command, args []string) error` — never `Run`.
- Always set `Use`, `Short`, and `Long`. Include a realistic `Example` block.
- Validate arguments with `cobra.ExactArgs`, `cobra.MinimumNArgs`, `cobra.NoArgs`, etc.,
  or a custom `Args` function — not ad-hoc `len(args)` checks inside `RunE`.
- Shared flags live in `cmd/shared_flags.go`. Add new cross-command flags there and
  register them via a helper (e.g., `addCommonFlags(cmd)`) to keep definitions DRY.
- Mark required flags with `cmd.MarkFlagRequired(...)`. Use `MarkFlagsMutuallyExclusive`
  / `MarkFlagsRequiredTogether` instead of manual validation when possible.
- Keep `cmd/*.go` thin: parse flags, construct a request/options struct, hand off to a
  processor in `internal/processors/`. Business logic does not belong in `cmd/`.

## pterm conventions

- Prefer pterm printers over `fmt`:
  - `pterm.Info`, `pterm.Success`, `pterm.Warning`, `pterm.Error` for status messages.
  - `pterm.DefaultSpinner` for long-running single operations.
  - `pterm.DefaultProgressbar` for loops with known total counts (e.g., processing a
    list of orgs/repos).
  - `pterm.DefaultTable.WithHasHeader()` for tabular results.
  - `pterm.DefaultInteractiveConfirm` / `InteractiveSelect` / `InteractiveTextInput`
    for prompts — wrapped in `internal/ui/` helpers. Use those wrappers instead of
    calling pterm directly from commands.
- Always call `.Start()` / `.Stop()` (or use the returned printer) so spinners and
  progress bars clean up on Ctrl-C and on error paths (`defer spinner.Stop()`).
- Do not mix raw `log`/`fmt.Println` with pterm output in the same command — it breaks
  progress bar/spinner rendering.
- When running concurrently, synchronize writes to pterm printers or funnel status
  updates through a single goroutine; pterm is not safe for unsynchronized concurrent
  rendering.

## Go project layout

Follow the existing `cmd/` + `internal/` split. `internal/` is not importable externally,
which is what we want for a CLI extension.

```
main.go                  // minimal: cmd.Execute()
cmd/                     // cobra command wiring only
internal/
  api/                   // go-gh clients + GitHub API calls
  processors/            // orchestration: apply, delete, generate, modify, concurrent/sequential
  types/                 // domain types and error types
  ui/                    // pterm wrappers (confirmation, display, input)
  utils/                 // csv, flags, replication, validation helpers
```

Guidelines:
- Put reusable logic in `internal/...`, not in `cmd/`.
- Keep package names short, lowercase, no underscores, no stutter
  (`types.Organization`, not `types.OrganizationType`).
- Exported identifiers need doc comments starting with the identifier's name.
- Return errors — don't `log.Fatal` or `os.Exit` outside `main`/`cmd.Execute`.
- Wrap errors with `fmt.Errorf("doing X: %w", err)` so callers can `errors.Is` / `errors.As`.
  Custom sentinel errors belong in `internal/types/errors.go`.
- Prefer small interfaces defined at the *consumer* site. The processor interface in
  `internal/processors/interface.go` is the canonical example.
- Use `context.Context` for anything that does I/O; accept it as the first parameter.

## Testing conventions

- Tests live alongside code as `*_test.go` (see `internal/processors/concurrent_test.go`,
  `internal/utils/csv_test.go`, etc.). Do **not** create a separate `tests/` directory.
- Use table-driven tests with subtests:

  ```go
  func TestParseCSV(t *testing.T) {
      tests := []struct {
          name    string
          input   string
          want    []types.Organization
          wantErr bool
      }{
          // ...
      }
      for _, tt := range tests {
          t.Run(tt.name, func(t *testing.T) {
              got, err := ParseCSV(strings.NewReader(tt.input))
              if (err != nil) != tt.wantErr {
                  t.Fatalf("ParseCSV() error = %v, wantErr %v", err, tt.wantErr)
              }
              if !reflect.DeepEqual(got, tt.want) {
                  t.Errorf("ParseCSV() = %v, want %v", got, tt.want)
              }
          })
      }
  }
  ```

- Run tests with `go test ./...`. Add `-race` for anything in `internal/processors/concurrent*`.
- Use `t.Helper()` in test helper functions, `t.TempDir()` for filesystem fixtures,
  `t.Cleanup()` to tear down state. Never write to the real `$HOME` or `$CWD`.
- Mock the GitHub API by defining a small interface in the consuming package and passing
  a fake implementation in tests — do not hit the real API. Wrap the `go-gh` client in
  `internal/api/` behind an interface so tests can substitute a fake.
- Keep tests hermetic: no network, no reliance on `gh auth` state, no reliance on env
  vars beyond what the test sets via `t.Setenv`.
- Name tests `TestXxx`, benchmarks `BenchmarkXxx`, examples `ExampleXxx`. Keep one
  logical assertion per subtest case.
- For concurrent code, prefer deterministic tests (inject a worker count of 1, or use
  channels/`sync.WaitGroup` to synchronize) over `time.Sleep`.

## Build, lint, release

- Build locally: `go build -o gh-security-config`.
- Install as a local extension for manual testing: `gh extension install .` (from repo root).
- Format and vet: `gofmt -s -w .` and `go vet ./...` before committing.
- Cross-compilation and release archives are handled by the precompile
  GitHub Action — do not hand-edit release assets.

## What not to do

- Don't add new top-level directories without discussion; the layout above is deliberate.
- Don't introduce a new CLI framework, UI library, or HTTP client — stick with cobra,
  pterm, and go-gh.
- Don't call `os.Exit` from library code.
- Don't print to stdout from `internal/api/` or `internal/processors/`; return data and
  let `cmd/` + `internal/ui/` handle presentation.
- Don't commit the compiled binary, personal `orgs.csv` data, or tokens.
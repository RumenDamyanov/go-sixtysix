# Contributing

Thanks for your interest in improving go-sixtysix!

## Ways to Help

- Report bugs (include reproduction & environment)
- Propose enhancements (outline use case + minimal API surface)
- Improve docs (clarity, examples, spelling)
- Add tests (edge cases, rule enforcement scenarios)

## Development Setup

```bash
git clone https://github.com/rumendamyanov/go-sixtysix
cd go-sixtysix
go test ./...
```

Run example server:

```bash
go run ./examples/server
```

## Coding Guidelines

- Go 1.22+ features allowed.
- Keep public API small & focused.
- Avoid adding external deps without strong justification (stdlib bias).
- Write tests for new rule logic or engine behavior.

## Commit Style

Conventional-ish, but pragmatic (e.g. `engine: enforce follow suit after close`).

## Pull Requests

1. Open issue (optional for small fixes) describing change.
2. Fork & branch (`feature/short-description`).
3. Add tests (or rationale if not testable).
4. Ensure `go test ./...` passes.
5. Submit PR referencing issue.

## Code of Conduct

Participation governed by [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).

## License

By contributing you agree your contributions are MIT licensed (see LICENSE.md).

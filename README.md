# go-sixtysix

[![CI](https://github.com/rumendamyanov/go-sixtysix/actions/workflows/ci.yml/badge.svg)](https://github.com/rumendamyanov/go-sixtysix/actions/workflows/ci.yml)
![CodeQL](https://github.com/rumendamyanov/go-sixtysix/actions/workflows/github-code-scanning/codeql/badge.svg)
![Dependabot](https://github.com/rumendamyanov/go-sixtysix/actions/workflows/dependabot/dependabot-updates/badge.svg)
[![codecov](https://codecov.io/gh/rumendamyanov/go-sixtysix/branch/master/graph/badge.svg)](https://codecov.io/gh/rumendamyanov/go-sixtysix)
[![Go Report](https://goreportcard.com/badge/go.rumenx.com/sixtysix?2)](https://goreportcard.com/report/go.rumenx.com/sixtysix)
[![Go Reference](https://pkg.go.dev/badge/go.rumenx.com/sixtysix.svg)](https://pkg.go.dev/go.rumenx.com/sixtysix)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE.md)

Minimal backend engine + HTTP API for the traditional 24‑card trick‑taking game **Sixty‑six** (AKA *Schnapsen* variant family). Built for frontend clients (web, mobile, CLI bots) that want a stateless, deterministic core.

Rules reference: [Wikipedia – Sixty-six](https://en.wikipedia.org/wiki/Sixty-six_(card_game)). This implementation models a standard two‑player deal with marriages, closing the stock, trump exchange, and last trick bonus.

## Contents

- [Features](#features)
- [Install](#install)
- [Quick Start](#quick-start)
- [Container Image](#container-image)
- [Concepts](#concepts)
- [HTTP API](#http-api)
- [Game Rules Summary](#game-rules-summary)
- [Frontend Integration Ideas](#frontend-integration-ideas)
- [Project Layout](#project-layout)
- [Contributing](#contributing)
- [Security](#security)
- [License](#license)

## Features

- Deterministic game state creation (seeded RNG) for reproducible replays
- Lightweight in-memory session store (pluggable interface)
- Clear `Game` interface (validate + apply immutable-ish state transitions)
- HTTP API with small surface (sessions + actions)
- OpenAPI spec (see `openapi/`)
- Test coverage across engine, store, rules
- Simple deployment (pure stdlib)

## Install

Requires Go 1.22+.

You must run `go get` inside your own module (a directory containing a `go.mod`). If you don't have one yet, create it first:

```bash
go mod init myapp
```

Then fetch the library module (root package now provides the game plus subpackages `engine`, `store`):

```bash
go get go.rumenx.com/sixtysix@latest
```

Or pin a version (example):

```bash
go get go.rumenx.com/sixtysix@v1.0.1
```

Typical imports:

```go
import (
  "go.rumenx.com/sixtysix"
  "go.rumenx.com/sixtysix/engine"
  "go.rumenx.com/sixtysix/store"
)

func newSession() {
  mem := store.NewMemory()
  e := engine.New(mem)
  e.Register(sixtysix.Game{})
  // create and use sessions via engine API
}
```

If you host behind the vanity domain ensure the meta tags resolve (already configured for go.rumenx.com). If using a fork, update the module path accordingly.

## Quick Start

1. Install the module (if not already): `go get go.rumenx.com/sixtysix@latest`
1. Run the demo server:

```bash
go run ./examples/server
```

1. Create a session (seed optional):

```bash
curl -s -X POST 'http://localhost:8080/sessions?game=sixtysix&seed=42' | jq
```

1. List sessions:

```bash
curl -s 'http://localhost:8080/sessions?game=sixtysix' | jq
```

1. Play a card (`card` is an encoded int suit*100+rankValue):

```bash
curl -s -X POST http://localhost:8080/sessions/{id} -H 'Content-Type: application/json' -d '{"type":"play","payload":{"card":3011}}' | jq
```

1. Close stock:

```bash
curl -s -X POST http://localhost:8080/sessions/{id} -d '{"type":"closeStock"}'
```

More examples: see [docs/api.md](docs/api.md).

## Container Image

Build locally (multi-stage):

```bash
docker build -t go-sixtysix:dev .
```

Run:

```bash
docker run --rm -p 8080:8080 go-sixtysix:dev
```

Pass a different listen port (env or flag):

```bash
docker run --rm -e PORT=9090 -p 9090:9090 go-sixtysix:dev
```

Embed version data at build time:

```bash
docker build --build-arg VERSION=v0.1.0 --build-arg COMMIT=$(git rev-parse --short HEAD) --build-arg DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ) -t go-sixtysix:v0.1.0 .
```

Then run and observe the startup log for the injected values. The runtime image is distroless (`gcr.io/distroless/static:nonroot`).

## Concepts

Engine pieces:

| Piece | Purpose |
|-------|---------|
| `engine.Game` | Rule set: `InitialState`, `Validate`, `Apply` |
| `engine.Engine` | Registers games, manages sessions, dispatches actions |
| `store.Store` | Persistence abstraction (memory impl provided) |
| `api.Server` | Minimal HTTP adapter (serves JSON) |

Card encoding: `suit*100 + rankValue` where suits: Clubs=0, Diamonds=1, Hearts=2, Spades=3; rank values: A=11,10=10,K=4,Q=3,J=2,9=0.

## HTTP API

Core endpoints:

| Method | Path | Description |
|--------|------|-------------|
| GET | `/healthz` | Liveness probe |
| GET | `/games` | List registered games |
| POST | `/sessions?game=sixtysix&seed=SEED` | Create session |
| GET | `/sessions?game=sixtysix&offset=0&limit=20` | Page sessions |
| GET | `/sessions/{id}` | Fetch session (state snapshot) |
| POST | `/sessions/{id}` | Apply action `{type,payload}` |
| DELETE | `/sessions/{id}` | Delete session |

Schemas + examples: [openapi/sixtysix.yaml](openapi/sixtysix.yaml) and [docs/api.md](docs/api.md).

### Actions

| Type | Payload | Effect |
|------|---------|--------|
| `play` | `{card:int}` | Play a card; resolves trick after 2 plays |
| `closeStock` | - | Close stock: no further drawing; must follow suit |
| `declare` | `{suit:int}` | Marriage (K+Q) scoring (20 / 40 trump) at lead |
| `exchangeTrump` | - | Swap 9 of trump with upcard (while stock open, at lead) |

## Game Rules Summary

Short form (see [docs/rules.md](docs/rules.md) for detail):

1. 24‑card deck (A 10 K Q J 9 in four suits). Deal 6 each (3+3), stock remainder, last card face-up = trump.
2. Leader plays any card when stock open; follower may play any card until stock closed or empty; then must follow suit if possible.
3. Trick winner: higher of suit led; trumps beat non‑trumps.
4. Winner scores captured card values; first to 66 ends deal; +10 last trick bonus.
5. Marriage declaration at lead (holding K+Q) scores 20 (non‑trump) or 40 (trump).
6. Trump 9 exchange allowed at lead while stock open.

## Frontend Integration Ideas

- Maintain local optimistic state while posting actions (server returns authoritative state version).
- Use the seed to recreate initial hands client‑side for replay / spectator mode.
- Visual mapping for encoded cards: `suit = c/100`, `rankValue = c%100` → show face; build a lookup table.
- Implement WebSocket push wrapper watching session updates for real-time UI (out of scope here, easy extension).

See [docs/integration.md](docs/integration.md) for architecture suggestions.

## Project Layout

```text
sixtysix.go    # Game rules implementation (root package)
engine/        # Core engine + session orchestration
store/         # In-memory store (interface for alt backends)
api/           # HTTP server wiring
examples/      # Example executable (demo server)
openapi/       # OpenAPI specification
docs/          # Extended docs (rules, API, integration)
```

---

Extended documents:

- [Detailed Rules](docs/rules.md)
- [API Guide](docs/api.md)
- [Integration Notes](docs/integration.md)

---

Future ideas: persistence backends, matchmaking service, WebSocket streaming, multi-deal match structure.

## Infrastructure Philosophy

This repository intentionally ships only:

- A minimal HTTP example (`examples/server`)
- A single multi-stage `Dockerfile`
- An optional ergonomic `Makefile` (build/test/run/docker shortcuts)

Deliberately omitted for now:

- `docker-compose.yml` (no external services are required; would add noise)
- Orchestration manifests (helm/kustomize) – deployment concerns belong in the
  consuming application / ops repo
- Additional runtime daemons (metrics, tracing) – premature until real usage

Rationale: keep the core game engine small, dependency‑free, and easy to vendor / embed.
Projects that adopt this library can define their own stack (DB, cache, auth, frontend) and introduce compose or infra templates there without having to unwind demo defaults baked here.

When to reconsider adding compose here:

1. We add a persistence backend example (e.g. Postgres) needing a service container.
2. We publish a combined demo (frontend + backend) requiring one-step spin up.
3. Integration tests demand ephemeral infrastructure.

Until then, simplicity wins.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Please follow the [Code of Conduct](CODE_OF_CONDUCT.md).

## Security

Report vulnerabilities privately – process described in [SECURITY.md](SECURITY.md).

## Funding / Support

If you find this useful, see [FUNDING.md](FUNDING.md).

## License

MIT – see [LICENSE.md](LICENSE.md).

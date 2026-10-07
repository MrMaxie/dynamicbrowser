# dynamicbrowser

Go Hello World starter, version `0.1.0`.

## Quick start

Install [mise](https://mise.jdx.dev/), then run:

```sh
mise trust
mise install
mise exec -- just run
mise exec -- just version
```

The commands print `Hello, World!` and `0.1.0`, respectively. Tool versions are pinned in `mise.toml`; Node.js is used only by the OpenSpec CLI.

## Development

```sh
mise exec -- just fmt         # format Go code with gofmt
mise exec -- just check       # lint, formatting check, tests, specs, build
mise exec -- just build       # executable in bin/
```

`just lint` uses golangci-lint with the standard linters and gofmt checks. In GoLand, open the repository as a Go project and use the Go SDK version from `mise.toml`.

## Project sources

- `VERSION`: application version, embedded at build time. Use Semantic Versioning; update it for each release.
- `todo.txt`: shared todo.txt queue, initially empty.
- `openspec/specs/`: accepted requirements. Propose future changes under `openspec/changes/` using the configured Arcantry schema.
- `arcantry.toml`: shared source configuration; todo.txt and OpenSpec are managed independently.

Licensed under the [Apache License 2.0](LICENSE).

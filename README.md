# Gothify

Gothify is an interactive CLI generator for modern full-stack Go applications.
It creates pragmatic MPA or SPA boilerplates without requiring a heavy
JavaScript framework for every project.

## Features

- Interactive project creation with `gothify create <project-name>`
- Non-interactive project creation through CLI flags
- Go HTTP router choices:
  - Chi
  - Fiber
  - Echo
  - Standard library `net/http`
- Frontend integration choices:
  - Go Templ + HTMX
  - Go `html/template` + Alpine.js
  - Go + Inertia.js + React
- CSS choices:
  - Tailwind CSS
  - Bulma/PicoCSS
  - Plain CSS
- Authentication starters:
  - Signed session cookies
  - JWT access and refresh tokens
  - OAuth2 with Goth for Google and GitHub
- Optional PWA manifest and service worker
- Optional static asset embedding with `go:embed`
- Automatic `go mod tidy` after generation
- Generated Vite development and production scripts for Inertia projects

## Requirements

- Go 1.23 or newer
- Node.js and npm are only required for generated Inertia/Tailwind frontend workflows

## Zero-Install Usage

After Gothify is published to a public Go repository, run it from anywhere
without installing a Gothify binary:

```powershell
go run github.com/fahmi-azzuhri/gothify@latest create my-app
go run github.com/fahmi-azzuhri/gothify@latest help
```

Replace `github.com/fahim-azzuhri/gothify` with the actual repository module path.
While developing Gothify locally, use:

```powershell
go run . help
```

## Interactive Creation

Run:

```powershell
go run . create my-app
```

The CLI asks for the router, frontend integration, CSS strategy,
authentication strategy, PWA support, embedded assets, and Go module path.

## Non-Interactive Creation

All choices can be passed as flags, which makes Gothify suitable for scripts
and CI workflows:

```powershell
go run . create my-app `
  --router chi `
  --template templ `
  --css tailwind `
  --auth session `
  --pwa `
  --embed `
  --module example.com/my-app
```

Supported values:

| Flag         | Values                              |
| ------------ | ----------------------------------- |
| `--router`   | `chi`, `fiber`, `echo`, `net-http`  |
| `--template` | `templ`, `alpine`, `inertia`        |
| `--css`      | `tailwind`, `bulma`, `plain`        |
| `--auth`     | `session`, `jwt`, `oauth2`          |
| `--module`   | Any valid Go module path            |
| `--pwa`      | Enable PWA files                    |
| `--embed`    | Embed static assets with `go:embed` |

## Generated Project

Generated projects include:

```text
my-app/
├── cmd/server/main.go
├── internal/auth/auth.go
├── internal/config/config.go
├── internal/httpserver/router.go
├── web/static/
├── web/templates/
├── go.mod
└── go.sum
```

The selected router is wired into the generated server. The selected
authentication strategy and frontend integration also determine generated
dependencies and starter files.

## Run a Generated Project

```powershell
cd my-app
go run ./cmd/server
```

Open http://localhost:8080.

Generated servers expose a `/health` endpoint and serve static assets from
`/static/`.

### Inertia Projects

For an Inertia project, install frontend dependencies and start Vite:

```powershell
npm install
npm run dev
```

Production frontend commands are also generated:

```powershell
npm run build
npm run preview
```

## Development

Run the Gothify test suite from the repository root:

```powershell
go test ./...
```

The project is split into focused modules:

- `main.go`: CLI entrypoint, prompts, and flag parsing
- `generator.go`: project file generation and dependency setup
- `server_templates.go`: generated router/server templates
- `auth_templates.go`: generated authentication and frontend templates
- `main_test.go`: generator tests

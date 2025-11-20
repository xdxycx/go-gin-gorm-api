# PR Notes for recent changes

Summary:
- Applied fixes to `Dockerfile` (multi-stage build, install `wget` for healthcheck, run as non-root)
- Cleaned and tidied `go.mod` and regenerated `go.sum`
- Fixed a Go syntax bug in `app/handlers/dynamic_api.go`
- Adjusted dynamic route to `/api/v1/dynamic/run/*path` to avoid Gin route conflicts
- Added query execution protections: timeout (5s), max rows (1000) and audit logs

Testing done locally:
- `go mod tidy` to populate `go.sum`
- `docker-compose build --no-cache` and `docker-compose up -d`
- Verified application starts, migrations run, and root endpoint returns HTTP 200

Recommendations / follow-ups:
- Add authentication for dynamic register/execute endpoints (API key or JWT)
- Consider adding an `audit` DB table to persist execution logs (currently logged to stdout)
- Add CI workflow to run `go test`, `go vet`, `golangci-lint`, and a build step
- Consider introducing DB migrations tool (`golang-migrate`) for production

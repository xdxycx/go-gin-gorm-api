# Changelog

All notable changes to this project will be documented in this file.

## Unreleased

- Fix: Dockerfile runtime dependencies and healthcheck (add wget)
- Fix: Clean up malformed `go.mod` and run `go mod tidy`
- Fix: Syntax bug in `dynamic_api` result handling
- Change: Dynamic execution route moved to `/api/v1/dynamic/run/*path` to avoid route conflicts
- Feature: Add execution timeout (5s), maximum rows limit (1000) and audit logging for dynamic SQL execution

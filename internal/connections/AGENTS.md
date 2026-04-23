# Connections Package Guide

## Scope

- Use this file for connection lifecycle, protocol-specific connection behavior, input buffering, and client settings in `internal/connections`.
- This package is protocol-sensitive and concurrency-sensitive.

## Working Rules

- Preserve the distinction between telnet, WebSocket, and SSH connection paths. Do not unify them casually if the protocol semantics differ.
- Be careful with connection state transitions, heartbeat behavior, and write locking. Small changes can cause dropped sessions or subtle races.
- Keep input-history and handler-chain behavior compatible unless the task explicitly changes user-facing input semantics.
- Treat screen dimensions, client capabilities, and protocol options as live client state rather than static config.

## Verification

- Run targeted package tests for connection-state or client-input changes.
- If the change affects WebSocket health, SSH behavior, or telnet option handling, verify the exact protocol path you changed.
- Call out any concurrency or live-network behavior that was not exercised directly.

## Documentation

- Keep only durable protocol and state-management rules here.

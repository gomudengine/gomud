# Term Package Guide

## Scope

- Use this file for terminal protocol helpers, telnet/IAC handling, escape-sequence behavior, and terminal formatting support in `internal/term`.
- This package is protocol-sensitive and client-visible.

## Working Rules

- Preserve wire-format compatibility for existing telnet and terminal behavior unless the task explicitly changes protocol support.
- Be careful with control sequences, negotiation constants, and encoding/escaping behavior. Small mistakes can break multiple client types.
- Keep protocol constants and low-level helpers here; do not add higher-level gameplay policy.
- If a change affects how clients negotiate or render data, inspect the matching `connections` or `inputhandlers` caller too.

## Verification

- Run targeted package tests for terminal helper changes.
- If protocol behavior changed, verify the exact client path affected or note that live client validation was not run.
- Use a higher-level check for changes that affect login or live session rendering.

## Documentation

- Keep only durable protocol guardrails here.

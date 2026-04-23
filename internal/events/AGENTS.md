# Events Package Guide

## Scope

- Use this file for event definitions, listener registration behavior, queue behavior, priorities, and event-flow control in `internal/events`.
- This package is a core integration point across the engine and modules.

## Working Rules

- Preserve compatibility of existing event types unless the task explicitly changes an event contract.
- Be careful with listener ordering, priority behavior, and unique-event semantics. Small changes here can shift game behavior far from this package.
- Prefer adding narrowly scoped events over overloading existing events with unrelated responsibilities.
- If a package relies on event ordering, document that assumption in the consuming package too instead of hiding it only here.
- Avoid introducing event-side policy that belongs in handlers or modules.

## Verification

- Run targeted `internal/events` tests for queue, ordering, or uniqueness changes.
- If the change affects a widely used event, verify at least one real consumer path in addition to package tests.
- Call out any untested downstream listeners when changing event contracts.

## Documentation

- Keep this file about event contracts and ordering risk, not an exhaustive event catalog.

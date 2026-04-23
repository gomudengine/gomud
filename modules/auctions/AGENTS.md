# Auctions Module Guide

## Scope

- Use this file for active-auction state, auction command behavior, periodic auction ticks, and auction persistence in `modules/auctions`.
- This module also has external integrations through broadcast behavior and Discord-facing auction updates.

## Working Rules

- Preserve auction lifecycle semantics unless the task explicitly changes player-facing auction rules.
- Be careful with round-based timing, persistence, and restart survival. Those behaviors are part of the module contract.
- If a change affects `AuctionUpdate` events, treat it as an integration-sensitive change because other systems consume those updates.
- Keep bid validation, auction state mutation, and rendered auction messages aligned.

## Verification

- Run targeted module tests for auction state, bidding, or persistence changes.
- If event payloads or reminders change, verify the exact event path or rendered output involved.
- Call out any untested restart-survival or cross-system notification behavior.

## Documentation

- Keep only durable auction-state and integration rules here.

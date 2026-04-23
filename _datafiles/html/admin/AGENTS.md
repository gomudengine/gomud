# Admin HTML Guide

## Scope

- Use this file for admin templates and shared admin-side static assets under `_datafiles/html/admin/`.
- These pages are server-rendered by `internal/web` and use the shared admin template shell.

## Working Rules

- Start pages with `{{template "header" .}}` and end them with `{{template "footer" .}}`.
- Keep page-specific CSS inline in a top-level `<style>` block and page-specific JavaScript inline in a `<script>` block near the end of the page.
- Use vanilla JS only. Do not introduce Bootstrap, HTMX, jQuery, Tailwind, or other client frameworks here.
- Reuse existing admin page layouts and utilities before inventing a new pattern.
- If a page calls admin APIs, give it a matching `*-api.html` reference page and link to it from the page subtitle.
- Use the shared `AdminAPI` helper for admin requests instead of raw `fetch`.
- Keep functions triggered from inline HTML handlers globally reachable with the same pattern used by existing pages.
- Do not mutate the visible data model optimistically before the admin API confirms success.

## Verification

- If you change shared admin JS assets, run `make js-lint` when the changed file is on the lint path.
- For page-template changes, load the affected admin page and verify the exact behavior you changed.
- For API-driven pages, verify create/edit/delete or refresh behavior against the matching admin endpoint when practical.
- If you add a new editor page, also verify its matching API reference page renders and links correctly.

## Documentation

- Keep long UI walkthroughs and component catalogs out of this file.
- Add only durable local rules that help agents avoid repeating admin-page mistakes.

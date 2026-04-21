# WebHelp Module Context

## Overview

The `modules/webhelp` module adds a web-based help browser to the GoMud admin/public web interface. It exposes two HTTP endpoints that render all in-game help topics as HTML pages, allowing players to browse help content from a browser without being connected to the game.

## Key Components

### Module (`webhelp.go`)

- Registered as plugin `webhelp` version `1.0`.
- Embeds its HTML templates from `files/` using `//go:embed files/*`.
- Registers two web pages via `plug.Web.WebPage`:
  - **`/help`** (`help.html`, public) — lists all non-admin help topics grouped by category.
  - **`/help-details`** (`help-details.html`, not listed in nav) — renders a single help topic by name, provided via `?search=<term>` query parameter.

### Data Handlers

- **`getHelpCategories`**: Iterates all help topics from `keywords.GetAllHelpTopicInfo()`, filters out admin-only topics, groups by category (skills get their own `"skills"` category regardless of their declared category), sorts categories and topic lists, and returns the data for the template.
- **`getHelpCommand`**: Looks up a single help topic by search term using `usercommands.GetHelpContents`. Converts the ANSI-tagged help text to HTML via `ansitags.Parse(..., ansitags.HTML)` for browser rendering.

## Dependencies

- `internal/keywords`: `GetAllHelpTopicInfo()` for the full topic list
- `internal/plugins`: Plugin registration and web page registration
- `internal/usercommands`: `GetHelpContents(term)` for individual help topic content
- `github.com/GoMudEngine/ansitags`: ANSI-to-HTML conversion

## File Structure

```
modules/webhelp/
  webhelp.go
  files/
    datafiles/html/public/
      help.html          # Category/topic listing page
      help-details.html  # Individual topic detail page
```

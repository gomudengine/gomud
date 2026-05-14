# Panel Layout Functions

Panel layouts let scripts build richly formatted terminal displays made up of
titled boxes arranged side by side or stacked vertically. The same system used
to render the in-game character status screen is fully available to scripts.

- [Panel Layout Functions](#panel-layout-functions)
  - [Concepts](#concepts)
    - [Layout model](#layout-model)
    - [Border styles](#border-styles)
    - [Character sets](#character-sets)
    - [Titles](#titles)
    - [Labels and values](#labels-and-values)
  - [PanelLayoutLoad(name) PanelLayoutObject](#panellayoutloadname-panellayoutobject)
  - [PanelLayoutNew(\[opts\]) PanelLayoutObject](#panellayoutnewopts-panellayoutobject)
  - [PanelLayoutObject.AddSlot() SlotObject](#panellayoutobjectaddslot-slotobject)
  - [PanelLayoutObject.Panel(id) PanelObject](#panellayoutobjectpanelid-panelobject)
  - [PanelLayoutObject.Render() string](#panellayoutobjectrender-string)
  - [SlotObject.AddRow(\[ids\]) SlotObject](#slotobjectaddrowids-slotobject)
  - [PanelObject.SetCharset(name) PanelObject](#panelobjectsetcharsetnname-panelobject)
  - [PanelObject.SetTitle(title) PanelObject](#panelobjectsettitletitle-panelobject)
  - [PanelObject.SetMinWidth(n) PanelObject](#panelobjectsetminwidthn-panelobject)
  - [PanelObject.SetColumns(n) PanelObject](#panelobjectsetcolumnsn-panelobject)
  - [PanelObject.SetColumnGap(n) PanelObject](#panelobjectsetcolumngapn-panelobject)
  - [PanelObject.Add(fullLabel, shortLabel, value) PanelObject](#panelobjectaddfullLabel-shortlabel-value-panelobject)
  - [PanelObject.AddBlank() PanelObject](#panelobjectaddblank-panelobject)
  - [Examples](#examples)
    - [Simple single panel](#simple-single-panel)
    - [Two panels side by side](#two-panels-side-by-side)
    - [Status-style nested layout](#status-style-nested-layout)
    - [Loading a layout from a YAML file](#loading-a-layout-from-a-yaml-file)
    - [Two-column stats panel](#two-column-stats-panel)
    - [Progress bar in a panel](#progress-bar-in-a-panel)

---

## Concepts

### Layout model

A layout is made up of **slots**, **rows**, and **panels**.

```
Layout
 ├── Slot (left column)
 │    └── Row
 │         └── Panel "info"
 └── Slot (right column)
      ├── Row
      │    └── Panel "attributes"
      └── Row
           ├── Panel "wealth"
           └── Panel "training"
```

- **Slots** are the top-level vertical columns of the layout. They sit side by
  side when rendered.
- **Rows** live inside a slot and stack vertically. Each row holds one or more
  panels placed side by side.
- **Panels** are the individual titled boxes that contain your data.

### Border styles

`border` controls which content rows show side-border characters (`│`).

| Value | Behaviour |
| --- | --- |
| `"full"` | Every content row has side borders: `│ label  value │` |
| `"open"` | Only the first and last content rows have side borders; interior rows are open |

### Character sets

`charset` selects the box-drawing characters used for all panel borders.

| Value | Top-left | Horizontal-top | Top-right | Vertical-left | Vertical-right | Bottom-left | Horizontal-bottom | Bottom-right |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `"single"` (default) | `┌` | `─` | `┐` | `│` | `│` | `└` | `─` | `┘` |
| `"double"` | `╔` | `═` | `╗` | `║` | `║` | `╚` | `═` | `╝` |
| `"rounded"` | `╭` | `─` | `╮` | `│` | `│` | `╰` | `─` | `╯` |

An **8-character literal** can be used instead of a named preset. The characters
are read in order: TopLeft, HorizontalTop, TopRight, VerticalLeft, VerticalRight,
BottomLeft, HorizontalBottom, BottomRight. This allows asymmetric left/right and
top/bottom borders:

```javascript
// Double top, single bottom, double left, single right
var layout = PanelLayoutNew({ charset: "╔═╗║│╚─┘" });
```

### Titles

The `title` string is placed verbatim into the top border of a panel. It may
contain ANSI colour tags. The visual width is measured with tags stripped, so
colour codes do not affect alignment.

The spaces around the title text become the padding between the corner
characters and the title. To match the standard GoMud `.:` prefix style:

```javascript
panel.SetTitle(' <ansi fg="black-bold">.:</ansi><ansi fg="20">My Panel</ansi> ');
```

To produce a plain unstyled title:

```javascript
panel.SetTitle(' My Panel ');
```

To produce a panel with no title at all, pass an empty string or omit
`SetTitle` entirely.

### Labels and values

Each content row has a **full label**, a **short label**, and a **value**.

- The renderer always tries the full label first.
- If the full label plus the value is too wide for the panel, the short label
  is used instead.
- Values are plain strings and may contain ANSI colour tags. Width is always
  measured with tags stripped.

---

## [PanelLayoutLoad(name) PanelLayoutObject](/internal/scripting/panel_func.go)

Loads a panel layout definition from the datafiles panel-layouts directory.
`name` is the path relative to `panel-layouts/`, without a file extension.

Throws a JavaScript error if the file does not exist or cannot be parsed.

| Argument | Explanation |
| --- | --- |
| name | Path to the layout file, e.g. `"character/status"` |

```javascript
var layout = PanelLayoutLoad("character/status");
layout.Panel("info").Add("Level:", "Lvl:", "42");
user.SendText(layout.Render());
```

---

## [PanelLayoutNew(\[opts\]) PanelLayoutObject](/internal/scripting/panel_func.go)

Creates a blank panel layout entirely in script, with no YAML file required.

`opts` is an optional object. All properties are optional.

| Property | Type | Default | Explanation |
| --- | --- | --- | --- |
| `border` | string | `"full"` | `"full"` or `"open"` — see [Border styles](#border-styles) |
| `charset` | string | `"single"` | `"single"`, `"double"`, `"rounded"`, or an 8-character literal — see [Character sets](#character-sets) |
| `gap` | int | `1` | Spaces between panels placed side by side, and between slots |
| `margin` | int | `0` | Spaces prepended to every output line |

```javascript
var layout = PanelLayoutNew({ border: "full", charset: "rounded", gap: 1, margin: 1 });
```

---

## [PanelLayoutObject.AddSlot() SlotObject](/internal/scripting/panel_func.go)

Adds a new vertical slot (column) to the layout and returns a `SlotObject` for
adding rows of panels into it. Slots are rendered side by side in the order
they are added.

```javascript
var leftSlot  = layout.AddSlot();
var rightSlot = layout.AddSlot();
```

---

## [PanelLayoutObject.Panel(id) PanelObject](/internal/scripting/panel_func.go)

Returns the `PanelObject` with the given `id` for configuration and data
population. Throws a JavaScript error if `id` does not exist in the layout.

| Argument | Explanation |
| --- | --- |
| id | The panel identifier as given to `SlotObject.AddRow()` or defined in a YAML layout file |

```javascript
layout.Panel("stats").Add("Strength:", "Str:", "42");
```

---

## [PanelLayoutObject.Render() string](/internal/scripting/panel_func.go)

Synthesizes the entire layout into a single terminal string ready to send to a
player. ANSI tags in titles, labels, and values are preserved as-is.

```javascript
user.SendText(layout.Render());
```

---

## [SlotObject.AddRow(\[ids\]) SlotObject](/internal/scripting/panel_func.go)

Adds a horizontal row of panels to the slot. `ids` is an array of string panel
identifiers, one per panel in the row. Panels are created with default settings
and can be configured afterwards via `layout.Panel(id)`.

Returns the slot for chaining.

| Argument | Explanation |
| --- | --- |
| ids | Array of string IDs for the panels to create in this row |

```javascript
var slot = layout.AddSlot();
slot.AddRow(["health", "mana"]);   // two panels side by side
slot.AddRow(["gold"]);             // one panel underneath
```

---

## [PanelObject.SetCharset(name) PanelObject](/internal/scripting/panel_func.go)

Overrides the border character set for this panel, ignoring the layout-level
`charset`. If not called, the panel inherits the layout's charset.

Returns the panel for chaining.

| Argument | Explanation |
| --- | --- |
| name | `"single"`, `"double"`, `"rounded"`, or an 8-character literal (TopLeft, HorizontalTop, TopRight, VerticalLeft, VerticalRight, BottomLeft, HorizontalBottom, BottomRight). An unrecognised value falls back to `"single"`. |

```javascript
// Give one panel a double border while the rest use the layout default
layout.Panel("highlight").SetCharset("double");

// Asymmetric: double top/left, single bottom/right
layout.Panel("custom").SetCharset("╔═╗║│╚─┘");
```

---

## [PanelObject.SetTitle(title) PanelObject](/internal/scripting/panel_func.go)

Sets the title string rendered verbatim into the top border. May contain ANSI
tags. Returns the panel for chaining.

| Argument | Explanation |
| --- | --- |
| title | Title string, including any surrounding spaces and ANSI tags |

```javascript
layout.Panel("info").SetTitle(' <ansi fg="20">Character Info</ansi> ');
```

---

## [PanelObject.SetMinWidth(n) PanelObject](/internal/scripting/panel_func.go)

Sets the minimum inner content width of the panel in characters. The panel
expands automatically if any row's content is wider. Returns the panel for
chaining.

| Argument | Explanation |
| --- | --- |
| n | Minimum inner width in terminal characters |

```javascript
layout.Panel("info").SetMinWidth(30);
```

---

## [PanelObject.SetColumns(n) PanelObject](/internal/scripting/panel_func.go)

Sets how many label+value pairs are placed on each rendered line. Supported
values are `1` (default) and `2`. When set to `2`, rows are paired: the first
and second rows share a line, the third and fourth share the next line, and so
on. An odd trailing row spans the full panel width.

Returns the panel for chaining.

| Argument | Explanation |
| --- | --- |
| n | `1` for single-column, `2` for two-column layout |

```javascript
layout.Panel("stats").SetColumns(2);
```

---

## [PanelObject.SetColumnGap(n) PanelObject](/internal/scripting/panel_func.go)

Sets the number of spaces between the two columns when `columns` is `2`.
Defaults to `2`. Returns the panel for chaining.

| Argument | Explanation |
| --- | --- |
| n | Number of spaces between columns |

```javascript
layout.Panel("stats").SetColumns(2).SetColumnGap(3);
```

---

## [PanelObject.Add(fullLabel, shortLabel, value) PanelObject](/internal/scripting/panel_func.go)

Appends a label+value row to the panel. The renderer uses `fullLabel` when it
fits; if the panel is too narrow it falls back to `shortLabel`. `value` can be
any string including ANSI tags. Returns the panel for chaining.

| Argument | Explanation |
| --- | --- |
| fullLabel | Preferred label text. May contain ANSI tags. |
| shortLabel | Fallback label used when the full label does not fit. |
| value | The value string. May contain ANSI tags. |

```javascript
layout.Panel("info")
    .Add('<ansi fg="yellow">Health:</ansi>', '<ansi fg="yellow">HP:</ansi>', "142/200")
    .Add('<ansi fg="yellow">Mana:  </ansi>', '<ansi fg="yellow">MP:</ansi>', "80/100");
```

---

## [PanelObject.AddBlank() PanelObject](/internal/scripting/panel_func.go)

Appends an empty spacer row to the panel. Useful for visual separation between
groups of rows. Returns the panel for chaining.

```javascript
layout.Panel("info")
    .Add("Name:", "N:", actorName)
    .AddBlank()
    .Add("Level:", "Lvl:", level);
```

---

## Examples

### Simple single panel

A single panel with a title and a few rows.

```javascript
function onCommand_stats(rest, user, room) {

    var layout = PanelLayoutNew({ border: "full", charset: "single", margin: 1 });

    var slot = layout.AddSlot();
    slot.AddRow(["info"]);

    layout.Panel("info")
        .SetTitle(' <ansi fg="black-bold">.:</ansi><ansi fg="20">Character</ansi> ')
        .SetMinWidth(28)
        .Add('<ansi fg="yellow">Name:  </ansi>', '<ansi fg="yellow">N:</ansi>', user.GetName())
        .Add('<ansi fg="yellow">Level: </ansi>', '<ansi fg="yellow">Lvl:</ansi>', String(user.GetLevel()))
        .Add('<ansi fg="yellow">Health:</ansi>', '<ansi fg="yellow">HP:</ansi>', user.GetHealth() + "/" + user.GetHealthMax());

    user.SendText(layout.Render());
    return true;
}
```

Output:
```
 ┌─ .:Character ───────────────┐
 │ Name:   Aldric               │
 │ Level:  12                   │
 │ Health: 142/200              │
 └──────────────────────────────┘
```

---

### Two panels side by side

Two panels in the same slot row, rendered side by side.

```javascript
var layout = PanelLayoutNew({ border: "full", gap: 1, margin: 1 });

var slot = layout.AddSlot();
slot.AddRow(["combat", "defense"]);

layout.Panel("combat")
    .SetTitle(' <ansi fg="20">Combat</ansi> ')
    .SetMinWidth(20)
    .Add('<ansi fg="yellow">Attacks:</ansi>', '<ansi fg="yellow">Atk:</ansi>', "2")
    .Add('<ansi fg="yellow">Damage: </ansi>', '<ansi fg="yellow">Dmg:</ansi>', "8-14");

layout.Panel("defense")
    .SetTitle(' <ansi fg="20">Defense</ansi> ')
    .SetMinWidth(20)
    .Add('<ansi fg="yellow">Armor:  </ansi>', '<ansi fg="yellow">Arm:</ansi>', "18")
    .Add('<ansi fg="yellow">Dodge:  </ansi>', '<ansi fg="yellow">Ddg:</ansi>', "12%");

user.SendText(layout.Render());
```

Output:
```
 ┌─ .:Combat ──────────────┐ ┌─ .:Defense ─────────────┐
 │ Attacks: 2               │ │ Armor:   18              │
 │ Damage:  8-14            │ │ Dodge:   12%             │
 └──────────────────────────┘ └──────────────────────────┘
```

---

### Status-style nested layout

The classic status page shape: a tall left panel next to a right column that
has two stacked rows.

```javascript
var layout = PanelLayoutNew({ border: "open", charset: "single", gap: 1, margin: 1 });

// Left slot: one tall panel
var left = layout.AddSlot();
left.AddRow(["info"]);

// Right slot: two stacked rows
var right = layout.AddSlot();
right.AddRow(["stats"]);
right.AddRow(["gold", "training"]);

layout.Panel("info")
    .SetTitle(' <ansi fg="black-bold">.:</ansi><ansi fg="20">Info</ansi> ')
    .SetMinWidth(26)
    .Add('<ansi fg="yellow">Name:  </ansi>', '<ansi fg="yellow">N:</ansi>',   user.GetName())
    .Add('<ansi fg="yellow">Race:  </ansi>', '<ansi fg="yellow">Rce:</ansi>', user.GetRace())
    .Add('<ansi fg="yellow">Level: </ansi>', '<ansi fg="yellow">Lvl:</ansi>', String(user.GetLevel()))
    .Add('<ansi fg="yellow">Health:</ansi>', '<ansi fg="yellow">HP:</ansi>',  user.GetHealth() + "/" + user.GetHealthMax())
    .Add('<ansi fg="yellow">Mana:  </ansi>', '<ansi fg="yellow">MP:</ansi>',  user.GetMana()   + "/" + user.GetManaMax());

layout.Panel("stats")
    .SetTitle(' <ansi fg="black-bold">.:</ansi><ansi fg="20">Attributes</ansi> ')
    .SetMinWidth(40)
    .SetColumns(2)
    .SetColumnGap(2)
    .Add('<ansi fg="yellow">Strength: </ansi>', '<ansi fg="yellow">Str:</ansi>', String(user.GetStat("strength")))
    .Add('<ansi fg="yellow">Vitality: </ansi>', '<ansi fg="yellow">Vit:</ansi>', String(user.GetStat("vitality")))
    .Add('<ansi fg="yellow">Speed:    </ansi>', '<ansi fg="yellow">Spd:</ansi>', String(user.GetStat("speed")))
    .Add('<ansi fg="yellow">Mysticism:</ansi>', '<ansi fg="yellow">Mys:</ansi>', String(user.GetStat("mysticism")))
    .Add('<ansi fg="yellow">Smarts:   </ansi>', '<ansi fg="yellow">Smt:</ansi>', String(user.GetStat("smarts")))
    .Add('<ansi fg="yellow">Percept:  </ansi>', '<ansi fg="yellow">Per:</ansi>', String(user.GetStat("perception")));

layout.Panel("gold")
    .SetTitle(' <ansi fg="black-bold">.:</ansi><ansi fg="20">Wealth</ansi> ')
    .SetMinWidth(18)
    .Add('<ansi fg="yellow">Gold:</ansi>', '<ansi fg="yellow">G:</ansi>', String(user.GetGold()))
    .Add('<ansi fg="yellow">Bank:</ansi>', '<ansi fg="yellow">B:</ansi>', String(user.GetBankGold()));

layout.Panel("training")
    .SetTitle(' <ansi fg="black-bold">.:</ansi><ansi fg="20">Training</ansi> ')
    .SetMinWidth(18)
    .Add('<ansi fg="yellow">Train Pts:</ansi>', '<ansi fg="yellow">Trn:</ansi>', String(user.GetTrainingPoints()))
    .Add('<ansi fg="yellow">Stat Pts: </ansi>', '<ansi fg="yellow">Sta:</ansi>', String(user.GetStatPoints()));

user.SendText(layout.Render());
```

Output:
```
 ┌─ .:Info ────────────────┐ ┌─ .:Attributes ───────────────────────────┐
 │ Name:   Aldric           │ │ Strength:  42(+3)  Vitality:  38(+0)     │
   Race:   Human                Speed:     55(+5)  Mysticism: 20(+0)
   Level:  12                    Smarts:    30(+1)  Percept:   44(+2)
   Health: 142/200           │ └──────────────────────────────────────────┘
 │ Mana:   80/100            │ ┌─ .:Wealth ──────────┐ ┌─ .:Training ────┐
 └──────────────────────────┘ │ Gold: 1,234          │ │ Train Pts: 5    │
                                 Bank: 10,000             Stat Pts:  2
                               │                     │ │                 │
                               └─────────────────────┘ └─────────────────┘
```

---

### Loading a layout from a YAML file

When a layout is defined in a YAML file under
`_datafiles/panel-layouts/<name>.yaml`, scripts can load it and populate the
panels without knowing anything about the structure.

```javascript
// Load the layout skeleton defined in the YAML file
var layout = PanelLayoutLoad("character/status");

// Populate panels by their id
layout.Panel("info")
    .Add('<ansi fg="yellow">Name:  </ansi>', '<ansi fg="yellow">N:</ansi>', user.GetName())
    .Add('<ansi fg="yellow">Level: </ansi>', '<ansi fg="yellow">Lvl:</ansi>', String(user.GetLevel()));

layout.Panel("attributes")
    .Add('<ansi fg="yellow">Strength:</ansi>', '<ansi fg="yellow">Str:</ansi>', String(user.GetStat("strength")))
    .Add('<ansi fg="yellow">Vitality:</ansi>', '<ansi fg="yellow">Vit:</ansi>', String(user.GetStat("vitality")));

user.SendText(layout.Render());
```

This approach is useful when the visual layout (which panels exist, how they
are arranged, minimum widths, border style) should be configurable by a server
operator without touching script code.

---

### Two-column stats panel

When `SetColumns(2)` is set on a panel, rows are paired left-to-right. The
first `.Add()` call and the second share a line, the third and fourth share the
next, and so on.

```javascript
var layout = PanelLayoutNew({ border: "full", margin: 1 });

var slot = layout.AddSlot();
slot.AddRow(["stats"]);

layout.Panel("stats")
    .SetTitle(' <ansi fg="20">Attributes</ansi> ')
    .SetMinWidth(44)
    .SetColumns(2)
    .SetColumnGap(2)
    .Add('<ansi fg="yellow">Strength: </ansi>', '<ansi fg="yellow">Str:</ansi>', "42 (+3)")
    .Add('<ansi fg="yellow">Vitality: </ansi>', '<ansi fg="yellow">Vit:</ansi>', "38 (+0)")  // paired with Strength
    .Add('<ansi fg="yellow">Speed:    </ansi>', '<ansi fg="yellow">Spd:</ansi>', "55 (+5)")
    .Add('<ansi fg="yellow">Mysticism:</ansi>', '<ansi fg="yellow">Mys:</ansi>', "20 (+0)")  // paired with Speed
    .Add('<ansi fg="yellow">Smarts:   </ansi>', '<ansi fg="yellow">Smt:</ansi>', "30 (+1)")
    .Add('<ansi fg="yellow">Percept:  </ansi>', '<ansi fg="yellow">Per:</ansi>', "44 (+2)"); // paired with Smarts

user.SendText(layout.Render());
```

Output:
```
 ┌─ .:Attributes ────────────────────────────────┐
 │ Strength:  42 (+3)  Vitality:  38 (+0)        │
 │ Speed:     55 (+5)  Mysticism: 20 (+0)        │
 │ Smarts:    30 (+1)  Percept:   44 (+2)        │
 └───────────────────────────────────────────────┘
```

---

### Progress bar in a panel

Values are plain strings, so any pre-built string — including a progress bar —
works as a value. The renderer only measures the visual width.

```javascript
function makeBar(current, max, width) {
    var filled = Math.round((current / max) * width);
    var empty  = width - filled;
    return '<ansi fg="green">' + Array(filled + 1).join("█") + '</ansi>' +
           '<ansi fg="black-bold">' + Array(empty + 1).join("░") + '</ansi>' +
           " " + current + "/" + max;
}

var layout = PanelLayoutNew({ border: "full", margin: 1 });
var slot = layout.AddSlot();
slot.AddRow(["vitals"]);

layout.Panel("vitals")
    .SetTitle(' <ansi fg="20">Vitals</ansi> ')
    .SetMinWidth(36)
    .Add('<ansi fg="yellow">Health:</ansi>', '<ansi fg="yellow">HP:</ansi>', makeBar(142, 200, 12))
    .Add('<ansi fg="yellow">Mana:  </ansi>', '<ansi fg="yellow">MP:</ansi>', makeBar(80, 100, 12));

user.SendText(layout.Render());
```

Output:
```
 ┌─ .:Vitals ──────────────────────────┐
 │ Health: ████████░░░░ 142/200        │
 │ Mana:   █████████░░░ 80/100         │
 └─────────────────────────────────────┘
```

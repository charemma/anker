# 0009: Template Engine Selection

**Date:** 2026-01-28
**Status:** Proposed
**Participants:** Charalambos Emmanouilidis, Claude.ai

## Problem

Users need different report formats (client timesheets, standup notes, weekly summaries). How should we implement customizable report templates?

Requirements:
- Flexible enough for complex formatting
- Users control what data appears in reports
- Not just text substitution, need data selection/filtering
- Ideally language-agnostic for community sharing
- Should feel natural for developers

## Options Considered

### Option A: Go templates (text/template)

```go
{{range .Repos}}
## {{.Name}}
{{range .Entries}}
- {{.Time.Format "15:04"}} {{.Content}}
{{end}}
{{end}}
```

**Good:**
- Zero dependencies (stdlib)
- Type-safe with Go structs
- Good performance
- Well documented
- Used by Hugo, Helm, kubectl

**Bad:**
- Syntax unfamiliar to non-Go developers
- Limited feature set (no filter chains)
- Go-specific (lock-in)
- Less powerful for complex logic
- Data structure is fixed (can't query dynamically)

### Option B: Jinja2-style (pongo2)

```jinja
{% for repo in repos %}
## {{ repo.name }}
{% for entry in repo.entries %}
- {{ entry.time|date:"15:04" }} {{ entry.content }}
{% endfor %}
{% endfor %}
```

**Good:**
- Very familiar (Python/web community)
- Powerful: filters, tests, inheritance
- User-friendly syntax
- Many online examples

**Bad:**
- External dependency (pongo2)
- Not "real" Jinja (Go port, slight differences)
- Still Go-bound (not truly language-agnostic)
- Data structure still fixed by our code

### Option C: Mustache/Handlebars

```mustache
{{#repos}}
## {{name}}
{{#entries}}
- {{time}} {{content}}
{{/entries}}
{{/repos}}
```

**Good:**
- Truly language-agnostic (spec exists)
- Simple, logic-less
- Many implementations (Go, Python, JS, Ruby)
- Used by GitHub, GitLab

**Bad:**
- Too limited for complex reports
- No real conditionals or loops
- Can't do date formatting
- All logic must be in Go code
- Can't query/filter data in template

### Option D: Lua (gopher-lua)

```lua
-- template.lua
function generate(data)
  local output = {}

  for _, repo in ipairs(data.repos) do
    table.insert(output, "## " .. repo.name)

    -- User controls what to show
    for _, entry in ipairs(repo.entries) do
      if entry.type == "commit" then  -- User decides filter
        table.insert(output, string.format("- %s %s",
          format_time(entry.time), entry.content))
      end
    end
  end

  return table.concat(output, "\n")
end
```

**Good:**
- Full programming language (maximum flexibility)
- Truly language-agnostic (many embeddings)
- User can query/filter data dynamically
- Can call functions we provide (format_time, filter_by_date, etc.)
- Used by Redis, nginx, Neovim, WoW
- Small, embeddable, fast
- Users control data selection, not just formatting

**Bad:**
- Learning curve (users need to know Lua)
- Potentially overkill for simple templates
- Security considerations (arbitrary code)
- More complex than declarative templates

### Option E: Starlark (Bazel's config language)

```python
# template.star
def generate(data):
    output = []
    for repo in data.repos:
        output.append("## " + repo.name)
        for entry in repo.entries:
            output.append("- %s %s" % (entry.time, entry.content))
    return "\n".join(output)
```

**Good:**
- Python-like syntax (familiar)
- Safe by design (no I/O, deterministic)
- Used by Bazel, Google
- Language-agnostic

**Bad:**
- Less known than Lua or Jinja
- Still learning curve
- External dependency
- Might be overkill

### Option F: Hybrid (Multiple engines)

Support both simple and powerful:
```bash
# Simple: Go templates for basic cases
anker report --template weekly.tmpl

# Powerful: Lua for complex reports
anker report --template client-report.lua
```

**Good:**
- Best of both worlds
- Users choose complexity level
- Can start simple, upgrade to Lua
- Built-in templates use Go (simple)
- Power users can use Lua

**Bad:**
- Two systems to maintain
- More complexity in codebase
- Documentation for both

## Decision

**Use Lua for all text templates. JSON as built-in Go serialization.**

### Format Types

**1. JSON (Built-in, Go)**
```bash
anker report april --format json
```

- Implemented in Go using `json.Marshal()`
- Why Go, not Lua:
  - Type-safe serialization
  - Guaranteed valid JSON
  - No floating point/encoding issues
  - Data export, not text formatting
  - Standard library handles edge cases

**2. Text Templates (Lua)**
```bash
anker report april                           # Uses embedded markdown.lua
anker report april --format html             # Uses embedded html.lua
anker report april --template custom.lua     # User's template
```

- All text formatting uses Lua
- Even built-in templates (markdown, html) are Lua
- Why Lua for built-ins too:
  - **Transparency**: User can see how it works
  - **Examples**: Built-ins are learning material
  - **Consistency**: One system, not "this is hardcoded, that is template"
  - **Customizable**: User copies built-in, tweaks it
  - **Shows power**: Built-ins demonstrate filtering, grouping

### Built-in Templates (Embedded Lua)

Shipped in binary via embedded filesystem:

```
anker (binary)
  ├─ templates/
  │   ├─ markdown.lua      # Default, rich formatting
  │   ├─ compact.lua       # Short, for quick review
  │   ├─ html.lua          # For email/browser
  │   └─ timesheet.lua     # For freelancers
  └─ lib/
      └─ helpers.lua       # Shared functions
```

### Example: markdown.lua (Built-in)

```lua
-- Built-in template users can learn from
function generate(data)
  local output = {}

  table.insert(output, "# Work Report")
  table.insert(output, string.format("Period: %s - %s\n",
    anker.format_date(data.from), anker.format_date(data.to)))

  for repo, entries in pairs(anker.group_by_repo(data)) do
    table.insert(output, "\n## " .. repo)

    for _, entry in ipairs(entries) do
      -- Example: Filter out WIP commits
      if not entry.content:match("^WIP") then
        table.insert(output, string.format("- %s", entry.content))
      end
    end
  end

  return table.concat(output, "\n")
end
```

### User Experience

```bash
# View built-in template source
anker template show markdown

# Copy built-in as starting point
anker template copy markdown my-custom.lua

# Edit to customize
vim ~/.anker/templates/my-custom.lua

# Use custom template
anker report april --template my-custom.lua
```

### Why This Approach

**One system:**
- No confusion between "hardcoded formats" and "templates"
- Users learn Lua once, can customize everything
- Built-ins are documentation by example

**Transparency:**
- User sees exact code that generates markdown
- Can copy and modify any built-in
- No "magic" - everything is inspectable

**Progressive complexity:**
- Simple: Use built-ins as-is
- Medium: Copy built-in, tweak filters
- Advanced: Write from scratch with full control

**Learning path:**
```lua
-- Simple: Use helpers
return anker.format_markdown(data)

-- Medium: Filter then format
local filtered = anker.filter(data.entries, function(e)
  return not e.content:match("^WIP")
end)
return anker.format_markdown({entries = filtered})

-- Advanced: Full control
local output = {}
-- ... custom logic ...
return table.concat(output, "\n")
```

### Implementation

**Lua Runtime:**
- Use `gopher-lua` (pure Go, no CGO)
- Bundle in binary (~2MB overhead)
- Cross-platform, static compilation works

**Helper Functions (anker.* API):**
```lua
-- Filtering
anker.filter(entries, predicate)
anker.filter_by_author(entries, email)
anker.filter_work_hours(entries)

-- Grouping
anker.group_by_repo(data)
anker.group_by_date(data)
anker.group_by_source(data)

-- Formatting
anker.format_date(timestamp, format)
anker.format_time(timestamp, format)
anker.format_markdown(data)  -- Quick helper

-- Data access
entry.content, entry.time, entry.source, entry.location
```

### Advantages

**For users:**
- ✅ Built-in templates work great out-of-box
- ✅ Can see and learn from built-in code
- ✅ Easy to customize (copy, modify)
- ✅ Full power when needed (filter WIP, custom grouping)

**For development:**
- ✅ One template system, not two
- ✅ Less code to maintain
- ✅ Built-ins are just files (easy to improve)
- ✅ Community can share templates (same language)

**Example use cases:**
```lua
-- Filter out WIP commits
if not entry.content:match("^WIP") then
  -- render it
end

-- Only show work hours
if entry.time.hour >= 9 and entry.time.hour <= 17 then
  -- render it
end

-- Group by week, then repo
local by_week = anker.group_by_week(data)
for week, entries in pairs(by_week) do
  -- format week section
end
```

## Open Questions

1. Is Lua too much for typical users? (Built-in templates hide this)
2. Should we sandbox Lua (no file I/O, network)? (Yes, definitely)
3. Provide Lua standard library or minimal? (Minimal + our helpers)
4. Template marketplace/sharing? (Later, community-driven)

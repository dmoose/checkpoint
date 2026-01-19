# ripgrep (rg)

Fast regex search tool for codebases.

## Purpose

Search for patterns across files quickly. Much faster than grep for large codebases.

## Common Commands

```bash
rg "pattern"                    # Basic search
rg "pattern" --type go          # Filter by filetype
rg "pattern" -g "*.go"          # Glob filter
rg "pattern" -A 3 -B 3          # With context lines
rg "pattern" -l                 # Just filenames
rg "pattern" -c                 # Count matches
rg "pattern" -i                 # Case insensitive
rg "pattern" -w                 # Word boundaries
rg "func \w+\(" --type go       # Regex: find functions
rg "TODO|FIXME"                 # Multiple patterns
```

## Tips

- Use --type instead of glob for common languages (go, py, js, etc.)
- Use -w for whole word matching
- Use -F for literal strings (no regex)
- Use --hidden to include dotfiles
- Use -g '!vendor' to exclude directories

## File Types

```bash
rg --type-list                  # Show all known types
rg "pattern" -t go -t rust      # Multiple types
```

# Dead Renderer Code in Converter

## Goal

Remove dead methods on `ConfluenceRenderer` whose output never reaches the live `MarkdownToStorage` pipeline because goldmark's GFM extension registers the same node kinds at a higher precedence. The result is a converter package that matches what actually runs in production, with no dead branches that suggest Confluence-specific rendering is in effect when it is not.

## Scope

In scope:

- Delete dead methods on `ConfluenceRenderer` whose live output is produced by goldmark's GFM renderers.
- Remove the matching `RegisterFuncs` entries.
- Adjust comments and section headings in `confluence_renderer.go` once the surviving code no longer matches the original groupings.

Out of scope:

- Changing renderer priority in `markdown.go` to make `ConfluenceRenderer` win over GFM.
- Any change to the HTML shape that `MarkdownToStorage` produces for tables, task lists, or strikethrough on existing inputs.
- Round-trip re-validation against a real Confluence instance.

## Current State

File: `internal/converter/confluence_renderer.go`. The renderer registers handlers for several node kinds that goldmark's `extension.GFM` also handles. GFM's renderers run at default priority; `ConfluenceRenderer` is registered at priority 1000. In goldmark, lower priority numbers run earlier and the last write wins, so GFM's output is what reaches Confluence whenever both register for the same kind.

Candidate dead methods (registered for kinds also covered by GFM):

| Function | Registered for |
|----------|----------------|
| renderTable | extast.KindTable |
| renderTableHeader | extast.KindTableHeader |
| renderTableRow | extast.KindTableRow |
| renderTableCell | extast.KindTableCell |
| renderTaskCheckBox | extast.KindTaskCheckBox |
| renderStrikethrough | extast.KindStrikethrough |

A repository-wide search confirms each of the six is referenced only by its definition and its single `RegisterFuncs` entry. There are no direct callers anywhere in the tree, so removal is a mechanical deletion with no caller-side consequences.

All six report 0.0% coverage today:

```bash
go test -coverprofile=/tmp/cov.out ./internal/converter
go tool cover -func=/tmp/cov.out | grep -E "renderTable|renderTaskCheckBox|renderStrikethrough"
```

`renderHTMLBlock` and `renderRawHTML` are not dead in the same sense: there is no GFM override for those kinds, so the project renderer is what runs. They are intentionally kept.

Build setup, `internal/converter/markdown.go:14`:

```go
goldmark.New(
    goldmark.WithExtensions(extension.GFM),
    goldmark.WithRenderer(
        renderer.NewRenderer(
            renderer.WithNodeRenderers(
                util.Prioritized(NewConfluenceRenderer(), 1000),
            ),
        ),
    ),
)
```

Observable divergence between the project renderer's intent and the live output:

| Concern | Project renderer would emit | Live pipeline emits |
|---------|----------------------------|---------------------|
| Table wrapper | `<table><tbody>` | `<table>\n<thead>...<tbody>` |
| Cell alignment | `align="right"` | `style="text-align:right"` |
| Task body | plain task body content | `<input disabled="" type="checkbox">` injected |
| Strikethrough | `<del>...</del>` | `<del>...</del>` (coincidentally identical) |

Confluence storage format accepts both forms, so the divergence has not caused a functional problem. The current `internal/converter/markdown_test.go` exercises tables, task lists, and strikethrough through `MarkdownToStorage` using substring assertions that match the GFM output. None of those tests rely on the dead methods.

## References

- Goldmark renderer priority: https://github.com/yuin/goldmark/blob/master/renderer/renderer.go
- Confluence storage format spec: https://confluence.atlassian.com/doc/confluence-storage-format-790796544.html
- GitHub issue 6 (inline code HTML escaping) — the work that surfaced this finding.

## Requirements

1. The six methods listed in Current State are removed from `internal/converter/confluence_renderer.go`.
2. The six matching `RegisterFuncs` entries are removed.
3. Comments and grouping headers in `confluence_renderer.go` are updated so they accurately describe what remains.
4. All existing tests in `internal/converter` continue to pass without modification to assertions.
5. Coverage output for the converter package no longer lists any of the deleted functions.

## Implementation Plan

1. Remove the `RegisterFuncs` entries for the six kinds in `confluence_renderer.go`.
2. Delete the six method definitions.
3. Update the section comments in `RegisterFuncs` to reflect the kinds that remain handled.
4. Run `go test ./internal/converter` and confirm all tests pass.
5. Run the coverage command in Current State and confirm none of the deleted functions appear in the output.
6. Run `gofmt -w .` and `golangci-lint run` to confirm no new issues.

## Constraints

- Go 1.25.4 toolchain.
- `MarkdownToStorage` output for any input that currently passes through the test suite must not change. The pre-existing assertions in `markdown_test.go` must remain valid after the deletion.
- Do not change the renderer priority in `markdown.go`. The chosen approach is deletion, not promotion.

## Issues Discovered

1. Option A vs Option B not formally locked (decision) — Resolved: Option B locked.

   The project assumes Option B (delete the dead methods). The rejected alternative, Option A, is to promote `ConfluenceRenderer` to priority 0 so it wins over GFM. Option A would change live HTML shape for tables, task lists, and the GFM strikethrough path, and would require re-validating round-trip behaviour against a real Confluence instance. Without an explicit lock the implementer could read the analysis and pick either path.
   Resolution: Option B. The live pipeline already produces working output, the divergence is cosmetic, and Option A invites regression risk against a Confluence instance that cannot be validated locally. Do not change renderer priority in `markdown.go`.

2. `renderStrikethrough` included by symmetry with table renderers (design) — Resolved: in scope for deletion.

   Earlier framing of the problem treated `renderStrikethrough` as a special case because GFM's strikethrough output happens to match the project renderer's output. By the priority-1000 rule that justifies deleting the table methods, `renderStrikethrough` is equally unreachable and should be deleted alongside them. The requirements above already include it; confirming this in writing prevents a future implementer from second-guessing the inclusion.
   Resolution: delete `renderStrikethrough` and its registration. Verified that the only references in the codebase are the definition and the single registration; the function reports 0% coverage today; the existing strikethrough test in `markdown_test.go` passes against GFM's output, not this method's. Deletion has no functional or test impact.

## Acceptance Criteria

- `internal/converter/confluence_renderer.go` no longer defines `renderTable`, `renderTableHeader`, `renderTableRow`, `renderTableCell`, `renderTaskCheckBox`, or `renderStrikethrough`, and no longer registers any of those kinds.
- `go test ./internal/converter` passes.
- `go tool cover -func=/tmp/cov.out` for the converter package lists none of the six deleted functions.
- Round-trip behaviour for `MarkdownToStorage` is unchanged: all pre-existing assertions in `internal/converter/markdown_test.go` still pass without edits.

# Tasks — Project Pages

_Feature slug: `project-pages`_ · Vertical slices, in build order.

- [x] **1. Data model** — `internal/project/project.go`: `Project` + `Metric`
      structs, `Load(dir)` (reads `projects/*.yaml`, sorts by `order`), `BySlug`.
      Test in `project_test.go`.
- [x] **2. Project content** — one `projects/<slug>.yaml` per project (5 files),
      seeded with verifiable copy + honest `.results` metrics.
- [x] **3. Templates** — `projects_template.html` (index) and
      `project_template.html` (detail), reusing the site chrome.
- [x] **4. Styles** — append project-detail + index styles to `static/css/styles.css`
      (stat tiles, detail layout, back link), all via tokens.
- [x] **5. Wire-up** — `main.go`: parse new templates, `readProjects()`, routes
      `GET /projects` and `GET /projects/{slug}`, pass projects to home/PDF/MD.
- [x] **6. Single source** — remove `projects:` from `cv.yaml` and the `Project`
      type + `Projects` field from `internal/cv`; update home Work section to read
      the loaded projects and link to detail pages; add `Projects` nav link.
- [x] **7. Verify** — `go build`, `go test ./...`, run server, smoke-test
      `/`, `/projects`, each `/projects/{slug}`, and `/projects/nope` → 404.

## Definition of done

- Every project has a working detail page reachable from the home list and the index.
- Achievements are truthful and carry a dated source note.
- No fabricated stack/timeline/URL — unknowns render as hidden sections.
- PDF and Markdown CV exports still list projects (from the new source).
